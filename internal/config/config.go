package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

type Config struct {
	Dir         string
	Dbfilename  string
	Port        int
	ReplicaHost string
	ReplicaPort int
}

func ParseFlags() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Dir, "dir", "", "Directory to store Redis data")
	flag.StringVar(&cfg.Dbfilename, "dbfilename", "", "Name of the Redis database file")
	flag.IntVar(&cfg.Port, "port", 6379, "Port to bind the Redis server to")
	replicaof := new(string)
	flag.StringVar(replicaof, "replicaof", "", "Master server to replicate from (format: <host> <port>)")

	flag.Parse()

	if *replicaof != "" {
		parts := strings.Fields(*replicaof)
		if len(parts) != 2 {
			return nil, errors.New("replicaof must be in the format <host> <port>")
		}
		cfg.ReplicaHost = parts[0]
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, errors.New("replicaof port must be a valid integer")
		}
		cfg.ReplicaPort = port
	}

	return cfg, nil
}
