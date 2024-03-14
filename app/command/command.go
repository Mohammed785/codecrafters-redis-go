package command

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
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

func NewCommandFromArray(arr resp.RespArray) (*Command, error) {
	name := strings.ToLower(arr[0].String())
	cmd := &Command{Name: name}
	switch name {
	case "ping":
		if len(arr) > 1 {
			cmd.Options = append(cmd.Options, arr[1].String())
		}
	case "echo":
		if len(arr) != 2 {
			return nil, fmt.Errorf("echo command needs a message")
		}
		cmd.Options = append(cmd.Options, arr[1].String())
	case "get":
		if len(arr) != 2 {
			return nil, fmt.Errorf("please provide only the key")
		}
		cmd.Options = append(cmd.Options, arr[1].String())
	case "set":
		if len(arr) < 3 {
			return nil, fmt.Errorf("please provide key and value")
		}
		cmd.Options = append(cmd.Options, arr[1].String(), arr[2].String())
		cmd.Args = make(map[string]string)
		i := 3
		for i < len(arr) {
			argLower := strings.ToLower(arr[i].String())
			if argLower == "nx" || argLower == "xx" || argLower == "get" || argLower == "keepttl" {
				cmd.Args[argLower] = ""
				i++
			} else {
				if i+1 >= len(arr) {
					return nil, fmt.Errorf("set '%v' flag needs a value", argLower)
				}
				val := arr[i+1].String()
				_, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid value '%s' for '%s'", val, argLower)
				}
				cmd.Args[argLower] = arr[i+1].String()
				i += 2
			}
		}
		for _, pairs := range SET_OPTIONAL {
			exists := 0
			for i := range pairs {
				if exists>1 {
					return nil, fmt.Errorf("you can only set one flag from %v", pairs)
				}
				if _, ok := cmd.Args[pairs[i]]; ok {
					exists++
				}
			}
		}
	}
	return cmd, nil
}

func (c *Command) Execute(conn net.Conn, db *storage.Storage) {
	switch c.Name {
	case "ping":
		if len(c.Options) == 0 {
			conn.Write(resp.SimpleString([]byte("PONG")).Serialize())
		} else {
			conn.Write(resp.BulkString([]byte(c.Options[0])).Serialize())
		}
	case "echo":
		conn.Write(resp.BulkString([]byte(c.Options[0])).Serialize())
	case "get":
		value, ok := db.Get(c.Options[0])
		if !ok {
			conn.Write([]byte("$-1\r\n"))
			return
		}
		conn.Write(resp.BulkString([]byte(value)).Serialize())
	case "set":
		res := db.Set(c.Options[0], c.Options[1], c.Args)
		conn.Write(res)
	}
}
