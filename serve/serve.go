
package serve

import (
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"slices"
	"strings"
	commandparse "tboard/command_parse"
	ed "tboard/ed"
	_ "github.com/mattn/go-sqlite3"
)

const welcome_message =
`Welcome to the tboard message board
    -Created by Geogory Tibisov
boards:

`

type Server struct{
	DBconn *sql.DB
	Listener net.Listener
	BoardList []string
}

type UserInfo struct{
	Conn net.Conn
	UID  [sha512.Size384]byte
	buf  []byte
	Current_PID int
	Current_Board string
}

func CreateServer(/*ConfigData []byte*/)(srv *Server){
	var err error
	srv = new(Server)
	srv.Listener,err = net.Listen("tcp","localhost:6969")
	if(err != nil){
		log.Panic(err)
	}
	srv.DBconn,err = sql.Open("sqlite3","./tboard.db")
	if(err != nil){
		log.Panic(err)
	}
	admin_table_statement := `CREATE TABLE IF NOT EXISTS admins(
    username TEXT PRIMARY KEY,
    password BLOB NOT NULL);`
	srv.DBconn.Exec(admin_table_statement)
	general_message_board_stmt := `CREATE TABLE IF NOT EXISTS general_board(
     post_id INTEGER PRIMARY KEY AUTOINCREMENT,
     poster_id INTEGER,
     posted_date DATE DEFAULT CURRENT_DATE,
     posted_time TEXT DEFAULT CURRENT_TIME,
     body TEXT);`
	srv.DBconn.Exec(general_message_board_stmt)
	desc_table_statement := `CREATE TABLE IF NOT EXISTS desc(
    board TEXT PRIMARY KEY,
    text TEXT NOT NULL);`
	srv.DBconn.Exec(desc_table_statement)
	return
}

func (srv *Server)GetBoards()string{
	builder := strings.Builder{}
	builder.WriteString(welcome_message)
	table_names_rows,err := srv.DBconn.Query("SELECT name FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite_%';")
	if(err != nil){
		log.Panic(err)
	}
	for table_names_rows.Next(){
		table_name := ""
		table_names_rows.Scan(&table_name)
		if !slices.Contains(srv.BoardList,table_name){
			srv.BoardList = append(srv.BoardList,table_name)
		}
		if(strings.Contains(table_name,"board")){
			desc_str := ""
			sql_stmt := `SELECT text FROM desc WHERE board = ?`
			err := srv.DBconn.QueryRow(sql_stmt,table_name).Scan(&desc_str)
			if err != nil{
				log.Print(err)
				fmt.Fprintf(&builder,"%s -- No Description given\r\n",table_name)
			}else{
				builder.WriteString(fmt.Sprintf("%s -- %s\n",table_name,desc_str))
			}
		}
	}
	return builder.String()
}

func (srv *Server)ViewPID(PID int,user *UserInfo)error{
	// `CREATE TABLE IF NOT EXISTS general_board(
    //  post_id INTEGER PRIMARY KEY AUTOINCREMENT,
    //  poster_id BLOB,
    //  posted_date DATE DEFAULT CURRENT_DATE,
    //  body TEXT)
	
	builder := strings.Builder{}
	poster_id := make([]byte,sha512.Size384)
	date_str := ""
	time_str := ""
	body := ""
	builder.WriteString("-----------------------------------------------------------------------------\n\r")
	sql_stmt := fmt.Sprintf("SELECT poster_id,date(posted_date),posted_time,body FROM %s WHERE post_id = ?",user.Current_Board)
	err := srv.DBconn.QueryRow(sql_stmt,PID).Scan(&poster_id,&date_str,&time_str,&body)
	if(err != nil){
		fmt.Fprintf(&builder,"No post with post_id %d  in board '%s\r\n",PID,user.Current_Board)
		log.Print(err)
		builder.WriteString("-----------------------------------------------------------------------------\n\r")
		fmt.Fprint(user.Conn,builder.String())
		err := errors.New(builder.String())
		return err
	}
	fmt.Fprintf(&builder,"post %d posted by %x at %s:%s\n %s\n",PID,poster_id,date_str,time_str,body)
	builder.WriteString("-----------------------------------------------------------------------------\n\r")
	fmt.Fprint(user.Conn,builder.String())
	return nil
}
func (srv *Server)Post(user *UserInfo){
	//TODO -- MAKE ed LIKE EDITOR
	post,err := ed.Post_Editor(user.Conn)
	if(len(post) == 0){
		return
	}
	sql_stmt := fmt.Sprintf("INSERT INTO %s (poster_id,body) VALUES (?, ?)",user.Current_Board)
	_,err = srv.DBconn.Exec(sql_stmt,user.UID[:],post)
	if err != nil{
		log.Print(err)
	}
}
func (srv *Server)BoardLoop(user *UserInfo,board string){
	if(!slices.Contains(srv.BoardList,board)){
		fmt.Fprintf(user.Conn,"The Board '%s' does not exist\r\n",board)
		return
	}
	user.Current_Board = board
	err := srv.goto_date("today",user)
	if(err != nil){
		fmt.Fprint(user.Conn,err.Error())
	}else{
		srv.ViewPID(user.Current_PID,user)
	}
	for{
		n,err := user.Conn.Read(user.buf)
		if err != nil{
			log.Print(err)
			return
		}
		user_in := string(user.buf[:n])
		user_in = strings.Trim(user_in,"\r\n")
		cmd,err :=commandparse.Parse_Command(strings.ToLower(user_in),commandparse.CommandTypeBoard)
		if(err != nil){
			user.Conn.Write([]byte(err.Error()))
		}else{
			if(cmd.Command_Type == commandparse.Command_Exit){
				return
			}
			err = srv.EvalCommand(cmd,user)
			if err != nil{
				log.Printf("User %x attempted command that led to error '%s'\n",user.UID,err.Error())
				user.Conn.Write([]byte(err.Error()))
			}
		}
		
	}
}

