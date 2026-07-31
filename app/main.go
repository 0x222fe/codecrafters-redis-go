package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/command"
	"github.com/0x222fe/codecrafters-redis-go/internal/config"
	"github.com/0x222fe/codecrafters-redis-go/internal/connection"
	"github.com/0x222fe/codecrafters-redis-go/internal/handler"
	"github.com/0x222fe/codecrafters-redis-go/internal/rdb"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/user"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
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
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleConnection(c, state)
	}
}

func initRedis(cfg *config.Config) (*state.AppState, error) {
	var r *rdb.RDB
	var err error

	if cfg.Dbfilename != "" {
		filename := filepath.Join(cfg.Dir, cfg.Dbfilename)
		r, err = rdb.ReadRDBFile(filename)
		if err != nil {
			fmt.Printf("Failed to read RDB file: %s\n", err.Error())
			r = nil
		}
	}

	if cfg.AppendOnly {
		path := filepath.Join(cfg.Dir, cfg.AppendDirname)
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("create dir %s: %w", path, err)
		}

		file := filepath.Join(path, cfg.AppendFilename+".1.incr.aof")
		f, err := os.OpenFile(file, os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		f.Close()
	}

	store := r.MapToStore()

	isReplica := cfg.MasterHost != "" && cfg.MasterPort != 0

	state := state.NewAppState(
		&state.ReplicaState{
			IsReplica:           isReplica,
			MasterReplicationID: "",
			ReplicationID:       "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb", //INFO: hardcoded for now
			ReplicationOffset:   0,
		}, cfg, store)

	if isReplica {
		err = initRepHandshake(state)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize replication handshake: %w", err)
		}
	}

	return state, nil
}

func handleConnection(rawConn net.Conn, s *state.AppState) {
	defer rawConn.Close()

	defaultUser, ok := s.GetUser(user.DefaultUserName)
	if !ok || !defaultUser.ValidatePassword("") {
		defaultUser = nil
	}
	conn := connection.NewConnection(rawConn, defaultUser)

	defer func() {
		s.RemoveReplica(conn.ID)
	}()

	reader := bufio.NewReader(rawConn)

	c := client.NewClient(context.Background(), conn)

	for {
		respVal, _, err := resp.DecodeRESPInputExact(reader, resp.RESPArr)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Connection closed by client: %s\n", rawConn.RemoteAddr().String())
				return
			}

			rawConn.Write(resp.NewError(err).Bytes())
			continue
		}

		cmd, err := command.ParseCommandFromRESP(respVal)
		if err != nil {
			rawConn.Write(resp.NewError(err).Bytes())
			continue
		}
		fmt.Printf("Received command: %s\n", cmd.Name)

		err = handler.RunCommand(c, s, cmd)
		if err != nil {
			rawConn.Write(resp.NewError(err).Bytes())
			continue
		}
	}
}

func serveMaster(appState *state.AppState, rawConn net.Conn, reader *bufio.Reader) {
	defer func() {
		defer rawConn.Close()
		fmt.Println("Master connection closed")
	}()

	defaultUser, _ := appState.GetUser(user.DefaultUserName)
	conn := connection.NewConnection(rawConn, defaultUser)

	c := client.NewClient(context.Background(), conn)
	c.Propagated = true
	for {
		respVal, bytesRead, err := resp.DecodeRESPInputExact(reader, resp.RESPArr)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Lost connection to master")
				return
			}

			fmt.Printf("Error reading from master: %s\n", err.Error())
			continue
		}

		cmd, err := command.ParseCommandFromRESP(respVal)
		if err != nil {
			fmt.Printf("Error parsing command from master: %s\n", err.Error())
			continue
		}
		fmt.Println("Received command from master:", cmd.Name)

		err = handler.RunCommand(c, appState, cmd)
		if err != nil {
			fmt.Printf("Error executing command from master: %s\n", err.Error())
			continue
		}

		appState.WriteState(func(st *state.ReplicaState) {
			st.ReplicationOffset += bytesRead
		})
	}
}

