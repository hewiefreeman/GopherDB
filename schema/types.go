/*
schema package Copyright 2020 Dominique Debergue

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific
language governing permissions and limitations under the License.
*/

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
	TimeFormatANSIC       = "Mon Jan _2 15:04:05 2006"            // ANSIC
	TimeFormatUnixDate    = "Mon Jan _2 15:04:05 MST 2006"        // Unix Date
	TimeFormatRubyDate    = "Mon Jan 02 15:04:05 -0700 2006"      // Ruby Date
	TimeFormatRFC822      = "02 Jan 06 15:04 MST"                 // RFC882
	TimeFormatRFC822Z     = "02 Jan 06 15:04 -0700"               // RFC822 with numeric zone
	TimeFormatRFC850      = "Monday, 02-Jan-06 15:04:05 MST"      // RFC850
	TimeFormatRFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"       // RFC1123
	TimeFormatRFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700"     // RFC1123 with numeric zone
	TimeFormatRFC3339     = "2006-01-02T15:04:05Z07:00"           // RFC3339
	TimeFormatRFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00" // RFC3339 nano
	TimeFormatKitchen     = "3:04PM"                              // Kitchen

	// Time stamp formats
	TimeFormatStamp      = "Jan _2 15:04:05"           // Time Stamp
	TimeFormatStampMilli = "Jan _2 15:04:05.000"       // Time Stamp with milliseconds
	TimeFormatStampMicro = "Jan _2 15:04:05.000000"    // Time Stamp with microseconds
	TimeFormatStampNano  = "Jan _2 15:04:05.000000000" // Time Stamp with nanoseconds
)

// Time type format initializers for table creation queries
var (
	timeFormatInitializor map[string]string = map[string]string{
		"ANSIC":       TimeFormatANSIC,
		"Unix":        TimeFormatUnixDate,
		"Ruby":        TimeFormatRubyDate,
		"RFC822":      TimeFormatRFC822,
		"RFC822Z":     TimeFormatRFC822Z,
		"RFC850":      TimeFormatRFC850,
		"RFC1123":     TimeFormatRFC1123,
		"RFC1123Z":    TimeFormatRFC1123Z,
		"RFC3339":     TimeFormatRFC3339,
		"RFC3339Nano": TimeFormatRFC3339Nano,
		"Kitchen":     TimeFormatKitchen,
		"Stamp":       TimeFormatStamp,
		"StampMilli":  TimeFormatStampMilli,
		"StampMicro":  TimeFormatStampMicro,
		"StampNano":   TimeFormatStampNano,
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
	dataType SchemaItem
	maxItems uint32
	required bool
}

type MapItem struct {
	dataType SchemaItem
	maxItems uint32
	required bool
}

type ObjectItem struct {
	schema Schema
}

type TimeItem struct {
	format   string
	required bool
}

/////////////////////////////////////////////////////////////////////////////
//   Get a default value   //////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func defaultVal(si SchemaItem) (interface{}, int) {
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
		// If min and max are the same, skip their checks
		if kind.min < kind.max {
			// Make sure the defaultValue isn't outside of min/max range
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
		o := make([]interface{}, len(kind.schema))
		for _, nsi := range kind.schema {
			var err int
			o[nsi.dataIndex], err = defaultVal(nsi)
			if err != 0 {
				return nil, err
			}
		}
		return o, 0

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
//   Item type format checks   //////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func checkTypeFormat(t string) func([]interface{}) bool {
	switch t {
	case ItemTypeBool:
		return checkBoolFormat
	case ItemTypeUint8, ItemTypeUint16,
		ItemTypeUint32, ItemTypeUint64:
		return checkNumericFormat
	case ItemTypeInt8, ItemTypeInt16,
		ItemTypeInt32, ItemTypeInt64,
		ItemTypeFloat32, ItemTypeFloat64:
		return checkNumericPlusFormat
	case ItemTypeString:
		return checkStringFormat
	case ItemTypeArray, ItemTypeMap:
		return checkListFormat
	case ItemTypeObject:
		return checkObjectFormat
	case ItemTypeTime:
		return checkTimeFormat
	default:
		return retFalse
	}
}

func retFalse(f []interface{}) bool {
	return false
}

func checkBoolFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 1 {
		return false
	}
	if _, ok := f[0].(bool); !ok {
		return false
	}
	return true
}

func checkNumericFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 5 {
		return false
	}
	if _, ok := f[0].(float64); !ok {
		return false
	}
	if _, ok := f[1].(float64); !ok {
		return false
	}
	if _, ok := f[2].(float64); !ok {
		return false
	}
	if _, ok := f[3].(bool); !ok {
		return false
	}
	if _, ok := f[4].(bool); !ok {
		return false
	}
	return true
}

func checkNumericPlusFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 6 {
		return false
	}
	if _, ok := f[0].(float64); !ok {
		return false
	}
	if _, ok := f[1].(float64); !ok {
		return false
	}
	if _, ok := f[2].(float64); !ok {
		return false
	}
	if _, ok := f[3].(bool); !ok {
		return false
	}
	if _, ok := f[4].(bool); !ok {
		return false
	}
	if _, ok := f[5].(bool); !ok {
		return false
	}
	return true
}

func checkStringFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 4 {
		return false
	}
	if _, ok := f[0].(string); !ok {
		return false
	}
	if _, ok := f[1].(float64); !ok {
		return false
	}
	if _, ok := f[2].(bool); !ok {
		return false
	}
	if _, ok := f[3].(bool); !ok {
		return false
	}
	return true
}

func checkListFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 3 {
		return false
	}
	if _, ok := f[0].([]interface{}); !ok {
		return false
	}
	if _, ok := f[1].(float64); !ok {
		return false
	}
	if _, ok := f[2].(bool); !ok {
		return false
	}
	return true
}

func checkObjectFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 1 {
		return false
	}
	// Creating new schema from query...
	if _, ok := f[0].(map[string]interface{}); ok {
		return true
	}
	// Restoring schema from config file...
	if _, ok := f[0].([]interface{}); ok {
		return true
	}
	return false
}

func checkTimeFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 2 {
		return false
	}
	if _, ok := f[0].(string); !ok {
		return false
	}
	if _, ok := f[1].(bool); !ok {
		return false
	}
	return true
}
