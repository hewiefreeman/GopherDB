package schema

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"reflect"
	"time"
)

// Item data type names
const (
	ItemTypeBool    = "Bool"
	ItemTypeInt8    = "Int8"
	ItemTypeInt16   = "Int16"
	ItemTypeInt32   = "Int32"
	ItemTypeInt64   = "Int64"
	ItemTypeUint8   = "Uint8"
	ItemTypeUint16  = "Uint16"
	ItemTypeUint32  = "Uint32"
	ItemTypeUint64  = "Uint64"
	ItemTypeFloat32 = "Float32"
	ItemTypeFloat64 = "Float64"
	ItemTypeString  = "String"
	ItemTypeArray   = "Array"
	ItemTypeMap     = "Map"
	ItemTypeObject  = "Object"
	ItemTypeTime    = "Time"
)

// Time formats
const (
	TimeFormatANSIC       = "Mon Jan _2 15:04:05 2006"
	TimeFormatUnixDate    = "Mon Jan _2 15:04:05 MST 2006"
	TimeFormatRubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
	TimeFormatRFC822      = "02 Jan 06 15:04 MST"
	TimeFormatRFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	TimeFormatRFC850      = "Monday, 02-Jan-06 15:04:05 MST"
	TimeFormatRFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
	TimeFormatRFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	TimeFormatRFC3339     = "2006-01-02T15:04:05Z07:00"
	TimeFormatRFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	TimeFormatKitchen     = "3:04PM"

	// Time stamp formats
	TimeFormatStamp      = "Jan _2 15:04:05"
	TimeFormatStampMilli = "Jan _2 15:04:05.000"
	TimeFormatStampMicro = "Jan _2 15:04:05.000000"
	TimeFormatStampNano  = "Jan _2 15:04:05.000000000"
)

// Item data type initializers for table creation queries
var (
	itemTypeInitializor map[string][]reflect.Kind = map[string][]reflect.Kind{
		ItemTypeBool:    []reflect.Kind{reflect.Bool},
		ItemTypeInt8:    []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		ItemTypeInt16:   []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		ItemTypeInt32:   []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		ItemTypeInt64:   []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		ItemTypeUint8:   []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		ItemTypeUint16:  []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		ItemTypeUint32:  []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		ItemTypeUint64:  []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		ItemTypeFloat32: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		ItemTypeFloat64: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		ItemTypeString:  []reflect.Kind{reflect.String, reflect.Float64, reflect.Bool, reflect.Bool},
		ItemTypeArray:   []reflect.Kind{reflect.Slice, reflect.Float64, reflect.Bool},
		ItemTypeMap:     []reflect.Kind{reflect.Slice, reflect.Float64, reflect.Bool},
		ItemTypeObject:  []reflect.Kind{reflect.Map, reflect.Bool},
		ItemTypeTime:    []reflect.Kind{reflect.String, reflect.Bool},
	}
)

// Time type format initializers for table creation queries
var (
	timeFormatInitializor map[string]string = map[string]string {
		"ANSIC": TimeFormatANSIC,
		"Unix":  TimeFormatUnixDate,
		"Ruby":  TimeFormatRubyDate,
		"RFC822": TimeFormatRFC822,
		"RFC822Z": TimeFormatRFC822Z,
		"RFC850": TimeFormatRFC850,
		"RFC1123": TimeFormatRFC1123,
		"RFC1123Z": TimeFormatRFC1123Z,
		"RFC3339": TimeFormatRFC3339,
		"RFC3339Nano": TimeFormatRFC3339Nano,
		"Kitchen": TimeFormatKitchen,
		"Stamp": TimeFormatStamp,
		"StampMilli": TimeFormatStampMilli,
		"StampMicro": TimeFormatStampMicro,
		"StampNano": TimeFormatStampNano,
	}
)

// Item data type reflections
var (
	itemTypeRefBool    = reflect.TypeOf(BoolItem{})
	itemTypeRefInt8    = reflect.TypeOf(Int8Item{})
	itemTypeRefInt16   = reflect.TypeOf(Int16Item{})
	itemTypeRefInt32   = reflect.TypeOf(Int32Item{})
	itemTypeRefInt64   = reflect.TypeOf(Int64Item{})
	itemTypeRefUint8   = reflect.TypeOf(Uint8Item{})
	itemTypeRefUint16  = reflect.TypeOf(Uint16Item{})
	itemTypeRefUint32  = reflect.TypeOf(Uint32Item{})
	itemTypeRefUint64  = reflect.TypeOf(Uint64Item{})
	itemTypeRefFloat32 = reflect.TypeOf(Float32Item{})
	itemTypeRefFloat64 = reflect.TypeOf(Float64Item{})
	itemTypeRefString  = reflect.TypeOf(StringItem{})
	itemTypeRefArray   = reflect.TypeOf(ArrayItem{})
	itemTypeRefMap     = reflect.TypeOf(MapItem{})
	itemTypeRefObject  = reflect.TypeOf(ObjectItem{})
	itemTypeRefTime    = reflect.TypeOf(TimeItem{})
)

