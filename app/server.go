package main

import (
	"fmt"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	data := make([]byte, 2048)
	defer conn.Close()
	for {
		n, err := conn.Read(data)
		if err != nil {
			fmt.Println("Error reading from connection ", err.Error())
			break
		}
		if n == 0 {
			break
		}
		// msg := string(data[:n])
		conn.Write([]byte("+PONG\r\n"))
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatalln("Failed to bind to port 6379")
	}
	for{
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting connection: ", err.Error())
		}
		go handleConnection(conn)
	}
}
