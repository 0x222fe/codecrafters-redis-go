package request

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

type CommandKey string
type CommandType int
type Command struct {
	Name CommandKey
	Args []string
}

const (
	CmdTypeRead CommandType = iota
	CmdTypeWrite
)

const (
	PING        CommandKey = "PING"
	ECHO        CommandKey = "ECHO"
	SET         CommandKey = "SET"
	GET         CommandKey = "GET"
	CONFIG      CommandKey = "CONFIG"
	KEYS        CommandKey = "KEYS"
	INFO        CommandKey = "INFO"
	REPLCONF    CommandKey = "REPLCONF"
	PSYNC       CommandKey = "PSYNC"
	WAIT        CommandKey = "WAIT"
	TYPE        CommandKey = "TYPE"
	XADD        CommandKey = "XADD"
	XRANGE      CommandKey = "XRANGE"
	XREAD       CommandKey = "XREAD"
	INCR        CommandKey = "INCR"
	MULTI       CommandKey = "MULTI"
	EXEC        CommandKey = "EXEC"
	DISCARD     CommandKey = "DISCARD"
	LPUSH       CommandKey = "LPUSH"
	RPUSH       CommandKey = "RPUSH"
	LRANGE      CommandKey = "LRANGE"
	LLEN        CommandKey = "LLEN"
	LPOP        CommandKey = "LPOP"
	BLPOP       CommandKey = "BLPOP"
	RPOP        CommandKey = "RPOP"
	SUBSCRIBE   CommandKey = "SUBSCRIBE"
	UNSUBSCRIBE CommandKey = "UNSUBSCRIBE"
	PUBLISH     CommandKey = "PUBLISH"
	ZADD        CommandKey = "ZADD"
	ZRANK       CommandKey = "ZRANK"
	ZRANGE      CommandKey = "ZRANGE"
	ZCARD       CommandKey = "ZCARD"
	ZSCORE      CommandKey = "ZSCORE"
	ZREM        CommandKey = "ZREM"
	GEOADD      CommandKey = "GEOADD"
	GEOPOS      CommandKey = "GEOPOS"
	GEODIST     CommandKey = "GEODIST"
	GEOSEARCH   CommandKey = "GEOSEARCH"
	ACL         CommandKey = "ACL"
	AUTH        CommandKey = "AUTH"
)

func ParseCommandFromRESP(v resp.RESPValue) (Command, error) {
	arr, ok := v.GetArrayValue()
	if !ok {
		return Command{}, fmt.Errorf("expected RESP array, got %s", v.GetType())
	}

	if len(arr) < 1 {
		return Command{}, errors.New("command array must have at least one element")
	}

	cmdName, ok := arr[0].GetBulkStringValue()
	if !ok || cmdName == nil {
		return Command{}, errors.New("first element of command array must be a bulk string")
	}

	args := make([]string, 0, len(arr)-1)
	for _, v := range arr[1:] {
		arg, ok := v.GetBulkStringValue()
		if !ok {
			return Command{}, fmt.Errorf("command argument %v is not a string", v)
		}

		if arg == nil {
			return Command{}, errors.New("command argument cannot be nil")
		}

		args = append(args, *arg)
	}

	return Command{
		Name: CommandKey(strings.ToUpper(*cmdName)),
		Args: args,
	}, nil
}
