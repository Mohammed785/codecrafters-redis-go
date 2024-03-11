package main

import (
	"fmt"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	conn.Write([]byte("+PONG\r\n"))
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
