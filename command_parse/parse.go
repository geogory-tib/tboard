
package commandparse

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const CommandTypeBoard = 0
const CommandTypeEditor = 1

const(
	Command_View = iota + 1
	Command_Post
	Command_Next
	Command_Goto
	Command_Exit
)

var command_table map[string]int

var board_commands = map[string]int{
	"view": Command_View,
	"post": Command_Post,
	"next": Command_Next,
	"goto": Command_Goto,
	"exit": Command_Exit,
}
const(
	View_Default = iota
	View_PostID
	View_Date
	View_Range
)
var view_table = map[string]int{
	"post_id": View_PostID,
	"date": View_Date,
	"range": View_Range,
}
const(
	Goto_PostID = iota
	Goto_Date
)

const(
	Next_Default = iota
	Next_Skip
)
const(
	Editor_View = iota + 1
	Editor_Append
	Editor_Swap
	Editor_Delete
	Editor_Exit
	Editor_Help
)
var editor_commands = map[string]int{
	"view": Editor_View,
	"append": Editor_Append,
	"swap": Editor_Swap,
	"delete": Editor_Delete,
	"exit": Editor_Exit,
	"help": Editor_Help,
}
const(
	Editor_View_Default = iota
	Editor_View_Line
	Editor_View_Range
	Editor_View_All
)

const(
	Editor_Append_Default = iota
	Editor_Append_Line
)

const(
	TOK_SYMBOL = iota
	TOK_NUM
	TOK_DATE
	TOK_STAR
	TOK_AMB
	TOK_EOF
)

type token struct{
	raw string
	tok_type int
	val int
}

type lexer struct{
	input string
	current_pos int
	tokens []token
}

func(lex *lexer)pull_ch() (ch byte){
	if(lex.current_pos >= len(lex.input)){
		return 0
	}
	ch = lex.input[lex.current_pos]
	lex.current_pos++
	return
}
func(lex *lexer)peek_ch() (ch byte){
	if(lex.current_pos >= len(lex.input)){
		return 0
	}
	ch = lex.input[lex.current_pos]
	return
}

func(lex *lexer)lex_symbol(){
	tok := token{}
	start := lex.current_pos
	for ch := lex.pull_ch(); unicode.IsLetter(rune(ch)) || unicode.IsNumber(rune(ch)) || ch == '_'; ch = lex.peek_ch(){
		lex.pull_ch()
	}
	tok.raw = lex.input[start:lex.current_pos]
	if(tok.raw == "today"){
		tok.tok_type = TOK_DATE
	}else{
		tok.tok_type =  TOK_SYMBOL
	}
	lex.tokens = append(lex.tokens,tok)
}
func(lex *lexer)check_if_date() bool{
	index := lex.current_pos
	if(len(lex.input[lex.current_pos:len(lex.input)]) < len(time.DateOnly)){
		return false
	}
	for i := 0;i < 4;i++{
		if(!unicode.IsDigit(rune(lex.input[i + index]))){
			return false
		}
	}
	index += 4
	if(lex.input[index] != '-'){
		return false
	}
	index++
	for i := 0; i < 2;i++{
		if(!unicode.IsDigit(rune(lex.input[i + index]))){
			return false
		}	
	}
	index += 2
	if(lex.input[index] != '-'){
		return false
	}
	index++
	for i := 0; i < 2;i++{
		if(!unicode.IsDigit(rune(lex.input[i + index]))){
			return false
		}	
	}
	return true
}

func(lex *lexer)lex_num()error{
	start := lex.current_pos
	tok := token{}
	tok.tok_type = TOK_NUM
	if(lex.input[start] == '-'){
		lex.pull_ch()
	}
	for ch := lex.peek_ch();unicode.IsDigit(rune(ch));ch = lex.peek_ch(){
		lex.pull_ch()
	}
	tok.raw = lex.input[start:lex.current_pos]
	num_val,err := strconv.Atoi(tok.raw)
	if(err != nil){
		return err
	}
	tok.val = num_val
	lex.tokens = append(lex.tokens,tok)
	return nil
}

