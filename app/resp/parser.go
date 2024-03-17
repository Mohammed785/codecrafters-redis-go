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

func (r *RESPParser) SetStream(s []byte) {
	r.stream = s
}
func (r *RESPParser) readByte() (byte, error) {
	if len(r.stream) == 0 {
		return 0, io.EOF
	}
	b := r.stream[0]
	r.stream = r.stream[1:]
	return b, nil
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

func (r *RESPParser) Parse() ([]BulkString, error) {
	t, err := r.readByte()
	if err != nil {
		return nil, err
	}
	if t!=ARRAY{
		return nil,fmt.Errorf("Err invalid syntax")
	}
	data, err := r.readUntilCRLF()
	if err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(string(data))
	if err != nil {
		return nil, err
	}
	parsed := make([]BulkString, 0, length)
	for ; length-1 > -1; length-- {
		item, err := r.parseBulkString()
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, item)
	}
	return parsed, nil
}

func (r *RESPParser) parseBulkString() (BulkString, error) {
	d, err := r.readUntilCRLF()
	length,_ := strconv.Atoi(string(d))
	if length==-1||err != nil{
		return BulkString(""), err
	}
	data, err := r.readUntilCRLF()
	return BulkString(data), err
}
