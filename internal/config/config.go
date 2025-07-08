package config

import "flag"

type Config struct {
	Dir        string
	Dbfilename string
}

var (
	Cfg Config
)

func ParseFlags() {
	flag.StringVar(&Cfg.Dir, "dir", "", "Directory to store Redis data")
	flag.StringVar(&Cfg.Dbfilename, "dbfilename", "", "Name of the Redis database file")
	flag.Parse()
}
