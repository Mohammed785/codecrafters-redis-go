package storage

import "github.com/codecrafters-io/redis-starter-go/app/resp"

type Storage struct {
	db map[string]string
}

func NewStorage()*Storage{
	return &Storage{
		db: make(map[string]string),
	}
}

func (s *Storage) Get(key string) resp.BulkString {
	val, ok := s.db[key]
	if !ok {
		return resp.BulkString{}
	}
	return resp.BulkString([]byte(val))
}

func (s *Storage) Set(key, val string) []byte {
	s.db[key] = val
	return []byte("+OK\r\n")
}
