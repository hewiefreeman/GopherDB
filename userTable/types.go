package userTable

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"reflect"
)

// Item data type initializers for table creation queries
var (
	itemTypeInitializor map[string][]reflect.Kind = map[string][]reflect.Kind {
		"Bool": []reflect.Kind{reflect.Bool},
		"Number": []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Bool},
		"String": []reflect.Kind{reflect.String, reflect.Float64, reflect.Bool, reflect.Bool},
		"Array": []reflect.Kind{reflect.Slice, reflect.Float64},
		"Object": []reflect.Kind{reflect.Map}}
)

// Item data type reflections
var (
	itemTypeRefBool = reflect.TypeOf(BoolItem{})
	itemTypeRefNumber = reflect.TypeOf(NumberItem{})
	itemTypeRefString = reflect.TypeOf(StringItem{})
	itemTypeRefArray = reflect.TypeOf(ArrayItem{})
	itemTypeRefObject = reflect.TypeOf(ObjectItem{})
)

type BoolItem struct {
	defaultValue bool
}

type NumberItem struct {
	defaultValue float64
	precision    uint8
	abs          bool
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
}

type ObjectItem struct {
	schema UserTableSchema
}

/////////////////////////////////////////////////////////////////////////////
//   CREATE   ///////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func createBool(params []interface{}) (UserTableSchemaItem, int){
	return UserTableSchemaItem{iType: BoolItem{defaultValue: params[0].(bool)}}, 0
}

func createNumber(params []interface{}) (UserTableSchemaItem, int){
	return UserTableSchemaItem{iType: NumberItem{defaultValue: params[0].(float64), precision: uint8(params[1].(float64)), abs: params[2].(bool)}}, 0
}

func createString(params []interface{}) (UserTableSchemaItem, int){
	return UserTableSchemaItem{iType: StringItem{defaultValue: params[0].(string), maxChars: uint32(params[1].(float64)), required: params[2].(bool), unique: params[3].(bool)}}, 0
}

func createArray(params []interface{}) (UserTableSchemaItem, int){
	schemaItem, iErr := makeSchemaItem(params[0].([]interface{}))
	if iErr != 0 {
		return UserTableSchemaItem{}, iErr
	}
	return UserTableSchemaItem{iType: ArrayItem{dataType: schemaItem, maxItems: uint32(params[1].(float64))}}, 0
}

func createObject(params []interface{}) (UserTableSchemaItem, int){
	if sObj, ok := params[0].(map[string]interface{}); ok {
		schema, schemaErr := NewSchema(sObj)
		if schemaErr != 0 {
			return UserTableSchemaItem{}, schemaErr
		}
		return UserTableSchemaItem{iType: ObjectItem{schema: schema}}, 0
	} else {
		return UserTableSchemaItem{}, helpers.ErrorSchemaInvalidItemParameters
	}
}