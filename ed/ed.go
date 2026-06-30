package ed

import (
	"errors"
	"net"
	"slices"
	"strings"
	commandparse "tboard/command_parse"

	"fmt"
)

const editor_help string = `This editor is super simple and should be easy to use!
----------------------------------------------------------------------------------
append is used to add text to the buffer,it has 3 modes
    append | this will add text to the bottom of the buffer.
    append 1(or any valid line number) | this will insert text starting at that point.
----------------------------------------------------------------------------------
the view command is also super simple there are 4 modes to the command.
    view | the default operation is to just view the current line!
    view 1(or any valid line number)  | will view the given line number.
    view range 1 3 (any 2 valid line numbers are supplied to this command) | this will view a range of the buffer!
    view * | this will view the WHOLE buffer.
----------------------------------------------------------------------------------
the swap command is simple and only has one mode of operation
    swap 1(or any valid line number) | you can easily swap certian lines with new lines of text!
----------------------------------------------------------------------------------
the delete command is also just as simple with one mode of operation
    swap 1(or any valid line number) | you can easily swap certian lines with new lines of text!
----------------------------------------------------------------------------------
`


type editor struct{
	conn net.Conn
	user_buffer []byte
	line_buf []string
	current_line int
}


func Post_Editor(user_conn net.Conn)(string,error){
	ed := editor{}
	ed.conn = user_conn
	ed.line_buf = make([]string,0,5)
	ed.user_buffer = make([]byte,512)
	user_conn.Write([]byte("This is a simple  editor please type 'help' for assitance\r\n"))
	return ed.ed_loop()
}

func(ed *editor)default_append()error{
	fmt.Fprint(ed.conn,"append mode - please type a single '.' to exit\r\n")
	for{
		n,err := ed.conn.Read(ed.user_buffer)
		if(err != nil){
			return err
		}
		user_str := string(ed.user_buffer[:n])
		if(user_str[0] == '.' && strings.Trim(user_str,"\r\n") == "."){
			break
		}
		ed.line_buf = append(ed.line_buf,user_str)
		ed.current_line++
	}
	return nil
}
func(ed *editor)append_line(line_num int)error{
	fmt.Fprint(ed.conn,"append mode - please type a single '.' to exit\r\n")
	if(line_num < 1 ||line_num > len(ed.line_buf)){
		return fmt.Errorf("Invaild line number %d\r\n",line_num)
	}
	line_num = line_num - 1
	for{
		n,err := ed.conn.Read(ed.user_buffer)
		if(err != nil){
			return err
		}
		user_str := string(ed.user_buffer[:n])
		if(user_str[0] == '.' && strings.Trim(user_str,"\r\n") == "."){
			break
		}
		ed.line_buf = slices.Insert(ed.line_buf,line_num,user_str)
		ed.current_line = line_num
		line_num++
	}
	return nil
}
func(ed *editor)handle_append(cmd commandparse.Parsed_Command)error{
	switch(cmd.SubCom_Type){
		case commandparse.Editor_Append_Default:
		{
			err := ed.default_append()
			if(err != nil){
				return err
			}
		}
		case commandparse.Editor_Append_Line:
		{
			line_num := cmd.Arguments[0].Get_Int()
			err := ed.append_line(line_num)
			if(err != nil){
				return err
			}
		}
		default:
		panic("INVAILD APPEND SUBCOMAND -- PANIC")
	}
	return nil
}
func(ed *editor)view_line(line_num int)error{
	if(line_num < 0 || line_num > len(ed.line_buf) - 1){
		return  fmt.Errorf("Invaild line number %d\r\n",line_num + 1)
	}
	fmt.Fprint(ed.conn,ed.line_buf[line_num])
	return nil
}
func(ed *editor)handle_view(cmd commandparse.Parsed_Command)error{
	switch (cmd.SubCom_Type){
		case commandparse.Editor_View_Default:
		{
			err := ed.view_line(ed.current_line - 1)
			if err != nil{
				return err
			}
		}
		case commandparse.Editor_View_Line:
		{
			line_num := cmd.Arguments[0].Get_Int()
			err := ed.view_line(line_num - 1)
			if(err != nil){
				return err
			}
		}
		case commandparse.Editor_View_Range:
		{
			start := cmd.Arguments[0].Get_Int()
			end := cmd.Arguments[1].Get_Int()
			for i := start - 1;i < end;i++{
				err := ed.view_line(i)
				if(err != nil){
					return err
				}
			}
		}
		case commandparse.Editor_View_All:
		{
			for line_number ,str := range ed.line_buf{
				fmt.Fprintf(ed.conn,"%d: %s",line_number + 1,str)
			}
		}
		
		
	}
	return nil
}
func(ed *editor)swap(cmd commandparse.Parsed_Command)error{
	line := (cmd.Arguments[0].Get_Int() - 1)
	if(line < 0 || line > len(ed.line_buf) - 1){
		return fmt.Errorf("Invaild line %d\r\n",line)
	}
	fmt.Fprintf(ed.conn,"Previous line: '%s' \r\n",ed.line_buf[line])
	n,err := ed.conn.Read(ed.user_buffer)
	if err != nil{
		return err
	}
	user_str := string(ed.user_buffer[:n])
	ed.line_buf[line] = user_str
	ed.current_line = line
	return nil
}
func(ed *editor)eval_command(cmd commandparse.Parsed_Command)error{
	switch(cmd.Command_Type){
		case commandparse.Editor_Append:
		{
			err := ed.handle_append(cmd)
			if(err != nil){
				return err
			}
		}
		case commandparse.Editor_View:
		{
			err := ed.handle_view(cmd)
			if(err != nil){
				return err
			}
		}
		case commandparse.Editor_Swap:
		{
			err := ed.swap(cmd)
			if err != nil{
				return err
			}
		}
		case commandparse.Editor_Delete:
		{
			line := cmd.Arguments[0].Get_Int() - 1
			if(line < 0 || line > len(ed.line_buf) - 1){
				return fmt.Errorf("Invaild line number %d\r\n",line + 1)
			}
			ed.line_buf = slices.Delete(ed.line_buf,line,line + 1)
		}
		case commandparse.Editor_Help:
		{
			fmt.Fprint(ed.conn,editor_help)
		}
		default:
		{
			return errors.New("Unimplemented command!\r\n")
		}
	}
	return nil
}


func (ed *editor)ed_loop()(string,error){
	for{
		n,err := ed.conn.Read(ed.user_buffer)
		if(err != nil){
			return "",err
		}
		user_in := string(ed.user_buffer[:n])
		user_in = strings.Trim(user_in,"\r\n")
		cmd,err := commandparse.Parse_Command(user_in,commandparse.CommandTypeEditor)
		if(err != nil){
			ed.conn.Write([]byte(err.Error()))
		}else{
			if(cmd.Command_Type == commandparse.Editor_Exit){
				break
			}
			err := ed.eval_command(cmd)
			if(err != nil){
				fmt.Fprint(ed.conn,err.Error())
			}
		}
	}
	fmt.Fprint(ed.conn,"If you would like to discard this post please type 'n' otherwise press any other char to post.\r\n")
	n,err := ed.conn.Read(ed.user_buffer)
	if(err != nil){
		return  "",err
	}
	if(strings.Trim(string(ed.user_buffer[:n]),"\r\n") == "n"){
		return "",nil
	}
	builder := strings.Builder{}
	for _,str := range ed.line_buf{
		builder.WriteString(str)
	}
	return builder.String(),nil
}
