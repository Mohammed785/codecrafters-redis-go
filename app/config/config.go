package config

import (
	"fmt"

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
	return fmt.Sprintf("# Replication\nrole:%s\nmaster_replid:%s\nmaster_repl_offset:%d", p.Role,p.MasterReplId,p.MasterReplOffset)
}

type App struct {
	Params Params
	DB     *storage.Storage
}
