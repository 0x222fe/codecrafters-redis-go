package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/redis"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	redis.ParseFlags()

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		cmd, args, err := redis.Parse(reader)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
			continue
		}
		result, err := redis.RunCommand(cmd, args)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
			continue
		}
		conn.Write(result)
	}
}
