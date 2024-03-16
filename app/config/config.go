package config

import "github.com/codecrafters-io/redis-starter-go/app/storage"



type Params struct {
	Port       int
	Master     bool
	MasterHost string
	MasterPort int
}

type App struct{
	Params Params
	DB *storage.Storage
}



