package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handleConnection(conn net.Conn, app *config.App) {
	parser := new(resp.RESPParser)
	defer conn.Close()
	data := make([]byte, 1024)
	for {
		n, err := conn.Read(data)
		if err != nil {
			if err==io.EOF{
				log.Printf("connection %v closed",conn.RemoteAddr())
			}else{
				log.Println("Error reading from connection ", err.Error())
			}
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
			cmd.Execute(conn, app)
		default:
			conn.Write([]byte("-invalid message\r\n")) // for now

		}
	}
}

func main() {
	app := &config.App{
		Params: config.Params{Role: "master",MasterReplOffset: 0,MasterReplId: "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"},
	}
	var replicaof string
	flag.IntVar(&app.Params.Port, "port", 6379, "tcp server port number")
	flag.StringVar(&replicaof, "replicaof", "", "run the instance in slave mode")
	flag.Parse()
	if replicaof != "" {
		app.Params.Role = "slave"
		app.Params.MasterHost = replicaof
		port, err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			log.Fatalln("invalid master port")
		}
		app.Params.MasterPort = port
	}

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", app.Params.Port))
	if err != nil {
		log.Fatalln("Failed to bind to port ", app.Params.Port)
	}
	defer l.Close()
	app.DB = storage.NewStorage()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn, app)
	}
}
