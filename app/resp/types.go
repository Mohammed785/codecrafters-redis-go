package resp

import (
	"fmt"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

var CRLF = []byte("\r\n")

type RESPItem interface {
	Serialize() []byte
}

type SimpleString string
type BulkString string
type SimpleError string
type Integer string
type RespArray []RESPItem

func (ss SimpleString) Serialize() []byte {
	return []byte(fmt.Sprintf("%c%s%s",STRING,ss,CRLF))
}

func (se SimpleError) Serialize() []byte {
	return []byte(fmt.Sprintf("%c%s%s",ERROR,se,CRLF))	
}
func (i Integer) Serialize() []byte {
	return []byte(fmt.Sprintf("%c%s%s",INTEGER,i,CRLF))
}

func (bs BulkString) Serialize() []byte {
	if len(bs)==0{
		return bs.SerializeNull()
	}
	return []byte(fmt.Sprintf("%c%v%s%s%s",BULK,len(bs),CRLF,bs,CRLF))
}

func (bs BulkString) SerializeNull()[]byte{
	return []byte("$-1\r\n")
}

func (a RespArray) Serialize() []byte {
	b := append([]byte{}, []byte(fmt.Sprintf("%c%v\r\n", ARRAY, len(a)))...)
	for i := range a {
		b = append(b, a[i].Serialize()...)
	}
	return b
}


func ConstructRespArray(options []string)[]byte{
	arr := make(RespArray,0,len(options))
	for _, v := range options {
		arr = append(arr, BulkString(v))
	}
	return arr.Serialize()

}
