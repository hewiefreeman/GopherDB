package schema

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"reflect"
	"strings"
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

// Item type query filters

var (
	queryFilters map[string]func(interface{}, interface{}, *SchemaItem)(interface{}, int)
)

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
//			> required: when true, the value must be specified when inserting (does not check on update)
//
//		    *FLOAT_TYPE*: "Float32" | "Float64"
//		- ["*FLOAT_TYPE*", defaultValue, min, max, absolute, required] : store as float32 | float64
//			> defaultValue: default value of the integer
//			> min: minimum value
//			> max: maximum value
//			> absolute: when true, the vale will always be a positive or 0 value (specifying a negative number will store it as positive)
//			> required: when true, the value must be specified when inserting (does not check on update)
//
//		- ["String", defaultValue, maxChars, required, unique] : store as string
//			> defaultValue: default value the of String
//			> maxChars: maximum characters the String can be
//			> required: when true, the value cannot be set to a blank string
//			> unique: when true, no two database entries can be assigned the same value (automatically sets required to true)
//				Note: a unique String (or a unique String Object item) inside an Array checks the containing Array, and not other database entries
//
//		- ["Array", dataType, maxItems, required] : store as []interface{}
//			> dataType: the data type of the Array's items
//			> maxItems: the maximum amount of items in the Array
//             > required: when true, there must always be items in the Array
//
//		- ["Object", schema, required] : store as map[string]interface{}
//			> schema: the schema that the Object must adhere to
//				Note: same as making the schema for a UserTable
//				Note: if an Object's parent is the database, any unique Strings in the Object with be checked against the rest of the database. Use an Array of Object to make locally (to a User entry) unique Object lists
//             > required: when true, there must always be items in the Object
//
//	Example JSON for a new schema:
//
//		{
//			"email": ["String", "", 0, true, true],
//			"friends": ["Array", ["Object", {
//								"name": ["String", "", 0, true, true],
//								"status": ["Number", 0, 0, false, 0, 0] // 0, 0 means no min/max
//			}], 50],0
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
		queryFilters = map[string]func(interface{}, interface{}, *SchemaItem)(interface{}, int){
			"Bool": boolFilter,
			"Int8": int8Filter,
			"Int16": int16Filter,
			"Int32": int32Filter,
			"Int64": int64Filter,
			"Uint8": uint8Filter,
			"Uint16": uint16Filter,
			"Uint32": uint32Filter,
			"Uint64": uint64Filter,
			"Float32": float32Filter,
			"Float64": float32Filter,
			"String": stringFilter,
			"Array": arrayFilter,
			"Object": objectFilter,
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
			to == itemTypeRefObject {
		return true
	}
	return false
}

// DataIndex gets the SchemaItem data index (table specific).
func (si *SchemaItem) DataIndex() uint32 {
	return si.dataIndex
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   QUERY ARITHMETIC   /////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func applyArithmetic(updateItem []interface{}, dbEntryData float64) (float64, int) {
	// Check format & get operator & number for math
	op, num, aErr := checkArithmeticFormat(updateItem)
	if aErr != 0 {
		return 0, aErr
	}
	// Apply arithmetic
	switch op {
		case OperatorAdd:
			return dbEntryData + num, 0

		case OperatorSub:
			return dbEntryData - num, 0

		case OperatorMul:
			return dbEntryData * num, 0

		case OperatorDiv:
			return dbEntryData / num, 0

		case OperatorMod:
			return float64(int(dbEntryData) % int(num)), 0
	}
	return 0, helpers.ErrorInvalidArithmeticParameters
}

func checkArithmeticFormat(updateItem []interface{}) (string, float64, int) {
	// Check format
	if len(updateItem) != 2 {
		return "", 0, helpers.ErrorInvalidArithmeticParameters
	}
	// Get operator
	var ok bool
	var op string
	if op, ok = updateItem[0].(string); !ok {
		return "", 0, helpers.ErrorInvalidArithmeticParameters
	}
	// Get number
	var num float64
	if num, ok = updateItem[1].(float64); !ok {
		return "", 0, helpers.ErrorInvalidArithmeticParameters
	}
	return op, num, 0
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   QUERY FILTER   /////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// QueryItemFilter takes in an item from a query, and filters/checks it for format/completion against the cooresponding SchemaItem data type.
func QueryItemFilter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if insertItem == nil {
		// Get default value
		defaultVal, defaultErr := defaultVal(itemType)
		if defaultErr != 0 {
			return nil, defaultErr
		}
		return defaultVal, 0
	} else {
		var iTypeErr int
		insertItem, iTypeErr = filterItemType(insertItem, dbEntryData, itemType)
		if iTypeErr != 0 {
			return nil, iTypeErr
		}
		return insertItem, 0
	}
}

