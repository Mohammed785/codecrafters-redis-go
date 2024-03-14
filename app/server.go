package main

import (
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleConnection(conn net.Conn) {
	parser := new(resp.RESPParser)
	defer conn.Close()
	data := make([]byte, 1024)
	for {
		n, err := conn.Read(data)
		if err != nil {
			fmt.Println("Error reading from connection ", err.Error())
			return
		}
		data=data[:n]
		parser.SetStream(data)
		parsed, err := parser.Parse()
		if err != nil {
			conn.Write([]byte(fmt.Sprintf("-%v\r\n", err.Error())))
			continue
		}
		switch parsed := parsed.(type) {
		case resp.RespArray:
			if len(parsed) == 0 {
				continue
			}
			cmd,err:=command.NewCommandFromArray(parsed)
			if err!=nil{
				conn.Write(resp.SimpleError([]byte(err.Error())).Serialize())
				continue
			}
			cmd.Execute(conn)
		default:
			conn.Write([]byte("-invalid message\r\n")) // for now

		}
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatalln("Failed to bind to port 6379")
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}
