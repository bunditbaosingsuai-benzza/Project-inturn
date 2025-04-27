package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {

	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		panic(err)
	}
	fmt.Println("# Listening on port 5050...")

	for {

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("# Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("# New connection from %s\n", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("# Log: %s\n", line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("# Error reading from connection:", err)
	}
}
