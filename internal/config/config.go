package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Dir        string
	Dbfilename string
	Port       int

	AppendOnly     bool
	AppendDirname  string
	AppendFilename string
	AppendFsync    string

	MasterHost string
	MasterPort int
}

func ParseFlags() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Dir, "dir", "", "Directory to store Redis data")
	flag.StringVar(&cfg.Dbfilename, "dbfilename", "", "Name of the Redis database file")
	flag.IntVar(&cfg.Port, "port", 6379, "Port to bind the Redis server to")

	var appendOnly string
	flag.StringVar(&appendOnly, "appendonly", "no", "Controls whether AOF persistence is enabled or disabled")
	flag.StringVar(&cfg.AppendDirname, "appenddirname", "appendonlydir", "The subdirectory under dir where AOF and manifest files are stored")
	flag.StringVar(&cfg.AppendFilename, "appendfilename", "appendonly.aof", "The name of the append-only file that records write operations")
	flag.StringVar(&cfg.AppendFsync, "appendfsync", "everysec", "How often buffered writes are flushed to the AOF file on disk")

	replicaof := new(string)
	flag.StringVar(replicaof, "replicaof", "", "Master server to replicate from (format: <host> <port>)")

	flag.Parse()

	switch appendOnly {
	case "yes":
		cfg.AppendOnly = true
	case "no":
		cfg.AppendOnly = false
	default:
		return nil, errors.New("appendonly must be 'yes' or 'no'")
	}

	if cfg.Dir == "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getwd: %w", err)
		}
		cfg.Dir = dir
	}

	if *replicaof != "" {
		parts := strings.Fields(*replicaof)
		if len(parts) != 2 {
			return nil, errors.New("replicaof must be in the format <host> <port>")
		}
		cfg.MasterHost = parts[0]
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, errors.New("replicaof port must be a valid integer")
		}
		cfg.MasterPort = port
	}

	return cfg, nil
}
