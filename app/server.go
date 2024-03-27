package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handleConnection(conn net.Conn, app *config.App) {
	parser := new(resp.RESPParser)
	defer conn.Close()
	data := make([]byte, 4096)
	for {
		n, err := conn.Read(data)
		if err != nil {
			log.Printf("connection %v closed %s\n", conn.RemoteAddr(), err.Error())
			break
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

func ConnectToMaster(app *config.App) {
	address := fmt.Sprintf("%s:%s", app.Configs.MasterHost, app.Configs.MasterPort)
	m, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalln("couldn't connect to master at ", address)
	}
	defer m.Close()
	data := make([]byte, 2048)
	m.Write(resp.ConstructRespArray([]string{"ping"}))
	m.Read(data)
	m.Write(resp.ConstructRespArray([]string{"REPLCONF", "listening-port", app.Configs.Port}))
	m.Read(data)
	m.Write(resp.ConstructRespArray([]string{"REPLCONF", "capa", "psync2"}))
	m.Read(data)
	m.Write(resp.ConstructRespArray([]string{"PSYNC", "?", "-1"}))
	m.Read(data)

	parser := new(resp.RESPParser)
	for {
		n, err := m.Read(data)
		if err != nil {
			fmt.Println("master connection closed")
			break
		}
		parser.SetStream(data[:n])
		parsed, err := parser.Parse()
		if err != nil {
			m.Write(resp.SimpleError(err.Error()).Serialize())
			continue
		}
		if len(parsed) == 0 {
			continue
		}
		cmd, err := command.NewCommandFromArray(parsed, m, app)
		if err != nil {
			m.Write(resp.SimpleError(err.Error()).Serialize())
			continue
		}
		cmd.Execute()
	}
}

func HandleReplicaWrites(app *config.App) {
	var wg sync.WaitGroup
	for cmd := range app.WriteQueue {
		for _, replica := range app.Replicas {
			wg.Add(1)
			go func(replica net.Conn, cmd []byte) {
				defer wg.Done()
				replica.Write(cmd)
			}(replica, cmd)
		}
		wg.Wait()
	}
}

func main() {
	app := config.NewApp()

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", app.Configs.Port))
	if err != nil {
		log.Fatalln("Failed to bind to port ", app.Configs.Port)
	}
	defer l.Close()
	app.DB = storage.NewStorage()
	if app.Configs.Role == config.RoleSlave {
		go ConnectToMaster(app)
	} else {
		go HandleReplicaWrites(app)
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