func initRepHandshake(appState *state.AppState) error {
	cfg := appState.ReadCfg()
	masterAddr := net.JoinHostPort(cfg.MasterHost, strconv.Itoa(cfg.MasterPort))

	fmt.Printf("Connecting to master server at %s...\n", masterAddr)
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		return errors.New("failed to connect to master server: " + err.Error())
	}

	handShakeOk := false
	defer func() {
		if !handShakeOk {
			fmt.Printf("Handshake failed, closing connection, %s\n", conn.RemoteAddr().String())
			conn.Close()
		}
	}()

	reader := bufio.NewReader(conn)

	pingCmd := resputil.BulkStringsToRESPArray([]string{"PING"})
	_, err = conn.Write(pingCmd.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send PING command: %w", err)
	}

	res, _, err := resp.DecodeRESPInputExact(reader, resp.RESPStr)
	if err != nil {
		return fmt.Errorf("failed to read response from master server: %w", err)
	}
	if val, ok := res.GetStringValue(); !ok || val != "PONG" {
		return errors.New("unexpected response from master server, expected 'PONG', got: " + val)
	}

	replconfRes := resputil.BulkStringsToRESPArray([]string{"REPLCONF", "listening-port", strconv.Itoa(cfg.Port)})
	_, err = conn.Write(replconfRes.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send REPLCONF listening-port command: %w", err)
	}

	res, _, err = resp.DecodeRESPInputExact(reader, resp.RESPStr)
	if err != nil {
		return fmt.Errorf("failed to read response from master server: %w", err)
	}
	if val, ok := res.GetStringValue(); !ok || val != "OK" {
		return errors.New("unexpected response from master server, expected 'OK', got: " + val)
	}

	replconfRes = resputil.BulkStringsToRESPArray([]string{"REPLCONF", "capa", "psync2"})
	_, err = conn.Write(replconfRes.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send REPLCONF capa command: %w", err)
	}

	res, _, err = resp.DecodeRESPInputExact(reader, resp.RESPStr)
	if err != nil {
		return fmt.Errorf("failed to read response from master server: %w", err)
	}
	if val, ok := res.GetStringValue(); !ok || val != "OK" {
		return errors.New("unexpected response from master server, expected 'OK', got: " + val)
	}

	psyncEncoded := resputil.BulkStringsToRESPArray([]string{"PSYNC", "?", "-1"}).Bytes()
	_, err = conn.Write(psyncEncoded)
	if err != nil {
		return fmt.Errorf("failed to send PSYNC command: %w", err)
	}
	res, _, err = resp.DecodeRESPInputExact(reader, resp.RESPStr)
	if err != nil {
		return fmt.Errorf("failed to read response from master server: %w", err)
	}

	content, ok := res.GetStringValue()
	if !ok {
		return errors.New("unexpected response from master server, expected string, got: " + res.GetType())
	}

	if !strings.HasPrefix(content, "FULLRESYNC ") {
		return errors.New("unexpected response from master server, expected 'FULLRESYNC', got: " + content)
	}

	parts := strings.SplitN(content, " ", 3)

	if len(parts) != 3 {
		return errors.New("unexpected response format from master server, expected 'FULLRESYNC <replication_id> <offset>', got: " + content)
	}

	masterRepID := parts[1]
	if masterRepID == "" {
		return errors.New("replication ID cannot be empty")
	}
	repOffset, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid replication offset: %w", err)
	}

	appState.WriteState(func(s *state.ReplicaState) {
		s.MasterReplicationID = masterRepID
		s.ReplicationOffset = repOffset
	})

	flag, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read RDB file length: %w", err)
	}
	if flag != '$' {
		return errors.New("unexpected response from master server when reading RDB file length, expected '$', got: " + string(flag))
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read RDB file length: %w", err)
	}

	rdbLen, err := strconv.Atoi(strings.TrimSuffix(line, "\r\n"))
	if err != nil {
		return fmt.Errorf("invalid RDB file length: %w", err)
	}

	if rdbLen <= 0 {
		return errors.New("invalid RDB file length, must be greater than 0")
	}

	rdbBytes := make([]byte, rdbLen)
	_, err = io.ReadFull(reader, rdbBytes)
	if err != nil {
		return fmt.Errorf("failed to read RDB file: %w", err)
	}

	r := bytes.NewReader(rdbBytes)
	rdbData, err := rdb.ParseRDB(r)

	store := rdbData.MapToStore()

	appState.SetStore(store)

	fmt.Printf("Connected to master server at %s\n", masterAddr)

	handShakeOk = true
	go serveMaster(appState, conn, reader)

	return nil
}
