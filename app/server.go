package main

import (
	"fmt"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	data := make([]byte, 2048)
	tmp := make([]byte,512)
	for {
		n,_:=conn.Read(tmp)
		if n == 0{
			break
		}
		data = append(data, tmp[:n]...)
		conn.Write([]byte("+PONG\r\n"))
		fmt.Println("HI")
		
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatalln("Failed to bind to port 6379")
	}
	conn, err := l.Accept()
	if err != nil {
		log.Fatal("Error accepting connection: ", err.Error())
	}
	handleConnection(conn)

}
