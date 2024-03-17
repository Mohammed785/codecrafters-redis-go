package command

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Command struct {
	Name    string
	Options []string
	Args    map[string]string
}

type Arg struct {
	Name  string
	Value string
}

var SET_OPTIONAL = [][]string{{"nx", "xx"}, {"ex", "px", "exat", "pxat", "keepttl"}}

func NewCommandFromArray(arr []resp.BulkString) (*Command, error) {
	name := strings.ToLower(string(arr[0]))
	cmd := &Command{Name: name}
	switch name {
	case "ping":
		if len(arr) > 1 {
			cmd.Options = append(cmd.Options, string(arr[1]))
		}
	case "echo":
		if len(arr) != 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for 'echo' command")
		}
		cmd.Options = append(cmd.Options, string(arr[1]))
	case "get":
		if len(arr) != 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for 'get' command")
		}
		cmd.Options = append(cmd.Options, string(arr[1]))
	case "set":
		if len(arr) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for 'set' command")
		}
		cmd.Options = append(cmd.Options, string(arr[1]), string(arr[2]))
		cmd.Args = make(map[string]string)
		i := 3
		for i < len(arr) {
			argLower := strings.ToLower(string(arr[i]))
			if argLower == "nx" || argLower == "xx" || argLower == "get" || argLower == "keepttl" {
				cmd.Args[argLower] = ""
				i++
			} else {
				if i+1 >= len(arr) {
					return nil, fmt.Errorf("ERR syntax error")
				}
				val := string(arr[i+1])
				_, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("ERR value is not an integer or out of range")
				}
				cmd.Args[argLower] = string(arr[i+1])
				i += 2
			}
		}
		for _, pairs := range SET_OPTIONAL {
			exists := 0
			for i := range pairs {
				if exists > 1 {
					return nil, fmt.Errorf("Err syntax error")
				}
				if _, ok := cmd.Args[pairs[i]]; ok {
					exists++
				}
			}
		}
	case "info":
		for _, section := range arr {
			cmd.Options = append(cmd.Options, string(section))
		}
	}
	return cmd, nil
}

func (c *Command) Execute(conn net.Conn, app *config.App) {
	switch c.Name {
	case "ping":
		if len(c.Options) == 0 {
			conn.Write(resp.SimpleString("PONG").Serialize())
		} else {
			conn.Write(resp.BulkString(c.Options[0]).Serialize())
		}
	case "echo":
		conn.Write(resp.BulkString(c.Options[0]).Serialize())
	case "get":
		value, ok := app.DB.Get(c.Options[0])
		if !ok {
			conn.Write(resp.BulkString.SerializeNull(""))
			return
		}
		conn.Write(resp.BulkString(value).Serialize())
	case "set":
		res := app.DB.Set(c.Options[0], c.Options[1], c.Args)
		conn.Write(res)
	case "info":
		conn.Write(resp.BulkString(app.Params.Replication()).Serialize())
	case "replconf":
		conn.Write(resp.SimpleString("OK").Serialize())
	case "psync":
		conn.Write(resp.SimpleString(fmt.Sprintf("FULLRESYNC %s 0", app.Params.MasterReplId)).Serialize())
		err:=app.FullResynchronization(conn)
		if err!=nil{
			log.Fatalln("couldn't perform full resynchronization with: ",conn.RemoteAddr())
		}
	}

}
