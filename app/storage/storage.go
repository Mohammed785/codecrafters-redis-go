package storage

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type entry struct {
	value string
	exp   time.Time
}

type Storage struct {
	db map[string]entry
}

func NewStorage() *Storage {
	return &Storage{
		db: make(map[string]entry),
	}
}

func (s *Storage) Get(key string) (string, bool) {
	value, ok := s.db[key]
	if !ok {
		return "", ok
	}
	if !value.exp.IsZero() && time.Now().After(value.exp) {
		delete(s.db, key)
		return "", false
	}
	return value.value, ok
}

func (s *Storage) Set(key, val string, args map[string]string) []byte {
	newValue := entry{value: val}
	oldValue, oldExists := s.db[key]
	returnOld := false
	for k, v := range args {
		t, _ := strconv.ParseInt(v, 10, 64)
		if k == "ex" {
			newValue.exp = time.Now().Add(time.Second * time.Duration(t))
		} else if k == "px" {
			newValue.exp = time.Now().Add(time.Millisecond * time.Duration(t))
		} else if k == "get" {
			returnOld = true
		} else if k == "exat" {
			newValue.exp = time.Unix(t, 0)
		} else if k == "pxat" {
			newValue.exp = time.UnixMilli(t)
		} else if k == "keepttl" && oldExists {
			newValue.exp = oldValue.exp
		} else if (k == "xx" && !oldExists) || (k == "nx" && oldExists) {
			return []byte("$-1\r\n")
		}
	}
	s.db[key] = newValue
	if returnOld {
		if !oldExists {
			return []byte("$-1\r\n")
		} else {
			return resp.BulkString([]byte(oldValue.value)).Serialize()
		}
	}
	return []byte("+OK\r\n")
}
