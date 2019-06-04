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
	ItemTypeObject = "Object"
)

// Arithmetic operators
const (
	OperatorAdd = "+"
	OperatorSub = "-"
	OperatorMul = "*"
	OperatorDiv = "/"
	OperatorMod = "%"
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

type ObjectItem struct {
	schema   *Schema
	required bool
}

/////////////////////////////////////////////////////////////////////////////
//   GET A DEFAULT VALUE   //////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func defaultVal(t interface{}) (interface{}, int) {
	kind := reflect.TypeOf(t)
	switch kind {
		// Bools
		case itemTypeRefBool:
			return t.(BoolItem).defaultValue, 0

		// Number types
		case itemTypeRefInt8:
			it := t.(Int8Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			return i, 0

		case itemTypeRefInt16:
			it := t.(Int16Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			return i, 0

		case itemTypeRefInt32:
			it := t.(Int32Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			return i, 0

		case itemTypeRefInt64:
			it := t.(Int64Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			return i, 0

		case itemTypeRefUint8:
			it := t.(Uint8Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			return i, 0

		case itemTypeRefUint16:
			it := t.(Uint16Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			return i, 0

		case itemTypeRefUint32:
			it := t.(Uint32Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			return i, 0

		case itemTypeRefUint64:
			it := t.(Uint64Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			return i, 0

		case itemTypeRefFloat32:
			it := t.(Float32Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			if it.abs && i < 0 {
				i = i*(-1)
			}
			return i, 0

		case itemTypeRefFloat64:
			it := t.(Float64Item)
			if it.required {
				return nil, helpers.ErrorMissingRequiredItem
			}
			i := it.defaultValue
			if it.min < it.max {
				if i > it.max {
					i = it.max
				} else if i < it.min {
					i = it.min
				}
			}
			if it.abs && i < 0 {
				i = i*(-1)
			}
			return i, 0

		// Strings
		case itemTypeRefString:
			si := t.(StringItem)
			if si.unique {
				return nil, helpers.ErrorMissingRequiredItem
			} else if si.required && len(si.defaultValue) == 0 {
				return nil, helpers.ErrorMissingRequiredItem
			}
			return si.defaultValue, 0

		// Arrays
		case itemTypeRefArray:
			return []interface{}{}, 0

		// Objects
		case itemTypeRefObject:
			return make(map[string]interface{}), 0

		default:
			return nil, helpers.ErrorUnexpected
	}
}