type BoolItem struct {
	defaultValue bool
}

type Int8Item struct {
	defaultValue int8
	min          int8
	max          int8
	abs          bool
	required     bool
	unique       bool
}

type Int16Item struct {
	defaultValue int16
	min          int16
	max          int16
	abs          bool
	required     bool
	unique       bool
}

type Int32Item struct {
	defaultValue int32
	min          int32
	max          int32
	abs          bool
	required     bool
	unique       bool
}

type Int64Item struct {
	defaultValue int64
	min          int64
	max          int64
	abs          bool
	required     bool
	unique       bool
}

type Uint8Item struct {
	defaultValue uint8
	min          uint8
	max          uint8
	required     bool
	unique       bool
}

type Uint16Item struct {
	defaultValue uint16
	min          uint16
	max          uint16
	required     bool
	unique       bool
}

type Uint32Item struct {
	defaultValue uint32
	min          uint32
	max          uint32
	required     bool
	unique       bool
}

type Uint64Item struct {
	defaultValue uint64
	min          uint64
	max          uint64
	required     bool
	unique       bool
}

type Float32Item struct {
	defaultValue float32
	min          float32
	max          float32
	abs          bool
	required     bool
	unique       bool
}

type Float64Item struct {
	defaultValue float64
	min          float64
	max          float64
	abs          bool
	required     bool
	unique       bool
}

type StringItem struct {
	defaultValue string
	maxChars     uint32
	required     bool
	unique       bool
}

type ArrayItem struct {
	dataType interface{}
	maxItems uint32
	required bool
}

type MapItem struct {
	dataType interface{}
	maxItems uint32
	required bool
}

type ObjectItem struct {
	schema   *Schema
	required bool
}

type TimeItem struct {
	format       string
	required     bool
}

/////////////////////////////////////////////////////////////////////////////
//   GET A DEFAULT VALUE   //////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func defaultVal(si *SchemaItem) (interface{}, int) {
	switch kind := si.iType.(type) {
	// Bools
	case BoolItem:
		return kind.defaultValue, 0

	// Number types
	case Int8Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		return i, 0

	case Int16Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		return i, 0

	case Int32Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		return i, 0

	case Int64Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		return i, 0

	case Uint8Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		return i, 0

	case Uint16Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		return i, 0

	case Uint32Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		return i, 0

	case Uint64Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		return i, 0

	case Float32Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		if kind.abs && i < 0 {
			i = i * (-1)
		}
		return i, 0

	case Float64Item:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		i := kind.defaultValue
		if kind.min < kind.max {
			if i > kind.max {
				i = kind.max
			} else if i < kind.min {
				i = kind.min
			}
		}
		if kind.abs && i < 0 {
			i = i * (-1)
		}
		return i, 0

	// Strings
	case StringItem:
		if kind.unique {
			return nil, helpers.ErrorMissingRequiredItem
		} else if kind.required && len(kind.defaultValue) == 0 {
			return nil, helpers.ErrorMissingRequiredItem
		}
		return kind.defaultValue, 0

	// Arrays
	case ArrayItem:
		if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		return []interface{}{}, 0

	// Maps
	case MapItem:
		if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		return make(map[string]interface{}), 0

	// Objects
	case ObjectItem:
		if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		return make(map[string]interface{}), 0

	// Time
	case TimeItem:
		if kind.required {
			return nil, helpers.ErrorMissingRequiredItem
		}
		return time.Now(), 0

	default:
		return nil, helpers.ErrorUnexpected
	}
}

/////////////////////////////////////////////////////////////////////////////
//   GET TYPE FILTER   //////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func getTypeFilter(typeName string) func(*Filter)(int) {
	switch typeName {
	case ItemTypeBool: return boolFilter
	case ItemTypeInt8: return int8Filter
	case ItemTypeInt16: return int16Filter
	case ItemTypeInt32: return int32Filter
	case ItemTypeInt64: return int64Filter
	case ItemTypeUint8: return uint8Filter
	case ItemTypeUint16: return uint16Filter
	case ItemTypeUint32: return uint32Filter
	case ItemTypeUint64: return uint64Filter
	case ItemTypeFloat32: return float32Filter
	case ItemTypeFloat64: return float64Filter
	case ItemTypeString: return stringFilter
	case ItemTypeArray: return arrayFilter
	case ItemTypeMap: return mapFilter
	case ItemTypeObject: return objectFilter
	case ItemTypeTime: return timeFilter
	default: return nil
	}
}