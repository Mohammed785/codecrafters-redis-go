package command

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Command struct {
	Name    string
	Options []string
	Args    []Arg
}

type Arg struct {
	Name  string
	Value string
}

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
	}
	return cmd, nil
}

func (c *Command) Execute(conn net.Conn) {
	switch c.Name {
	case "ping":
		if len(c.Options) == 0 {
			conn.Write(resp.SimpleString([]byte("PONG")).Serialize())
		} else {
			conn.Write(resp.BulkString([]byte(c.Options[0])).Serialize())
		}
	case "echo":
		conn.Write(resp.BulkString([]byte(c.Options[0])).Serialize())
	}
}
