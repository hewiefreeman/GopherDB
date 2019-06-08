package schema

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"reflect"
	"strings"
	"sync"
	//"fmt"
)

// Schema represents a database schema that one or more tables must adhere to.
type Schema map[string]*SchemaItem

// SchemaItem represents one of the items in a Schema. SchemaItem also holds info about the data type and it's settings.
type SchemaItem struct {
	dataIndex uint32
	typeName  string
	iType     interface{}
}

// NOTES:
//
//	Type Declarations:
//		- ["Bool", defaultValue] : store as Boolean
//			> defaultValue: default value of the Bool
//
//		    *INT_TYPE*: "Int8" | "Int16" | "Int32" | "Int64" | "Uint8" | "Uint16" | "Uint32" | "Uint64"
//		- ["*INT_TYPE*", defaultValue, min, max, required] : store as int(8-64) | uint(8-64)
//			> defaultValue: default value of the integer
//			> min: minimum value
//			> max: maximum value
//			> required: when true, the value must be specified when inserting (does not check on updates)
//
//		    *FLOAT_TYPE*: "Float32" | "Float64"
//		- ["*FLOAT_TYPE*", defaultValue, min, max, absolute, required] : store as float32 | float64
//			> defaultValue: default value of the integer
//			> min: minimum value
//			> max: maximum value
//			> absolute: when true, the vale will always be a positive or 0 value (specifying a negative number will store it as positive)
//			> required: when true, the value must be specified when inserting (does not check on updates)
//
//		- ["String", defaultValue, maxChars, required, unique] : store as string
//			> defaultValue: default value the of String
//			> maxChars: maximum characters the String can be
//			> required: when true, the value cannot be set to a blank string. When inserting, the value must be specified unless there is a valid default value
//			> unique: when true, no two database entries can be assigned the same value (automatically sets required to true)
//				Note: a unique String (or a unique String Object item) inside an Array checks the containing Array, and not other database entries
//
//		- ["Array", dataType, maxItems, required] : store as []interface{}
//			> dataType: the data type of the Array's items
//			> maxItems: the maximum amount of items in the Array
//             > required: when true, there must always be items in the Array
//
//		- ["Map", dataType, maxItems, required] : store as map[string]interface{}
//			> dataType: the data type of the Map's items
//			> maxItems: the maximum amount of items in the Map
//             > required: when true, there must always be items in the Map
//
//		- ["Object", schema, required] : store as map[string]interface{}
//			> schema: the schema that the Object must adhere to
//				Note: same as making the schema for a UserTable
//				Note: if an Object's parent is the database, any unique Strings in the Object with be checked against the rest of the database. Use an Array of Object to make locally (to a User entry) unique Object lists
//             > required: when true, the value must be specified when inserting (does not check on updates)
//
//		- ["Time", format, required] : store as time.Time (default value is current database time)
//			> format: the format of time/date the database will accept as input (eg: "Unix", "RFC3339", "Stamp" - see constants in types.go)
//			> required: when true, the value must be specified when inserting (does not check on updates)
//
//	Example JSON for a new schema:
//
//		{
//			"email": ["String", "", 0, true, true],
//			"friends": ["Array", ["Object", {
//										"name": ["String", "", 0, true, true],
//										"status": ["Uint8", 0, 0, 2, false] // default 0, min 0, max 2
//								}, false],
//						50, false],
//			"vCode": ["String", "", 0, true, false],
//			"verified": ["Bool", false]
//		}
//

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   CREATING A SCHEMA   ////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// New creates a new schema from a JSON schema object
func New(schema map[string]interface{}) (*Schema, int) {
	// INIT queryFilters
	if queryFilters == nil {
		queryFilters = map[string]func(interface{}, []string, interface{}, *SchemaItem, *sync.Mutex, interface{}) (interface{}, int){
			ItemTypeBool:    boolFilter,
			ItemTypeInt8:    int8Filter,
			ItemTypeInt16:   int16Filter,
			ItemTypeInt32:   int32Filter,
			ItemTypeInt64:   int64Filter,
			ItemTypeUint8:   uint8Filter,
			ItemTypeUint16:  uint16Filter,
			ItemTypeUint32:  uint32Filter,
			ItemTypeUint64:  uint64Filter,
			ItemTypeFloat32: float32Filter,
			ItemTypeFloat64: float32Filter,
			ItemTypeString:  stringFilter,
			ItemTypeArray:   arrayFilter,
			ItemTypeMap:     mapFilter,
			ItemTypeObject:  objectFilter,
			ItemTypeTime:    timeFilter,
		}
	}

	s := make(Schema)

	// Make Schema
	var i uint32
	for itemName, itemParams := range schema {
		// Names cannot have "." or "*"
		if strings.Contains(itemName, ".") || strings.Contains(itemName, "*") {
			return nil, helpers.ErrorSchemaInvalidItemName
		}
		// Check item format
		if params, ok := itemParams.([]interface{}); ok {
			schemaItem, iErr := makeSchemaItem(params)
			if iErr != 0 {
				return nil, iErr
			}
			schemaItem.dataIndex = i
			s[itemName] = schemaItem
			i++
		} else {
			// Invalid format
			return nil, helpers.ErrorSchemaInvalidFormat
		}
	}

	//
	return &s, 0
}

