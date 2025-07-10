package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/command"
	"github.com/0x222fe/codecrafters-redis-go/internal/config"
	"github.com/0x222fe/codecrafters-redis-go/internal/parser"
	"github.com/0x222fe/codecrafters-redis-go/internal/rdb"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	cfg, err := config.ParseFlags()
	if err != nil {
		fmt.Printf("Failed to parse flags: %s\n", err.Error())
		os.Exit(1)
	}

	state, err := initRedis(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize Redis: %s\n", err.Error())
		os.Exit(1)
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

func initRedis(cfg *config.Config) (*state.AppState, error) {
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

	isReplica := cfg.MasterHost != "" && cfg.MasterPort != 0

	if isReplica {
		replicaAddr := net.JoinHostPort(cfg.MasterHost, strconv.Itoa(cfg.MasterPort))
		conn, err := net.Dial("tcp", replicaAddr)
		if err != nil {
			return nil, errors.New("failed to connect to master server: " + err.Error())
		}
		defer conn.Close()

		pingEncoded, err := resp.RESPEncode([]string{"PING"})
		if err != nil {
			return nil, errors.New("failed to encode PING command: " + err.Error())
		}
		_, err = conn.Write(pingEncoded)
		if err != nil {
			return nil, errors.New("failed to send PING command: " + err.Error())
		}

		reader := bufio.NewReader(conn)

		res, err := reader.ReadString('\n')
		if err != nil {
			return nil, errors.New("failed to read response from master server: " + err.Error())
		}
		if res != "+PONG\r\n" {
			return nil, errors.New("unexpected response from master server: " + res)
		}

		replconfEncoded, err := resp.RESPEncode([]string{"REPLCONF", "listening-port", strconv.Itoa(cfg.Port)})
		if err != nil {
			return nil, errors.New("failed to encode REPLCONF command: " + err.Error())
		}
		_, err = conn.Write(replconfEncoded)
		if err != nil {
			return nil, errors.New("failed to send REPLCONF command: " + err.Error())
		}

		res, err = reader.ReadString('\n')
		if err != nil {
			return nil, errors.New("failed to read response from master server: " + err.Error())
		}
		if res != "+OK\r\n" {
			return nil, errors.New("unexpected response from master server: " + res)
		}

		replconfEncoded, err = resp.RESPEncode([]string{"REPLCONF", "capa", "psync2"})
		if err != nil {
			return nil, errors.New("failed to encode REPLCONF capa command: " + err.Error())
		}
		_, err = conn.Write(replconfEncoded)
		if err != nil {
			return nil, errors.New("failed to send REPLCONF capa command: " + err.Error())
		}

		res, err = reader.ReadString('\n')
		if err != nil {
			return nil, errors.New("failed to read response from master server: " + err.Error())
		}
		if res != "+OK\r\n" {
			return nil, errors.New("unexpected response from master server: " + res)
		}

		psyncEncoded, err := resp.RESPEncode([]string{"PSYNC", "?", "-1"})
		if err != nil {
			return nil, errors.New("failed to encode PSYNC command: " + err.Error())
		}

		_, err = conn.Write(psyncEncoded)
		if err != nil {
			return nil, errors.New("failed to send PSYNC command: " + err.Error())
		}

		res, err = reader.ReadString('\n')
	}

	state := &state.AppState{
		Cfg:       cfg,
		Store:     store,
		IsReplica: isReplica,
	}

	//INFO: hardcoded for now
	state.ReplicantionID = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	state.ReplicantionOffset = 0

	return state, nil
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
		err = command.RunCommand(state, cmd, args, conn)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
			continue
		}
	}
}
