package resp

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
	RESPNilBulkString = NewRESPBulkString(nil)
	RESPNilArray      = NewRESPArray(nil)
)

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