func (srv *Server)EvalCommand(cmd commandparse.Parsed_Command,user *UserInfo)error{
	switch(cmd.Command_Type){
		case commandparse.Command_View:
		{
			err:= srv.handle_view(cmd,user)
			if(err != nil){
				return err
			}
		}
		
		case commandparse.Command_Next:
		{
			if(cmd.SubCom_Type == commandparse.Next_Skip){
				skip_val := cmd.Arguments[0].Get_Int()
				err := srv.goto_PID(user.Current_PID + skip_val,user)
				if err != nil{
					return err
				}
			}else{
				err := srv.goto_PID(user.Current_PID + 1,user)
				if err != nil{
					return err
				}
			}
		}
		case commandparse.Command_Goto:
		err := srv.handle_goto(cmd,user)
		if(err != nil){
			return  err
		}
		case commandparse.Command_Post:
		{
			srv.Post(user)
		}
		default:
		{
			return errors.New("Unimplemented command")
		}
	}
	return nil
}

func(srv *Server)handle_view(cmd commandparse.Parsed_Command,user *UserInfo)error{
	switch(cmd.SubCom_Type){
		case commandparse.View_Default:
		{
			err := srv.ViewPID(user.Current_PID,user)
			if(err != nil){
				return err
			}
		}
		case commandparse.View_Range:
		{
			start := cmd.Arguments[0].Get_Int()
			end := cmd.Arguments[1].Get_Int()
			for i := start;i <= end;i++{
				err := srv.ViewPID(i,user)
				if(err != nil){
					return err
				}
			}
		}
		case commandparse.View_PostID:
		{
			post_id := cmd.Arguments[0].Get_Int()
			srv.ViewPID(post_id,user)
		}
		case commandparse.View_Date:
		{
			date := cmd.Arguments[0].Get_String()
			prev_post_id := user.Current_PID
			err := srv.goto_date(date,user)
			if(err != nil){
				return err
			}
			srv.ViewPID(user.Current_PID,user)
			user.Current_PID = prev_post_id
		}
		default:
		{
			fmt.Fprintf(user.Conn,"Unimplemented view subcommand\r\n")
		}
	}
	return nil
}

func (srv *Server)handle_goto(cmd commandparse.Parsed_Command,user *UserInfo)error{
	switch(cmd.SubCom_Type){
		case commandparse.Goto_PostID:{
			pid := cmd.Arguments[0].Get_Int()
			err := srv.goto_PID(pid,user)
			if(err != nil){
				return err
			}
		}
		case commandparse.Goto_Date:
		{
			date_str := cmd.Arguments[0].Get_String()
			err := srv.goto_date(date_str,user)
			if(err != nil){
				return err
			}
		}
		
	}
	return nil
}
func (srv *Server)goto_PID(PID int,user *UserInfo) error{
	//check if post_id exists
	sql_stmt := fmt.Sprintf("SELECT post_id FROM %s WHERE post_id = ?",user.Current_Board)
	post_id := 0
	err :=  srv.DBconn.QueryRow(sql_stmt,PID).Scan(&post_id)
	if(err != nil){
		
		return err
	}
	user.Current_PID = post_id
	return nil
}
func (srv *Server)goto_date(date string,user *UserInfo)error{
	if(date == "today"){
		sql_stmt := fmt.Sprintf("SELECT post_id FROM %s WHERE posted_date = date('now');",user.Current_Board)
		err := srv.DBconn.QueryRow(sql_stmt).Scan(&user.Current_PID)
		if(err != nil){
			err = fmt.Errorf("There are no posts in the Board '%s' Today! Create one!\r\n",user.Current_Board)
			return err
		}
		return nil
	}
	sql_stmt := fmt.Sprintf("SELECT post_id FROM %s WHERE date(posted_date) = ?",user.Current_Board)
	post_id := 0
	err := srv.DBconn.QueryRow(sql_stmt,date).Scan(&post_id)
	if(err != nil){
		log.Print(err)
		return err
	}
	user.Current_PID = post_id
	return nil
}
func (srv *Server)UserLoop(info *UserInfo){
	welcome_str := srv.GetBoards()
	info.buf = make([]byte,1024)
	for{
		info.Conn.Write([]byte(welcome_str))
		n,err := info.Conn.Read(info.buf)
		if(err != nil){
			log.Print(err)
			info.Conn.Close()
			return
		}
		input := info.buf[:n]
		usr_str := string(input)
		usr_str = strings.Trim(usr_str,"\n\r")
		switch(strings.ToLower(usr_str)){
			case "admin":
			{
				panic("TODO -- CREATE ADMIN CONSOLE")
			}
			
			case "quit","q","exit","bye":
			{
				info.Conn.Close()
				return
			}
			default:
			{
				srv.BoardLoop(info,strings.ToLower(usr_str))
			}
		}
		
	}
}


func (srv *Server)ServerLoop(){
	for{
		conn,err := srv.Listener.Accept()
		if(err != nil){
			log.Print(err)
			conn.Close()
		}else{
			userinfo := new(UserInfo)
			userinfo.Conn = conn
			userinfo.UID = sha512.Sum384([]byte(strings.Split(conn.RemoteAddr().String(),":")[0] + os.Getenv("TBOARD_SALT")))
			go srv.UserLoop(userinfo)
		}
	}
}
