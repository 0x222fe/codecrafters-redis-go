package config

import "flag"

type Config struct {
	Dir        string
	Dbfilename string
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Dir, "dir", "", "Directory to store Redis data")
	flag.StringVar(&cfg.Dbfilename, "dbfilename", "", "Name of the Redis database file")
	flag.Parse()

	return cfg
}
