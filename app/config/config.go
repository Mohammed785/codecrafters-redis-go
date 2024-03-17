package config

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type Params struct {
	Port             string
	Role             string
	MasterReplId     string
	MasterReplOffset int
	MasterHost       string
	MasterPort       string
}

func (p Params) Replication() string {
	return fmt.Sprintf("# Replication\nrole:%s\nmaster_replid:%s\nmaster_repl_offset:%d", p.Role, p.MasterReplId, p.MasterReplOffset)
}

type replica struct {
	conn net.Conn
}

func (r replica) Write(b []byte) (int, error) {
	return r.conn.Write(b)
}

type App struct {
	Params      Params
	DB          *storage.Storage
	MasterConn  net.Conn
	replicas    map[string]replica
	writeBuffer [][]byte
}

func (a *App) AddReplica(conn net.Conn) {
	if a.replicas==nil{
		a.replicas = make(map[string]replica)
	}
	a.replicas[conn.RemoteAddr().String()] = replica{conn: conn}
}

func (a *App) AppendToWriteBuffer(writeCmd []byte) {
	a.writeBuffer = append(a.writeBuffer, writeCmd)
}

func (a *App) PropagateWriteBufferToAll(lastOnly bool) error {
	for k := range a.replicas {
		a.PropagateWriteBufferTo(k,lastOnly)
	}
	return nil
}

func (a *App) PropagateWriteBufferTo(address string,lastOnly bool) error {
	replica, ok := a.replicas[address]
	if !ok {
		return fmt.Errorf("couldn't propagate writes to replica at address %s", address)
	}
	if lastOnly{
		replica.Write(a.writeBuffer[len(a.writeBuffer)-1])
	}else{
		for _, cmd := range a.writeBuffer {
			replica.Write(cmd)
		}
	}
	return nil
}

func (a *App) FullResynchronization(conn net.Conn) error {
	rdbEmpty, err := hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")
	if err != nil {
		return fmt.Errorf("couldn't convert rdb empty hex to bytes: %s", err.Error())
	}
	rdbRes := append([]byte(fmt.Sprintf("$%d\r\n", len(rdbEmpty))), rdbEmpty...)
	conn.Write(rdbRes)
	return nil
}
