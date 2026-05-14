
package commandparse

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)
const(
	Command_View = iota + 1
	Command_Post
	Command_Next
	Command_Goto
	Command_Exit
)

var command_table = map[string]int{
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
	TOK_SYMBOL = iota
	TOK_NUM
	TOK_DATE
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
func (parser *command_parser)parse()error{
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

func Parse_Command(user_in string)(cmd Parsed_Command,err error){
	toks,err := LexInput(user_in)
	if(err != nil){
		return cmd,err
	}
	parser := command_parser{}
	parser.input = toks
	err = parser.parse()
	if(err != nil){
		return cmd,err
	}
	return parser.command,nil
}