func(lex *lexer)lex()error{
	for{
		ch := lex.peek_ch()
		if(unicode.IsLetter(rune(ch))){
			lex.lex_symbol()
		}else if(unicode.IsNumber(rune(ch)) || ch == '-'){
			if(lex.check_if_date()){
				tok := token{}
				start := lex.current_pos
				for _ = range len(time.DateOnly){
					lex.pull_ch()
				}
				tok.tok_type = TOK_DATE
				tok.raw = lex.input[start:lex.current_pos]
				lex.tokens = append(lex.tokens,tok)
			}else{
				err := lex.lex_num()
				if err != nil{
					return err
				}
			}
		}else{
			switch(ch){
				case ' ':
				{
					lex.pull_ch()
				}
				case '*':
				{
					tok := token{
						tok_type:TOK_STAR,
						raw:lex.input[lex.current_pos:lex.current_pos],
						val:0,
					}
					lex.tokens = append(lex.tokens,tok)
					lex.pull_ch()
				}
				case '&':
				{
					tok := token{
						tok_type:TOK_STAR,
						raw:lex.input[lex.current_pos:lex.current_pos],
						val:0,
					}
					lex.tokens = append(lex.tokens,tok)
					lex.pull_ch()
				}
				case 0:
				{
					tok := token{}
					tok.tok_type = TOK_EOF
					lex.tokens = append(lex.tokens,tok)
					goto end
				}
				default:
				{
					err := errors.New(fmt.Sprintf("Invaild character '%c'\n\r",ch))
					return err
				}
			}
		}
	}
	end:
	return nil
}
func LexInput(input string)(toks []token, err error){
	lex := lexer{}
	lex.input = input
	lex.lex()
	return lex.tokens,nil
}
const(
	Arg_String = iota
	Arg_Int
)
type Arg struct{
	Tag int
	type_s *string
	type_i *int
}
func(a Arg)Get_String()(val string){
	if(a.Tag != Arg_String){
		panic("UNEXPECTED TYPE: THIS SHOULD BE UNREACHABEL")
	}
	return *a.type_s
}
func(a Arg)Get_Int()(val int){
	if(a.Tag != Arg_Int){
		panic("UNEXPECTED TYPE: THIS SHOULD BE UNREACHABEL")
	}
	return *a.type_i
}
type Parsed_Command struct{
	Command_Type int
	SubCom_Type  int
	Arguments []Arg
}

type command_parser struct{
	input []token
	current_pos int
	command Parsed_Command
}
func(parser *command_parser)pull_tok() (tok token){
	if(parser.current_pos >= len(parser.input)){
		return token{tok_type:TOK_EOF}
	}
	tok = parser.input[parser.current_pos]
	parser.current_pos++
	return
}
func(parser *command_parser)peek_tok() (tok token){
	if(parser.current_pos >= len(parser.input)){
		return token{tok_type:TOK_EOF}
	}
	tok = parser.input[parser.current_pos]
	return
}

func (parser *command_parser)parse_goto()error{
	sub_tok := parser.pull_tok()
	switch(sub_tok.tok_type){
		case TOK_DATE:
		{			
			parser.command.SubCom_Type = Goto_Date
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_String,type_s:&sub_tok.raw})
			eof_tok := parser.pull_tok()
			if(eof_tok.tok_type != TOK_EOF){
				return fmt.Errorf("Unexpected token '%s' after argument '%s'\r\n",eof_tok.raw,sub_tok.raw)
			}
		}
		case TOK_NUM:
		{
			parser.command.SubCom_Type = Goto_PostID
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&sub_tok.val})
			eof_tok := parser.pull_tok()
			if(eof_tok.tok_type != TOK_EOF){
				return fmt.Errorf("Unexpected token '%s' after argument '%s'\r\n",eof_tok.raw,sub_tok.raw)
			}
		}
		default:
		{
			return fmt.Errorf("Invaild argument '%s' to goto, please provide a date or post id\r\n",sub_tok.raw)
		}
	}
	return nil
}

