package resp

import (
	"fmt"
)

type respValueType int

type RESPValue struct {
	valType respValueType
	strVal  *string
	intVal  int64
	arrVal  []RESPValue
}

const (
	RESPStr respValueType = iota
	RESPInt
	RESPArr
	RESPBulkStr
)

var (
	RESPNIL           = []byte("$-1\r\n")
	RESPNilBulkString = NewRESPBulkString(nil)
	RESPNilArray      = NewRESPArray(nil)
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

func NewRESPString(s string) RESPValue {
	return RESPValue{valType: RESPStr, strVal: &s}
}

func NewRESPInt(i int64) RESPValue {
	return RESPValue{valType: RESPInt, intVal: i}
}

func NewRESPArray(arr []RESPValue) RESPValue {
	return RESPValue{valType: RESPArr, arrVal: arr}
}

func NewRESPBulkString(s *string) RESPValue {
	return RESPValue{valType: RESPBulkStr, strVal: s}
}

func (v RESPValue) GetType() string {
	switch v.valType {
	case RESPStr:
		return "RESPStr"
	case RESPInt:
		return "RESPInt"
	case RESPArr:
		return "RESPArr"
	case RESPBulkStr:
		return "RESPBulkStr"
	default:
		return "Unknown"
	}
}

func (v RESPValue) GetStringValue() (string, bool) {
	if v.valType == RESPStr {
		return *v.strVal, true
	}
	return "", false
}

func (v RESPValue) GetBulkStringValue() (*string, bool) {
	if v.valType == RESPBulkStr {
		return v.strVal, true
	}
	return nil, false
}

func (v RESPValue) GetIntValue() (int64, bool) {
	if v.valType == RESPInt {
		return v.intVal, true
	}
	return 0, false
}

func (v RESPValue) GetArrayValue() ([]RESPValue, bool) {
	if v.valType == RESPArr {
		return v.arrVal, true
	}
	return nil, false
}
