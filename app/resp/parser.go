package resp

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)


type RESPParser struct {
	stream []byte
}
func (r *RESPParser) SetStream(s []byte){
	r.stream = s
}
func (r *RESPParser) readByte() (byte,error) {
	if len(r.stream)==0{
		return 0,io.EOF
	}
	b := r.stream[0]
	r.stream = r.stream[1:]
	return b,nil
}

func (r *RESPParser) readUntilCRLF() ([]byte, error) {
	crlfIdx := bytes.Index(r.stream, CRLF)
	if crlfIdx == -1 {
		return nil, fmt.Errorf("couldn't find CRLF")
	}
	data := r.stream[:crlfIdx]
	r.stream = r.stream[crlfIdx+2:]
	return data, nil
}

func (r *RESPParser) Parse() (RESPItem,error){
	t,err:=r.readByte()
	if err!=nil{
		return nil,nil
	}
	switch t {
	case STRING:
		return r.parseSimpleString()
	case ERROR:
		return r.parseSimpleError()
	case INTEGER:
		return r.parseInteger()
	case BULK:
		return r.parseBulkString()
	case ARRAY:
		data,err :=r.readUntilCRLF()
		if err!=nil{
			return nil,err
		}
		length,err:=strconv.Atoi(string(data))
		if err!=nil{
			return nil,err
		}
		parsed := make(RespArray,0,length)
		for ;length-1> -1;length--{
			item,err:=r.Parse()
			if err!=nil{
				return nil,err
			}
			if item==nil{
				continue
			}
			parsed=append(parsed, item)
		}
		return parsed,nil
	}
	return nil,fmt.Errorf("couldn't identify type")
}

func (r *RESPParser) parseSimpleString() (SimpleString, error) {
	data, err := r.readUntilCRLF()
	return SimpleString(data), err
}

func (r *RESPParser) parseInteger() (Integer, error) {
	data, err := r.readUntilCRLF()
	return Integer(data), err
}

func (r *RESPParser) parseSimpleError() (SimpleError, error) {
	data, err := r.readUntilCRLF()
	return SimpleError(data), err
}

func (r *RESPParser) parseBulkString() (BulkString, error) {
	_, err := r.readUntilCRLF()
	if err != nil {
		return nil, err
	}
	data, err := r.readUntilCRLF()
	return BulkString(data), err
}