func filterItemType(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	return queryFilters[itemType.typeName](insertItem, dbEntryData, itemType)
}

func boolFilter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if i, ok := insertItem.(bool); ok {
		return i, 0
	}
	return nil, helpers.ErrorInvalidItemValue
}

func int8Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic int8
	if i, ok := insertItem.(float64); ok {
		ic = int8(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, float64(dbEntryData.(int8)))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = int8(arithRes)
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Int8Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func int16Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic int16
	if i, ok := insertItem.(float64); ok {
		ic = int16(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, float64(dbEntryData.(int16)))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = int16(arithRes)
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Int16Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func int32Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic int32
	if i, ok := insertItem.(float64); ok {
		ic = int32(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, float64(dbEntryData.(int32)))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = int32(arithRes)
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Int32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func int64Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic int64
	if i, ok := insertItem.(float64); ok {
		ic = int64(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, float64(dbEntryData.(int64)))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = int64(arithRes)
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Int64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func uint8Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic uint8
	if i, ok := insertItem.(float64); ok {
		ic = uint8(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, float64(dbEntryData.(uint8)))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = uint8(arithRes)
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Uint8Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func uint16Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic uint16
	if i, ok := insertItem.(float64); ok {
		ic = uint16(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, float64(dbEntryData.(uint16)))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = uint16(arithRes)
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Uint16Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func uint32Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic uint32
	if i, ok := insertItem.(float64); ok {
		ic = uint32(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, float64(dbEntryData.(uint32)))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = uint32(arithRes)
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Uint32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func uint64Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic uint64
	if i, ok := insertItem.(float64); ok {
		ic = uint64(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, float64(dbEntryData.(uint64)))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = uint64(arithRes)
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Uint64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func float32Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic float32
	if i, ok := insertItem.(float64); ok {
		ic = float32(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, float64(dbEntryData.(float32)))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = float32(arithRes)
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Float32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic*(-1)
	}
	return ic, 0
}

func float64Filter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic float64
	if i, ok := insertItem.(float64); ok {
		ic = i
	} else if i, ok := insertItem.([]interface{}); ok {
		arithRes, arithErr := applyArithmetic(i, dbEntryData.(float64))
		if arithErr != 0 {
			return nil, arithErr
		}
		ic = arithRes
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Float64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic*(-1)
	}
	return ic, 0
}

func stringFilter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if i, ok := insertItem.(string); ok {
		it := itemType.iType.(StringItem)
		l := uint32(len(i))
		// Check length and if required
		if it.maxChars > 0 && l > it.maxChars {
			return nil, helpers.ErrorStringTooLarge
		} else if it.required && l == 0 {
			return nil, helpers.ErrorStringRequired
		}
		// Check if unique
		if it.unique {
			// unique checks !!!!!!
		}
		return i, 0
	}
	return nil, helpers.ErrorInvalidItemValue
}

func arrayFilter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if i, ok := insertItem.([]interface{}); ok {
		it := itemType.iType.(ArrayItem)
		var iTypeErr int
		// Check inner item type
		for k := 0; k < len(i); k++ {
			i[k], iTypeErr = filterItemTypeWithMap(i[k], nil, it.dataType.(*SchemaItem))
			if iTypeErr != 0 {
				return nil, iTypeErr
			}
		}
		return i, 0
	}
	return nil, helpers.ErrorInvalidItemValue
}

func objectFilter(insertItem interface{}, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if i, ok := insertItem.(map[string]interface{}); ok {
		it := itemType.iType.(ObjectItem)
		newObj := make(map[string]interface{})
		for itemName, schemaItem := range *(it.schema) {
			innerItem := i[itemName]
			var filterErr int
			newObj[itemName], filterErr = QueryItemFilter(innerItem, nil, schemaItem)
			if filterErr != 0 {
				return nil, filterErr
			}
		}
		return newObj, 0
	}
	return nil, helpers.ErrorInvalidItemValue
}