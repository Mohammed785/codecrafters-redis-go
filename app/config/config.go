package config

import (
	"encoding/hex"
	"flag"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/storage"

	"net"
	"sync"
)

const (
	RoleMaster = "master"
	RoleSlave  = "slave"
)

type Configs struct {
	Port             string
	Role             string
	MasterReplId     string
	MasterReplOffset int
	MasterHost       string
	MasterPort       string
}

func (c Configs) Replication() string {
	return fmt.Sprintf("# Replication\nrole:%s\nmaster_replid:%s\nmaster_repl_offset:%d", c.Role, c.MasterReplId, c.MasterReplOffset)
}

func NewConfig() Configs {
	cfg := Configs{Role: RoleMaster, MasterReplOffset: 0, MasterReplId: "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"}
	flag.StringVar(&cfg.Port, "port", "6379", "tcp server port number")
	flag.StringVar(&cfg.MasterHost, "replicaof", "", "run the instance in slave mode")
	flag.Parse()
	if cfg.MasterHost != "" {
		cfg.Role = RoleSlave
		cfg.MasterPort = flag.Arg(0)
	}
	return cfg
}

type App struct {
	Configs    Configs
	DB         *storage.Storage
	Replicas   []net.Conn
	WriteQueue chan []byte
	mu         sync.Mutex
}

func NewApp() *App {
	cfg := NewConfig()
	return &App{
		Configs:    cfg,
		WriteQueue: make(chan []byte, 50),
		mu:         sync.Mutex{},
	}
}

func (a *App) AddReplica(conn net.Conn) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Replicas = append(a.Replicas, conn)
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
