package command

import "errors"

func echoHandler(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("ECHO requires at least one argument")
	}

	response := "+" + args[0] + "\r\n"
	return []byte(response), nil
}
