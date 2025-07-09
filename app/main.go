package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/parser"
	"github.com/codecrafters-io/redis-starter-go/internal/rdb"
	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	cfg := config.ParseFlags()

	var r *rdb.RDB
	var err error

	filename := filepath.Join(cfg.Dir, cfg.Dbfilename)
	if filename != "" {
		r, err = rdb.ReadRDBFile(filename)
		if err != nil {
			fmt.Printf("Failed to read RDB file: %s\n", err.Error())
			r = nil
		}
	}

	store := r.MapToStore()

	state := &state.AppState{
		Cfg:   cfg,
		Store: store,
	}

	l, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(cfg.Port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\r\n", cfg.Port)
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleConnection(conn, state)
	}
}

func handleConnection(conn net.Conn, state *state.AppState) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		cmd, args, err := parser.Parse(reader)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
			continue
		}
		result, err := command.RunCommand(state, cmd, args)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
			continue
		}
		conn.Write(result)
	}
}
