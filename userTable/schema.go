package userTable

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"reflect"
	//"fmt"
)

type UserTableSchema map[string]UserTableSchemaItem

type UserTableSchemaItem struct {
	dataIndex int32
	iType interface{}
}

// NOTES:
//
//	Type Declarations:
//		- ["Bool", defaultValue] : store as Boolean
//			> defaultValue: default value of the Bool
//
//		- ["Number", defaultValue, decimalPrecision, absolute] : store as float64
//			> defaultValue: default value of the Number
//			> decimalPrecision: precision for Number decimal
//				Note: 0 stores as int64, anything higher is float64
//			> absolute: when true, the Number can only be positive
//				Note: if a negative value is assigned to an absolute Number, it will be set to 0
//
//		- ["String", defaultValue, maxChars, required, unique] : store as string
//			> defaultValue: default value the of String
//			> maxChars: maximum characters the String can be
//			> required: when true, the value cannot be set to blank
//			> unique: when true, no two database entries can be assigned the same value
//				Note: a unique String (or a unique String Object item) inside an Array checks the containing Array, and not other database entries
//
//		- ["Array", dataType, maxItems] : store as []interface{}
//			> dataType: the data type of the Array's items
//			> maxItems: the maximum amount of items in the Array
//
//		- ["Object", schema] : store as map[string]interface{}
//			> schema: the schema that the Object must adhere to
//				Note: same as making the schema for a UserTable
//				Note: if an Object's parent is the database, any unique Strings in the Object with be checked against the rest of the database. Use an Array of Object to make locally (to a User entry) unique Object lists
//
//	Example query to make a new UserTable:
//
//		{"NewUserTable": [
//			"users",
//			{
//				"email": ["String", "", 0, true, true],
//				"friends": ["Array", ["Object", {
//									"name": ["String", "", 0, true, true],
//									"status": ["Number", 0, 0, false]
//				}], 50],
//				"vCode": ["String", "", 0, true, false],
//				"verified": ["Bool", false]
//			}
//		]};
//

func NewSchema(schema map[string]interface{}) (UserTableSchema, int) {
	if len(schema) == 0 {
		return nil, helpers.ErrorSchemaItemsRequired
	}
	s := make(UserTableSchema)

	// Make Schema
	for itemName, itemParams := range schema {
		// Check item format
		if params, ok := itemParams.([]interface{}); ok {
			schemaItem, iErr := makeSchemaItem(params)
			if iErr != 0 {
				return nil, iErr
			}
			s[itemName] = schemaItem
		} else {
			// Invalid format
			return nil, helpers.ErrorSchemaInvalidFormat
		}
	}

	//
	return s, 0
}

func makeSchemaItem(params []interface{}) (UserTableSchemaItem, int) {
	if len(params) <= 1 {
		// Invalid format - requires at least a length of 2 for any item data type
		return UserTableSchemaItem{}, helpers.ErrorSchemaInvalidItemParameters
	}

	// Get data type
	if t, ok := params[0].(string); ok {
		dti := itemTypeInitializor[t]
		dtiPL := len(dti.paramTypes)
		if dtiPL == 0 {
			return UserTableSchemaItem{}, helpers.ErrorSchemaInvalidItemType
		} else if dtiPL != len(params[1:]) {
			return UserTableSchemaItem{}, helpers.ErrorSchemaInvalidItemParameters
		}
		// Check for valid parameter data types
		for i := 0; i < dtiPL; i++ {
			if reflect.TypeOf(params[i+1]).Kind() != dti.paramTypes[i] {
				return UserTableSchemaItem{}, helpers.ErrorSchemaInvalidItemParameters
			}
		}
		// Execute create for the type
		if t == "Bool" {
			return createBool(params[1:])
		} else if t == "Number" {
			return createNumber(params[1:])
		} else if t == "String" {
			return createString(params[1:])
		} else if t == "Array" {
			return createArray(params[1:])
		} else if t == "Object" {
			return createObject(params[1:])
		}
		return UserTableSchemaItem{}, helpers.ErrorUnexpected
	} else {
		return UserTableSchemaItem{}, helpers.ErrorSchemaInvalidFormat
	}
}

func validSchema(ts UserTableSchema) bool {
	if ts == nil || len(ts) == 0 {
		return false
	}
	for _, v := range ts {
		if !validSchemaItem(v){
			return false
		}
	}
	return true
}

func validSchemaItem(si UserTableSchemaItem) bool {
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