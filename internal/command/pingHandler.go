package command

func pingHandler(_ []string) ([]byte, error) {
	return []byte("+PONG\r\n"), nil
}