func (parser *command_parser)parser_view()error{
	sub_tok := parser.pull_tok()
	if(sub_tok.tok_type == TOK_EOF){
		//the comand should already be default so I am not going to set it
		return nil
	}
	val,ok := view_table[sub_tok.raw]
	if(!ok){
		builder := strings.Builder{}
		for cmd := range view_table{
			builder.WriteString(cmd + "\r\n")
		}
		return errors.New(fmt.Sprintf("Unknown view sub-command '%s',vald commands are\r\n%s",sub_tok.val,builder.String()))
	}
	parser.command.SubCom_Type = val
	switch(parser.command.SubCom_Type){
		case View_Date:
		{
			date_tok := parser.pull_tok()
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_String,type_s:&date_tok.raw})
			eof_tok := parser.pull_tok()
			if(eof_tok.tok_type != TOK_EOF){
				return errors.New(fmt.Sprintf("Unexpected token '%s' after argument '%s'\r\n",eof_tok.raw,sub_tok.raw))
			}
		}
		case View_Range:
		{
			range_start := parser.pull_tok()
			if(range_start.tok_type != TOK_NUM){
				return errors.New(fmt.Sprintf("Unexpected token '%s'. view range Expects 2 number arguments'\r\n",range_start.raw))
			}
			range_end := parser.pull_tok()
			if(range_start.tok_type != TOK_NUM){
				return errors.New(fmt.Sprintf("Unexpected token '%s' after '%s'. view range Expects 2 number arguments'\r\n",range_end.raw,range_start.raw))
			}
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&range_start.val})
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&range_end.val})
		}
		case View_PostID:
		{
			post_id := parser.pull_tok()
			if(post_id.tok_type != TOK_NUM){
				return errors.New(fmt.Sprintf("Unexpected token '%s' after  '%s'. expected post_id"))
			}
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&post_id.val})
		}
		default:
		{
			panic("UNREACHABLE")
		}
	}
	return nil
}
func (parser *command_parser)parse_board_cmd()error{
	primary_command_tok := parser.pull_tok()
	val,ok := command_table[primary_command_tok.raw]
	if(!ok){
		builder := strings.Builder{}
		for cmd := range command_table{
			builder.WriteString(cmd + "\r\n")
		}
		return errors.New(fmt.Sprintf("Unknown primary command '%s',vald commands are\r\n%s",primary_command_tok.raw,builder.String()))
	}
	parser.command.Command_Type = val
	switch(parser.command.Command_Type){
		case Command_Exit:
		{
			eof_tok := parser.pull_tok()
			if(eof_tok.tok_type != TOK_EOF){
				return 	errors.New(fmt.Sprintf("Unexpected token '%s' after 'exit'\r\n",eof_tok.raw))
			}
		}
		case Command_Post:
		{

			eof_tok := parser.pull_tok()
			if(eof_tok.tok_type != TOK_EOF){
				return errors.New(fmt.Sprintf("Unexpected token '%s' after 'exit'\r\n",eof_tok.raw))
			}

		}
		case Command_Next:
		{
			arg_tok := parser.pull_tok()
			if(arg_tok.tok_type == TOK_EOF){
				parser.command.SubCom_Type = Next_Default
				break;
			}else if(arg_tok.tok_type == TOK_NUM){
				parser.command.SubCom_Type = Next_Skip
				parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&arg_tok.val})
				eof_tok := parser.pull_tok()
				if(eof_tok.tok_type != TOK_EOF){
					return errors.New(fmt.Sprintf("Unexpected token '%s' after argument '%s'\r\n",eof_tok.raw,arg_tok.raw))
				}
			}else{
				return errors.New(fmt.Sprintf("Expected number or EOF. Unexepcted  token '%s' after 'next'\r\n",arg_tok.raw))
			}
		
		}
		case Command_Goto:
		{
			err := parser.parse_goto()
			if err != nil{
				return err
			}
		}
		case Command_View:
		{
			err := parser.parser_view()
			if err != nil{
				return err
			}
		}
	}
	return nil
}
func (parser *command_parser)fail_if_not_eof(prev_tok_raw string)error{
	eof_tok := parser.pull_tok()
	if(eof_tok.tok_type != TOK_EOF){
		return 	errors.New(fmt.Sprintf("Unexpected token '%s' after '%s'\r\n",eof_tok.raw,prev_tok_raw))
	}
	return nil
}
func(parser *command_parser)parse_editor_view()error{
	sub_tok := parser.pull_tok()
	switch(sub_tok.tok_type){
		case TOK_NUM:
		{
			parser.command.SubCom_Type = Editor_View_Line
			err := parser.fail_if_not_eof(sub_tok.raw)
			if(err != nil){
				return err
			}
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&sub_tok.val})
		}
		case TOK_EOF:
		{
			parser.command.SubCom_Type = Editor_View_Default
		}
		case TOK_STAR:
		{
			parser.command.SubCom_Type = Editor_View_All
			err := parser.fail_if_not_eof(sub_tok.raw)
			if(err != nil){
				return err
			}
		}
		case TOK_SYMBOL:
		{
			if(sub_tok.raw != "range"){
				return fmt.Errorf("Invaild subcommand '%s' for editor view command",sub_tok.raw)
			}
			parser.command.SubCom_Type = Editor_View_Range
			start := parser.pull_tok()
			if(start.tok_type != TOK_NUM){
				return fmt.Errorf("View range expects two number arguments '%s'\r\n",sub_tok.raw)
			}
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&start.val})
			end := parser.pull_tok()
			if(end.tok_type != TOK_NUM){
				return fmt.Errorf("View range expects two number arguments '%s'\r\n",sub_tok.raw)
			}
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&end.val})
		}
		default:
		return fmt.Errorf("Invaild subcommand '%s' for editor view command",sub_tok.raw)
	}
	return nil
}
func(parser *command_parser)parser_editor_commands()error{
	command_tok := parser.pull_tok()
	val,ok := command_table[command_tok.raw]
	if(!ok){
		if(command_tok.tok_type == TOK_NUM){
			parser.command.Command_Type = Editor_View
			parser.command.SubCom_Type = Editor_View_Line
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&command_tok.val})
			goto end
		}
		builder := strings.Builder{}
		for cmd := range command_table{
			builder.WriteString(cmd + "\r\n")
		}
		return errors.New(fmt.Sprintf("Unknown primary command '%s',vald commands are\r\n%s",command_tok.raw,builder.String()))
	}
	parser.command.Command_Type = val
	switch(val){
		case Editor_View:
		{
			err := parser.parse_editor_view()
			if err != nil{
				return err
			}
		}
		case Editor_Append:
		{
			sub_tok := parser.pull_tok()
			if(sub_tok.tok_type == TOK_EOF){
				goto end
			}else if(sub_tok.tok_type == TOK_NUM){
				err := parser.fail_if_not_eof(sub_tok.raw)
				if(err != nil){
					return err
				}
				parser.command.SubCom_Type = Editor_Append_Line
				parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&sub_tok.val})
			}else{
				return fmt.Errorf("Invaild subcommand '%s' for append command\r\n",sub_tok.raw)
			}
		}
		case Editor_Swap, Editor_Delete:
		{
			num_tok := parser.pull_tok()
			if(num_tok.tok_type != TOK_NUM){
				return fmt.Errorf("Invaild token '%s'. Swap command expects a number argument\n\r",num_tok.raw)
			}
			err := parser.fail_if_not_eof(num_tok.raw)
			if(err != nil){
				return err
			}
			parser.command.Arguments = append(parser.command.Arguments,Arg{Tag:Arg_Int,type_i:&num_tok.val})
		}
		case Editor_Exit,Editor_Help:
		{
			err := parser.fail_if_not_eof(command_tok.raw)
			if(err != nil){
				return err
			}
		}
		default:
		panic("Unreachable state")
	}
	end:
	return nil
}

func Parse_Command(user_in string,cmd_type int)(cmd Parsed_Command,err error){
	
	toks,err := LexInput(user_in)
	if(err != nil){
		return cmd,err
	}
	parser := command_parser{}
	parser.input = toks
	switch cmd_type{
		case CommandTypeBoard:
		command_table = board_commands
		err = parser.parse_board_cmd()
		case CommandTypeEditor:
		command_table = editor_commands
		err = parser.parser_editor_commands()
		default:
		panic("Invaild command type for Parse_Command Function")
	}
	if(err != nil){
		return cmd,err
	}
	return parser.command,nil
}
