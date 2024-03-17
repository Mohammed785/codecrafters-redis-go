package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handleConnection(conn net.Conn, app *config.App) {
	parser := new(resp.RESPParser)
	defer conn.Close()
	data := make([]byte, 2048)
	for {
		n, err := conn.Read(data)
		if err != nil {
			if err == io.EOF {
				log.Printf("connection %v closed", conn.RemoteAddr())
				return
			}
		}
		parser.SetStream(data[:n])
		parsed, err := parser.Parse()
		if err != nil {
			conn.Write(resp.SimpleError(err.Error()).Serialize())
			continue
		}
		if len(parsed) == 0 {
			continue
		}
		cmd, err := command.NewCommandFromArray(parsed, conn, app)
		if err != nil {
			conn.Write(resp.SimpleError(err.Error()).Serialize())
			continue
		}
		cmd.Execute()
	}
}
func SendHandshake(app *config.App) {
	address := fmt.Sprintf("%s:%s", app.Params.MasterHost, app.Params.MasterPort)
	m, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalln("couldn't connect to master at ", address)
	}
	app.MasterConn = m
	m.Write(resp.ConstructRespArray([]string{"ping"}))
	m.Write(resp.ConstructRespArray([]string{"REPLCONF", "listening-port", app.Params.Port}))
	m.Write(resp.ConstructRespArray([]string{"REPLCONF", "capa", "psync2"}))
	m.Write(resp.ConstructRespArray([]string{"PSYNC", "?", "-1"}))
}

func main() {
	app := &config.App{
		Params: config.Params{Role: "master", MasterReplOffset: 0, MasterReplId: "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"},
	}
	var replicaof string
	flag.StringVar(&app.Params.Port, "port", "6379", "tcp server port number")
	flag.StringVar(&replicaof, "replicaof", "", "run the instance in slave mode")
	flag.Parse()

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", app.Params.Port))
	if err != nil {
		log.Fatalln("Failed to bind to port ", app.Params.Port)
	}
	defer l.Close()
	app.DB = storage.NewStorage()
	if replicaof != "" {
		app.Params.Role = "slave"
		app.Params.MasterHost = replicaof
		app.Params.MasterPort = flag.Arg(0)
		SendHandshake(app)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn, app)
	}
}
