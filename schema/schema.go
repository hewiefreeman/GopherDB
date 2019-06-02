package schema

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"reflect"
)

type Schema map[string]*SchemaItem

type SchemaItem struct {
	dataIndex uint32
	iType     interface{}
}

// NOTES:
//
//	Type Declarations:
//		- ["Bool", defaultValue] : store as Boolean
//			> defaultValue: default value of the Bool
//
//		- ["Number", defaultValue, decimalPrecision, absolute, min, max] : store as float64
//			> defaultValue: default value of the Number
//			> decimalPrecision: precision for Number decimal
//				Note: 0 stores as int64, anything higher is float64
//			> absolute: when true, the Number can only be positive
//				Note: if a negative value is assigned to an absolute Number, it will be set to 0
//			> min: minimum value
//			> max: maximum value
//
//		- ["String", defaultValue, maxChars, required, unique] : store as string
//			> defaultValue: default value the of String
//			> maxChars: maximum characters the String can be
//			> required: when true, the value cannot be set to blank
//			> unique: when true, no two database entries can be assigned the same value
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
//	In this example we make 4 items that a

func New(schema map[string]interface{}) (*Schema, int) {
	if len(schema) == 0 {
		return nil, helpers.ErrorSchemaItemsRequired
	}
	s := make(Schema)

	// Make Schema
	var i uint32
	for itemName, itemParams := range schema {
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
		case "Bool":
			si := SchemaItem{iType: BoolItem{defaultValue: params[1].(bool)}}
			return &si, 0

		case "Number":
			si := SchemaItem{iType: NumberItem{defaultValue: params[1].(float64), precision: uint8(params[2].(float64)), abs: params[3].(bool), min: params[4].(float64), max: params[5].(float64)}}
			return &si, 0

		case "String":
			si := SchemaItem{iType: StringItem{defaultValue: params[1].(string), maxChars: uint32(params[2].(float64)), required: params[3].(bool), unique: params[4].(bool)}}
			return &si, 0

		case "Array":
			schemaItem, iErr := makeSchemaItem(params[1].([]interface{}))
			if iErr != 0 {
				return nil, iErr
			}
			si := SchemaItem{iType: ArrayItem{dataType: schemaItem, maxItems: uint32(params[2].(float64))}}
			return &si, 0

		case "Object":
			if sObj, ok := params[1].(map[string]interface{}); ok {
				schema, schemaErr := New(sObj)
				if schemaErr != 0 {
					return nil, schemaErr
				}
				si := SchemaItem{iType: ObjectItem{schema: schema}}
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

func (si *SchemaItem) ValidSchemaItem() bool {
	to := reflect.TypeOf(si.iType)
	if to == itemTypeRefBool ||
			to == itemTypeRefNumber ||
			to == itemTypeRefString ||
			to == itemTypeRefArray ||
			to == itemTypeRefObject {
		return true
	}
	return false
}

func (si *SchemaItem) ItemType() interface{} {
	return si.iType
}

func (si *SchemaItem) DataIndex() uint32 {
	return si.dataIndex
}

func SchemaFilter(insertItem interface{}, schemaItem *SchemaItem) (interface{}, int) {
	if insertItem == nil {
		// Get default value
		defaultVal, defaultErr := defaultVal(schemaItem.iType)
		if defaultErr != 0 {
			return nil, defaultErr
		}
		return defaultVal, 0
	} else {
		var iTypeErr int
		insertItem, iTypeErr = filterItemType(insertItem, schemaItem.iType)
		if iTypeErr != 0 {
			return nil, iTypeErr
		}
		return insertItem, 0
	}
}

func filterItemType(item interface{}, itemType interface{}) (interface{}, int) {
	kind := reflect.TypeOf(itemType)
	switch kind {
		// Handle Bools
		case itemTypeRefBool:
			if i, ok := item.(bool); ok {
				return i, 0
			}
			return nil, helpers.ErrorInvalidItemValue

		// Handle Numbers
		case itemTypeRefNumber:
			if i, ok := item.(float64); ok {
				it := itemType.(NumberItem)
				// Check min/max unless both are the same
				if it.min < it.max {
					if i > it.max {
						i = it.max
					} else if i < it.min {
						i = it.min
					}
				}

				return i, 0
			}
			return nil, helpers.ErrorInvalidItemValue

		// Handle Strings
		case itemTypeRefString:
			if i, ok := item.(string); ok {
				it := itemType.(StringItem)
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

		// Handle Arrays
		case itemTypeRefArray:
			if i, ok := item.([]interface{}); ok {
				it := itemType.(ArrayItem)
				var iTypeErr int
				// Check inner item type
				for k := 0; k < len(i); k++ {
					i[k], iTypeErr = CheckQueryItemType(i[k], it.dataType.(*SchemaItem).iType)
					if iTypeErr != 0 {
						return nil, iTypeErr
					}
				}
				return i, 0
			}
			return nil, helpers.ErrorInvalidItemValue

		// Handle Objects
		case itemTypeRefObject:
			if i, ok := item.(map[string]interface{}); ok {
				it := itemType.(ObjectItem)
				newObj := make(map[string]interface{})
				for itemName, schemaItem := range *(it.schema) {
					insertItem := i[itemName]
					var filterErr int
					newObj[itemName], filterErr = SchemaFilter(schemaItem)
				}
				return newObj, 0
			}
			return nil, helpers.ErrorInvalidItemValue

		default:
			return nil, helpers.ErrorSchemaInvalidItemType
	}
}