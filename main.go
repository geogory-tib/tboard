package main

import(
	"tboard/serve"
)

func main(){
	Server := serve.CreateServer()
	Server.ServerLoop()
}
