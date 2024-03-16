package resp

import (
	"bytes"
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
	String() string
}

type SimpleString []byte
type BulkString []byte
type SimpleError []byte
type Integer []byte
type RespArray []RESPItem

func (ss SimpleString) String() string {
	return string(ss)
}

func (se SimpleError) String() string {
	return string(se)

}
func (i Integer) String() string {
	return string(i)
}

func (bs BulkString) String() string {
	return string(bs)
}

func (a RespArray) String() string {
	b := "{"
	for i := range a {
		b += fmt.Sprintf(" %s ",a[i].String())
	}
	b+="}"
	return b
}

func (ss SimpleString) Serialize() []byte {
	return bytes.Join([][]byte{{STRING}, ss, CRLF}, nil)
}

func (se SimpleError) Serialize() []byte {
	return bytes.Join([][]byte{{ERROR}, se, CRLF}, nil)

}
func (i Integer) Serialize() []byte {
	return bytes.Join([][]byte{{INTEGER}, i, CRLF}, nil)
}

func (bs BulkString) Serialize() []byte {
	if len(bs)==0{
		return bs.SerializeNull()
	}
	return bytes.Join([][]byte{[]byte(fmt.Sprintf("%c%v", BULK, len(bs))), append(bs, CRLF...)}, []byte("\r\n"))
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


func ConstructRespArray(options []string)string{
	arr := fmt.Sprintf("*%d\r\n",len(options))
	for _, v := range options {
		arr += fmt.Sprintf("$%d\r\n%s\r\n",len(v),v)
	}
	return arr

}
