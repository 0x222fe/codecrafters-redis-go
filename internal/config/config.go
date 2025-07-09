package config

import "flag"

type Config struct {
	Dir        string
	Dbfilename string
	Port       int
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Dir, "dir", "", "Directory to store Redis data")
	flag.StringVar(&cfg.Dbfilename, "dbfilename", "", "Name of the Redis database file")
	flag.IntVar(&cfg.Port, "port", 6379, "Port to bind the Redis server to")
	flag.Parse()

	return cfg
}
