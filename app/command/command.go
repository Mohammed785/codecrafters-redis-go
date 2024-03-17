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

	conn net.Conn
	app  *config.App
}

type Arg struct {
	Name  string
	Value string
}

var SET_OPTIONAL = [][]string{{"nx", "xx"}, {"ex", "px", "exat", "pxat", "keepttl"}}

func NewCommandFromArray(arr []resp.BulkString, conn net.Conn, app *config.App) (*Command, error) {
	name := strings.ToLower(string(arr[0]))
	cmd := &Command{Name: name, conn: conn, app: app}
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
func (c *Command) toArray()[]string{
	arr:=append([]string{c.Name},c.Options...) 
	for k,v:=range c.Args{
		arr = append(arr, k,v)
	}
	return arr
}
func (c *Command) Execute() {
	switch c.Name {
	case "ping":
		c.ping()
	case "echo":
		c.echo()
	case "get":
		c.get()
	case "set":
		c.set()
	case "info":
		c.info()
	case "replconf":
		c.replconf()
	case "psync":
		c.psync()
	}
}

func (c *Command) ping() {
	if len(c.Options) == 0 {
		c.conn.Write(resp.SimpleString("PONG").Serialize())
	} else {
		c.conn.Write(resp.BulkString(c.Options[0]).Serialize())
	}
}

func (c *Command) echo() {
	c.conn.Write(resp.BulkString(c.Options[0]).Serialize())
}

func (c *Command) get() {
	value, ok := c.app.DB.Get(c.Options[0])
	if !ok {
		c.conn.Write(resp.BulkString.SerializeNull(""))
		return
	}
	c.conn.Write(resp.BulkString(value).Serialize())
}

func (c *Command) set() {
	res := c.app.DB.Set(c.Options[0], c.Options[1], c.Args)
	c.app.AppendToWriteBuffer(resp.ConstructRespArray(c.toArray()))
	c.conn.Write(res)
}

func (c *Command) info() {
	c.conn.Write(resp.BulkString(c.app.Params.Replication()).Serialize())
}

func (c *Command) replconf() {
	c.conn.Write(resp.SimpleString("OK").Serialize())
}

func (c *Command) psync() {
	c.conn.Write(resp.SimpleString(fmt.Sprintf("FULLRESYNC %s 0", c.app.Params.MasterReplId)).Serialize())
	err := c.app.FullResynchronization(c.conn)
	if err != nil {
		log.Fatalln("couldn't perform full resynchronization with: ", c.conn.RemoteAddr())
	}
}
