package schema

import (

)

func makeFloat64(i interface{}) (float64, bool) {
	switch t := i.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case float32:
		return float64(t), true
	}
	return 0, false
}

func makeFloat32(i interface{}) (float32, bool) {
	switch t := i.(type) {
	case float64:
		return float32(t), true
	case int:
		return float32(t), true
	case int8:
		return float32(t), true
	case int16:
		return float32(t), true
	case int32:
		return float32(t), true
	case int64:
		return float32(t), true
	case uint8:
		return float32(t), true
	case uint16:
		return float32(t), true
	case uint32:
		return float32(t), true
	case uint64:
		return float32(t), true
	case float32:
		return t, true
	}
	return 0, false
}

func makeInt(i interface{}) (int, bool) {
	switch t := i.(type) {
	case float64:
		return int(t), true
	case int:
		return t, true
	case int8:
		return int(t), true
	case int16:
		return int(t), true
	case int32:
		return int(t), true
	case int64:
		return int(t), true
	case uint8:
		return int(t), true
	case uint16:
		return int(t), true
	case uint32:
		return int(t), true
	case uint64:
		return int(t), true
	case float32:
		return int(t), true
	}
	return 0, false
}

func makeInt64(i interface{}) (int64, bool) {
	switch t := i.(type) {
	case float64:
		return int64(t), true
	case int:
		return int64(t), true
	case int8:
		return int64(t), true
	case int16:
		return int64(t), true
	case int32:
		return int64(t), true
	case int64:
		return t, true
	case uint8:
		return int64(t), true
	case uint16:
		return int64(t), true
	case uint32:
		return int64(t), true
	case uint64:
		return int64(t), true
	case float32:
		return int64(t), true
	}
	return 0, false
}

func makeInt32(i interface{}) (int32, bool) {
	switch t := i.(type) {
	case float64:
		return int32(t), true
	case int:
		return int32(t), true
	case int8:
		return int32(t), true
	case int16:
		return int32(t), true
	case int32:
		return t, true
	case int64:
		return int32(t), true
	case uint8:
		return int32(t), true
	case uint16:
		return int32(t), true
	case uint32:
		return int32(t), true
	case uint64:
		return int32(t), true
	case float32:
		return int32(t), true
	}
	return 0, false
}

func makeInt16(i interface{}) (int16, bool) {
	switch t := i.(type) {
	case float64:
		return int16(t), true
	case int:
		return int16(t), true
	case int8:
		return int16(t), true
	case int16:
		return t, true
	case int32:
		return int16(t), true
	case int64:
		return int16(t), true
	case uint8:
		return int16(t), true
	case uint16:
		return int16(t), true
	case uint32:
		return int16(t), true
	case uint64:
		return int16(t), true
	case float32:
		return int16(t), true
	}
	return 0, false
}

func makeInt8(i interface{}) (int8, bool) {
	switch t := i.(type) {
	case float64:
		return int8(t), true
	case int:
		return int8(t), true
	case int8:
		return t, true
	case int16:
		return int8(t), true
	case int32:
		return int8(t), true
	case int64:
		return int8(t), true
	case uint8:
		return int8(t), true
	case uint16:
		return int8(t), true
	case uint32:
		return int8(t), true
	case uint64:
		return int8(t), true
	case float32:
		return int8(t), true
	}
	return 0, false
}

func makeUint64(i interface{}) (uint64, bool) {
	switch t := i.(type) {
	case float64:
		return uint64(t), true
	case int:
		return uint64(t), true
	case int8:
		return uint64(t), true
	case int16:
		return uint64(t), true
	case int32:
		return uint64(t), true
	case int64:
		return uint64(t), true
	case uint8:
		return uint64(t), true
	case uint16:
		return uint64(t), true
	case uint32:
		return uint64(t), true
	case uint64:
		return t, true
	case float32:
		return uint64(t), true
	}
	return 0, false
}

func makeUint32(i interface{}) (uint32, bool) {
	switch t := i.(type) {
	case float64:
		return uint32(t), true
	case int:
		return uint32(t), true
	case int8:
		return uint32(t), true
	case int16:
		return uint32(t), true
	case int32:
		return uint32(t), true
	case int64:
		return uint32(t), true
	case uint8:
		return uint32(t), true
	case uint16:
		return uint32(t), true
	case uint32:
		return t, true
	case uint64:
		return uint32(t), true
	case float32:
		return uint32(t), true
	}
	return 0, false
}

func makeUint16(i interface{}) (uint16, bool) {
	switch t := i.(type) {
	case float64:
		return uint16(t), true
	case int:
		return uint16(t), true
	case int8:
		return uint16(t), true
	case int16:
		return uint16(t), true
	case int32:
		return uint16(t), true
	case int64:
		return uint16(t), true
	case uint8:
		return uint16(t), true
	case uint16:
		return t, true
	case uint32:
		return uint16(t), true
	case uint64:
		return uint16(t), true
	case float32:
		return uint16(t), true
	}
	return 0, false
}

func makeUint8(i interface{}) (uint8, bool) {
	switch t := i.(type) {
	case float64:
		return uint8(t), true
	case int:
		return uint8(t), true
	case int8:
		return uint8(t), true
	case int16:
		return uint8(t), true
	case int32:
		return uint8(t), true
	case int64:
		return uint8(t), true
	case uint8:
		return t, true
	case uint16:
		return uint8(t), true
	case uint32:
		return uint8(t), true
	case uint64:
		return uint8(t), true
	case float32:
		return uint8(t), true
	}
	return 0, false
}