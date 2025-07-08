package redis

import "flag"

type config struct {
	dir        string
	dbfilename string
}

var (
	cfg config
)

func ParseFlags() {
	flag.StringVar(&cfg.dir, "dir", "", "Directory to store Redis data")
	flag.StringVar(&cfg.dbfilename, "dbfilename", "", "Name of the Redis database file")
	flag.Parse()
}
