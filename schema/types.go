package schema

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"reflect"
)

// Item data type names
const (
	ItemTypeBool   = "Bool"
	ItemTypeInt8 = "Int8"
	ItemTypeInt16 = "Int16"
	ItemTypeInt32 = "Int32"
	ItemTypeInt64 = "Int64"
	ItemTypeUint8 = "Uint8"
	ItemTypeUint16 = "Uint16"
	ItemTypeUint32 = "Uint32"
	ItemTypeUint64 = "Uint64"
	ItemTypeFloat32 = "Float32"
	ItemTypeFloat64 = "Float64"
	ItemTypeString = "String"
	ItemTypeArray  = "Array"
	ItemTypeMap    = "Map"
	ItemTypeObject = "Object"
)

// Item data type initializers for table creation queries
var (
	itemTypeInitializor map[string][]reflect.Kind = map[string][]reflect.Kind{
		ItemTypeBool:   []reflect.Kind{reflect.Bool},
		ItemTypeInt8: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool},
		ItemTypeInt16: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool},
		ItemTypeInt32: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool},
		ItemTypeInt64: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool},
		ItemTypeUint8: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool},
		ItemTypeUint16: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool},
		ItemTypeUint32: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool},
		ItemTypeUint64: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool},
		ItemTypeFloat32: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		ItemTypeFloat64: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		ItemTypeString: []reflect.Kind{reflect.String, reflect.Float64, reflect.Bool, reflect.Bool},
		ItemTypeArray:  []reflect.Kind{reflect.Slice, reflect.Float64, reflect.Bool},
		ItemTypeMap:  []reflect.Kind{reflect.Slice, reflect.Float64, reflect.Bool},
		ItemTypeObject: []reflect.Kind{reflect.Map, reflect.Bool}}
)

// Item data type reflections
var (
	itemTypeRefBool   = reflect.TypeOf(BoolItem{})
	itemTypeRefInt8 = reflect.TypeOf(Int8Item{})
	itemTypeRefInt16 = reflect.TypeOf(Int16Item{})
	itemTypeRefInt32 = reflect.TypeOf(Int32Item{})
	itemTypeRefInt64 = reflect.TypeOf(Int64Item{})
	itemTypeRefUint8 = reflect.TypeOf(Uint8Item{})
	itemTypeRefUint16 = reflect.TypeOf(Uint16Item{})
	itemTypeRefUint32 = reflect.TypeOf(Uint32Item{})
	itemTypeRefUint64 = reflect.TypeOf(Uint64Item{})
	itemTypeRefFloat32 = reflect.TypeOf(Float32Item{})
	itemTypeRefFloat64 = reflect.TypeOf(Float64Item{})
	itemTypeRefString = reflect.TypeOf(StringItem{})
	itemTypeRefArray  = reflect.TypeOf(ArrayItem{})
	itemTypeRefMap  = reflect.TypeOf(MapItem{})
	itemTypeRefObject = reflect.TypeOf(ObjectItem{})
)

type BoolItem struct {
	defaultValue bool
}

type Int8Item struct {
	defaultValue int8
	min          int8
	max          int8
	required     bool
}

type Int16Item struct {
	defaultValue int16
	min          int16
	max          int16
	required     bool
}

type Int32Item struct {
	defaultValue int32
	min          int32
	max          int32
	required     bool
}

type Int64Item struct {
	defaultValue int64
	min          int64
	max          int64
	required     bool
}

type Uint8Item struct {
	defaultValue uint8
	min          uint8
	max          uint8
	required     bool
}

type Uint16Item struct {
	defaultValue uint16
	min          uint16
	max          uint16
	required     bool
}

type Uint32Item struct {
	defaultValue uint32
	min          uint32
	max          uint32
	required     bool
}

type Uint64Item struct {
	defaultValue uint64
	min          uint64
	max          uint64
	required     bool
}

type Float32Item struct {
	defaultValue float32
	min          float32
	max          float32
	abs          bool
	required     bool
}

type Float64Item struct {
	defaultValue float64
	min          float64
	max          float64
	abs          bool
	required     bool
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

/////////////////////////////////////////////////////////////////////////////
//   GET A DEFAULT VALUE   //////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func defaultVal(si *SchemaItem) (interface{}, int) {
	t := si.iType
	switch kind := t.(type) {
		// Bools
		case BoolItem:
			return kind.defaultValue, 0

		// Number types
		case Int8Item:
			if kind.required {
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
			if kind.required {
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
			if kind.required {
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
			if kind.required {
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
			if kind.required {
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
			if kind.required {
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
			if kind.required {
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
			if kind.required {
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
			if kind.required {
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
				i = i*(-1)
			}
			return i, 0

		case Float64Item:
			if kind.required {
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
				i = i*(-1)
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
			return []interface{}{}, 0

		// Maps
		case MapItem:
			return make(map[string]interface{}), 0

		// Objects
		case ObjectItem:
			return make(map[string]interface{}), 0

		default:
			return nil, helpers.ErrorUnexpected
	}
}