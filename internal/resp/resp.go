package resp

import (
	"fmt"
)

var (
	RESPNIL = []byte("$-1\r\n")
)

func RESPEncode(value any) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(v), v), nil
	case int:
		return fmt.Appendf(nil, ":%v\r\n", v), nil
	case []int:
		tmp := make([]any, len(v))
		for i, v := range v {
			tmp[i] = v
		}
		return RESPEncode(tmp)
	case []string:
		tmp := make([]any, len(v))
		for i, v := range v {
			tmp[i] = v
		}
		return RESPEncode(tmp)
	case []any:
		result := fmt.Appendf(nil, "*%d\r\n", len(v))
		for _, elem := range v {
			elemBytes, err := RESPEncode(elem)
			if err != nil {
				return nil, fmt.Errorf("error encoding element: %v", err)
			}
			result = append(result, elemBytes...)
		}
		return result, nil
	case nil:
		return RESPNIL, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}