func makeSchemaItem(params []interface{}) (*SchemaItem, int) {
	if len(params) <= 1 {
		// Invalid format - requires at least a length of 2 for any item data type
		return nil, helpers.ErrorSchemaInvalidItemParameters
	}

	// Get data type
	if t, ok := params[0].(string); ok {
		dti := itemTypeInitializor[t]
		// Check for valid params length
		dtiPL := len(dti)
		if dtiPL == 0 {
			return nil, helpers.ErrorSchemaInvalidItemType
		} else if dtiPL != len(params)-1 {
			return nil, helpers.ErrorSchemaInvalidItemParameters
		}
		// Check for valid parameter data types
		for i := 0; i < dtiPL; i++ {
			if params[i+1] == nil {
				return nil, helpers.ErrorSchemaInvalidItemParameters
			}
			if reflect.TypeOf(params[i+1]).Kind() != dti[i] {
				return nil, helpers.ErrorSchemaInvalidItemParameters
			}
		}
		// Execute create for the type
		switch t {
		case ItemTypeBool:
			si := SchemaItem{typeName: t, iType: BoolItem{defaultValue: params[1].(bool)}}
			return &si, 0

		case ItemTypeInt8:
			si := SchemaItem{typeName: t, iType: Int8Item{defaultValue: int8(params[1].(float64)), min: int8(params[2].(float64)), max: int8(params[3].(float64)), required: params[4].(bool)}}
			return &si, 0

		case ItemTypeInt16:
			si := SchemaItem{typeName: t, iType: Int16Item{defaultValue: int16(params[1].(float64)), min: int16(params[2].(float64)), max: int16(params[3].(float64)), required: params[4].(bool)}}
			return &si, 0

		case ItemTypeInt32:
			si := SchemaItem{typeName: t, iType: Int32Item{defaultValue: int32(params[1].(float64)), min: int32(params[2].(float64)), max: int32(params[3].(float64)), required: params[4].(bool)}}
			return &si, 0

		case ItemTypeInt64:
			si := SchemaItem{typeName: t, iType: Int64Item{defaultValue: int64(params[1].(float64)), min: int64(params[2].(float64)), max: int64(params[3].(float64)), required: params[4].(bool)}}
			return &si, 0

		case ItemTypeUint8:
			si := SchemaItem{typeName: t, iType: Uint8Item{defaultValue: uint8(params[1].(float64)), min: uint8(params[2].(float64)), max: uint8(params[3].(float64)), required: params[4].(bool)}}
			return &si, 0

		case ItemTypeUint16:
			si := SchemaItem{typeName: t, iType: Uint16Item{defaultValue: uint16(params[1].(float64)), min: uint16(params[2].(float64)), max: uint16(params[3].(float64)), required: params[4].(bool)}}
			return &si, 0

		case ItemTypeUint32:
			si := SchemaItem{typeName: t, iType: Uint32Item{defaultValue: uint32(params[1].(float64)), min: uint32(params[2].(float64)), max: uint32(params[3].(float64)), required: params[4].(bool)}}
			return &si, 0

		case ItemTypeUint64:
			si := SchemaItem{typeName: t, iType: Uint64Item{defaultValue: uint64(params[1].(float64)), min: uint64(params[2].(float64)), max: uint64(params[3].(float64)), required: params[4].(bool)}}
			return &si, 0

		case ItemTypeFloat32:
			si := SchemaItem{typeName: t, iType: Float32Item{defaultValue: float32(params[1].(float64)), min: float32(params[2].(float64)), max: float32(params[3].(float64)), abs: params[4].(bool), required: params[5].(bool)}}
			return &si, 0

		case ItemTypeFloat64:
			si := SchemaItem{typeName: t, iType: Float64Item{defaultValue: params[1].(float64), min: params[2].(float64), max: params[3].(float64), abs: params[4].(bool), required: params[5].(bool)}}
			return &si, 0

		case ItemTypeString:
			si := SchemaItem{typeName: t, iType: StringItem{defaultValue: params[1].(string), maxChars: uint32(params[2].(float64)), required: params[3].(bool), unique: params[4].(bool)}}
			return &si, 0

		case ItemTypeArray:
			schemaItem, iErr := makeSchemaItem(params[1].([]interface{}))
			if iErr != 0 {
				return nil, iErr
			}
			si := SchemaItem{typeName: t, iType: ArrayItem{dataType: schemaItem, maxItems: uint32(params[2].(float64))}}
			return &si, 0

		case ItemTypeMap:
			schemaItem, iErr := makeSchemaItem(params[1].([]interface{}))
			if iErr != 0 {
				return nil, iErr
			}
			si := SchemaItem{typeName: t, iType: MapItem{dataType: schemaItem, maxItems: uint32(params[2].(float64))}}
			return &si, 0

		case ItemTypeObject:
			if sObj, ok := params[1].(map[string]interface{}); ok {
				schema, schemaErr := New(sObj)
				if schemaErr != 0 {
					return nil, schemaErr
				}
				si := SchemaItem{typeName: t, iType: ObjectItem{schema: schema}}
				return &si, 0
			}
			return nil, helpers.ErrorSchemaInvalidItemParameters

		case ItemTypeTime:
			var format string = timeFormatInitializor[params[1].(string)]
			if format == "" {
				return nil, helpers.ErrorSchemaInvalidTimeFormat
			}
			si := SchemaItem{typeName: t, iType: TimeItem{format: format, required: params[2].(bool)}}
			return &si, 0

		default:
			return nil, helpers.ErrorUnexpected
		}
	} else {
		return nil, helpers.ErrorSchemaInvalidFormat
	}
}

