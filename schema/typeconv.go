package schema

import (
	"time"
)

func makeType(i interface{}, t *SchemaItem) (interface{}, bool) {
	switch t.typeName {
	case ItemTypeInt8: return makeInt8(i)
	case ItemTypeInt16: return makeInt16(i)
	case ItemTypeInt32: return makeInt32(i)
	case ItemTypeInt64: return makeInt64(i)
	case ItemTypeUint8: return makeUint8(i)
	case ItemTypeUint16: return makeUint16(i)
	case ItemTypeUint32: return makeUint32(i)
	case ItemTypeUint64: return makeUint64(i)
	case ItemTypeFloat32: return makeFloat32(i)
	case ItemTypeFloat64: return makeFloat64(i)
	}
	return nil, false
}

func makeTime(i interface{}, si *SchemaItem) (time.Time, bool) {
	switch t := i.(type) {
	case time.Time: return t, true
	case string:
		var ti time.Time
		var tErr error
		// Try to parse as RFC3339 (directly from storage engine)
		ti, tErr = time.Parse(TimeFormatRFC3339, t)
		// If the TimeItem's format isn't also RFC3339, try that instead
		if tErr != nil && si.iType.(TimeItem).format == TimeFormatRFC3339 {
			return time.Time{}, false
		} else if ti, tErr = time.Parse(si.iType.(TimeItem).format, t); tErr != nil {
			return time.Time{}, false
		}
		return ti, true
	default:
		return time.Time{}, false
	}
}

func makeFloat64(i interface{}) (float64, bool) {
	switch t := i.(type) {
	case float64: return t, true
	case int: return float64(t), true
	case int8: return float64(t), true
	case int16: return float64(t), true
	case int32: return float64(t), true
	case int64: return float64(t), true
	case uint8: return float64(t), true
	case uint16: return float64(t), true
	case uint32: return float64(t), true
	case uint64: return float64(t), true
	case float32: return float64(t), true
	}
	return 0, false
}

func makeFloat32(i interface{}) (float32, bool) {
	switch t := i.(type) {
	case float32: return t, true
	case float64: return float32(t), true
	case int: return float32(t), true
	case int8: return float32(t), true
	case int16: return float32(t), true
	case int32: return float32(t), true
	case int64: return float32(t), true
	case uint8: return float32(t), true
	case uint16: return float32(t), true
	case uint32: return float32(t), true
	case uint64: return float32(t), true
	}
	return 0, false
}

// TO-DO: Negative floats to negative int (t - 0.5)
func makeInt(i interface{}) (int, bool) {
	switch t := i.(type) {
	case int: return t, true
	case float64: return int(t + 0.5), true
	case int8: return int(t), true
	case int16: return int(t), true
	case int32: return int(t), true
	case int64: return int(t), true
	case uint8: return int(t), true
	case uint16: return int(t), true
	case uint32: return int(t), true
	case uint64: return int(t), true
	case float32: return int(t + 0.5), true
	}
	return 0, false
}

func makeInt64(i interface{}) (int64, bool) {
	switch t := i.(type) {
	case float64: return int64(t + 0.5), true
	case int: return int64(t), true
	case int8: return int64(t), true
	case int16: return int64(t), true
	case int32: return int64(t), true
	case int64: return t, true
	case uint8: return int64(t), true
	case uint16: return int64(t), true
	case uint32: return int64(t), true
	case uint64: return int64(t), true
	case float32: return int64(t + 0.5), true
	}
	return 0, false
}

func makeInt32(i interface{}) (int32, bool) {
	switch t := i.(type) {
	case float64: return int32(t + 0.5), true
	case int: return int32(t), true
	case int8: return int32(t), true
	case int16: return int32(t), true
	case int32: return t, true
	case int64: return int32(t), true
	case uint8: return int32(t), true
	case uint16: return int32(t), true
	case uint32: return int32(t), true
	case uint64: return int32(t), true
	case float32: return int32(t + 0.5), true
	}
	return 0, false
}

func makeInt16(i interface{}) (int16, bool) {
	switch t := i.(type) {
	case float64: return int16(t + 0.5), true
	case int: return int16(t), true
	case int8: return int16(t), true
	case int16: return t, true
	case int32: return int16(t), true
	case int64: return int16(t), true
	case uint8: return int16(t), true
	case uint16: return int16(t), true
	case uint32: return int16(t), true
	case uint64: return int16(t), true
	case float32: return int16(t + 0.5), true
	}
	return 0, false
}

func makeInt8(i interface{}) (int8, bool) {
	switch t := i.(type) {
	case float64: return int8(t + 0.5), true
	case int: return int8(t), true
	case int8: return t, true
	case int16: return int8(t), true
	case int32: return int8(t), true
	case int64: return int8(t), true
	case uint8: return int8(t), true
	case uint16: return int8(t), true
	case uint32: return int8(t), true
	case uint64: return int8(t), true
	case float32: return int8(t + 0.5), true
	}
	return 0, false
}

func makeUint64(i interface{}) (uint64, bool) {
	switch t := i.(type) {
	case float64: return uint64(t + 0.5), true
	case int: return uint64(t), true
	case int8: return uint64(t), true
	case int16: return uint64(t), true
	case int32: return uint64(t), true
	case int64: return uint64(t), true
	case uint8: return uint64(t), true
	case uint16: return uint64(t), true
	case uint32: return uint64(t), true
	case uint64: return t, true
	case float32: return uint64(t + 0.5), true
	}
	return 0, false
}

func makeUint32(i interface{}) (uint32, bool) {
	switch t := i.(type) {
	case float64: return uint32(t + 0.5), true
	case int: return uint32(t), true
	case int8: return uint32(t), true
	case int16: return uint32(t), true
	case int32: return uint32(t), true
	case int64: return uint32(t), true
	case uint8: return uint32(t), true
	case uint16: return uint32(t), true
	case uint32: return t, true
	case uint64: return uint32(t), true
	case float32: return uint32(t + 0.5), true
	}
	return 0, false
}

func makeUint16(i interface{}) (uint16, bool) {
	switch t := i.(type) {
		case float64: return uint16(t + 0.5), true
		case int: return uint16(t), true
		case int8: return uint16(t), true
		case int16: return uint16(t), true
		case int32: return uint16(t), true
		case int64: return uint16(t), true
		case uint8: return uint16(t), true
		case uint16: return t, true
		case uint32: return uint16(t), true
		case uint64: return uint16(t), true
		case float32: return uint16(t + 0.5), true
	}
	return 0, false
}

func makeUint8(i interface{}) (uint8, bool) {
	switch t := i.(type) {
	case float64: return uint8(t + 0.5), true
	case int: return uint8(t), true
	case int8: return uint8(t), true
	case int16: return uint8(t), true
	case int32: return uint8(t), true
	case int64: return uint8(t), true
	case uint8: return t, true
	case uint16: return uint8(t), true
	case uint32: return uint8(t), true
	case uint64: return uint8(t), true
	case float32: return uint8(t + 0.5), true
	}
	return 0, false
}