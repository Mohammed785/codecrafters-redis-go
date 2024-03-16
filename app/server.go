package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handleConnection(conn net.Conn, db *storage.Storage) {
	parser := new(resp.RESPParser)
	defer conn.Close()
	data := make([]byte, 1024)
	for {
		n, err := conn.Read(data)
		if err != nil {
			fmt.Println("Error reading from connection ", err.Error())
			return
		}
		data = data[:n]
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
			cmd, err := command.NewCommandFromArray(parsed)
			if err != nil {
				conn.Write(resp.SimpleError([]byte(err.Error())).Serialize())
				continue
			}
			cmd.Execute(conn, db)
		default:
			conn.Write([]byte("-invalid message\r\n")) // for now

		}
	}
}

func main() {
	var port int
	flag.IntVar(&port, "port", 6379, "tcp server port number")
	flag.Parse()
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v",port))
	if err != nil {
		log.Fatalln("Failed to bind to port ",port)
	}
	defer l.Close()
	db := storage.NewStorage()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn, db)
	}
}