// ValidSchema checks if a *Schema is valid format
func (ts *Schema) ValidSchema() bool {
	if ts == nil || len(*ts) == 0 {
		return false
	}
	for _, v := range *ts {
		if !v.ValidSchemaItem() {
			return false
		}
	}
	return true
}

// ValidSchemaItem checks if a *SchemaItem is valid format
func (si *SchemaItem) ValidSchemaItem() bool {
	to := reflect.TypeOf(si.iType)
	if to == itemTypeRefBool ||
		to == itemTypeRefInt8 ||
		to == itemTypeRefInt16 ||
		to == itemTypeRefInt32 ||
		to == itemTypeRefInt64 ||
		to == itemTypeRefUint8 ||
		to == itemTypeRefUint16 ||
		to == itemTypeRefUint32 ||
		to == itemTypeRefUint64 ||
		to == itemTypeRefFloat32 ||
		to == itemTypeRefFloat64 ||
		to == itemTypeRefString ||
		to == itemTypeRefArray ||
		to == itemTypeRefMap ||
		to == itemTypeRefObject ||
		to == itemTypeRefTime {
		return true
	}
	return false
}

// DataIndex gets the SchemaItem data index (table specific).
func (si *SchemaItem) DataIndex() uint32 {
	return si.dataIndex
}