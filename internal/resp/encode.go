package resp

import "fmt"

func (r RESPValue) Encode() []byte {
	switch r.valType {
	case RESPStr:
		//INFO: strVal should never be nil for respStr and respErr
		return fmt.Appendf(nil, "+%s\r\n", *r.strVal)
	case RESPErr:
		return fmt.Appendf(nil, "-ERR %s\r\n", *r.strVal)
	case RESPBulkStr:
		if r.strVal != nil {
			return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(*r.strVal), *r.strVal)
		}
		return []byte("$-1\r\n")
	case RESPInt:
		return fmt.Appendf(nil, ":%d\r\n", r.intVal)
	case RESPArr:
		if r.arrVal == nil {
			return []byte("*-1\r\n")
		}
		bytes := fmt.Appendf(nil, "*%d\r\n", len(r.arrVal))
		for _, elem := range r.arrVal {
			elemBytes := elem.Encode()
			bytes = append(bytes, elemBytes...)
		}
		return bytes
	default:
		panic(fmt.Sprintf("unknown RESP value type: %d", r.valType))
	}
}
