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
	"strings"
)

// Schema represents a database schema that one or more tables must adhere to.
type Schema map[string]SchemaItem

// SchemaItem represents one of the items in a Schema. SchemaItem also holds info about the data type and it's settings.
type SchemaItem struct {
	dataIndex uint32
	name      string
	typeName  string
	iType     interface{}
	rawParams []interface{}
}

type SchemaConfigItem struct {
	Position uint32
	Name     string
	DataType []interface{}
}

// NOTES:
//
//	Type Declarations:
//		- ["Bool", defaultValue] : store as Boolean
//			> defaultValue: default value of the Bool
//
//		    *INT_TYPE*: "Int8" | "Int16" | "Int32" | "Int64" | "Uint8" | "Uint16" | "Uint32" | "Uint64"
//		- ["*INT_TYPE*", defaultValue, min, max, absolute, required, unique] : store as int(8-64) | uint(8-64)
//			> defaultValue: default value of the integer
//			> min: minimum value
//			> max: maximum value
//			> absolute: when true, the vale will always be a positive or 0 value (specifying a negative number will store it as positive)
//			> required: when true, the value must be specified when inserting (does not check on updates)
//			> unique: when true, no two database entries can be assigned the same value (automatically sets required to true)
//				Note: a unique value (or a unique value Object item) inside an Array/Map checks the containing Array/Map, and not the whole database
//
//		    *FLOAT_TYPE*: "Float32" | "Float64"
//		- ["*FLOAT_TYPE*", defaultValue, min, max, absolute, required] : store as float32 | float64
//			> defaultValue: default value of the integer
//			> min: minimum value
//			> max: maximum value
//			> absolute: when true, the vale will always be a positive or 0 value (specifying a negative number will store it as positive)
//			> required: when true, the value must be specified when inserting (does not check on updates)
//			> unique: when true, no two database entries can be assigned the same value (automatically sets required to true)
//				Note: a unique value (or a unique value Object item) inside an Array/Map checks the containing Array/Map, and not the whole database
//
//		- ["String", defaultValue, maxChars, required, unique] : store as string
//			> defaultValue: default value the of String
//			> maxChars: maximum characters the String can be
//			> required: when true, the value cannot be set to a blank string. When inserting, the value must be specified unless there is a valid default value
//			> unique: when true, no two database entries can be assigned the same value (automatically sets required to true)
//				Note: a unique value (or a unique value Object item) inside an Array/Map checks the containing Array/Map, and not the whole database
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
//				Note: same as making the schema for a AuthTable
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
//										"status": ["Uint8", 0, 0, 2, false, false] // defaultValue, min, max, absolute, required
//								}, false],
//						50, false],
//			"vCode": ["String", "", 0, true, false],
//			"verified": ["Bool", false]
//		}
//

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Creating a Schema   ////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// New creates a new schema from a JSON schema object
func New(schema interface{}, restore bool) (Schema, int) {
	s := make(Schema)
	var i uint32
	if !restore {
		if schemaInput, ok := schema.(map[string]interface{}); ok {
			// Standard query input - assign data indexes
			for itemName, itemParams := range schemaInput {
				// Names cannot have "." or "*"
				if len(itemName) == 0 || strings.ContainsAny(itemName, ".*\n\t\r") {
					return nil, helpers.ErrorSchemaInvalidItemName
				}
				// Check item format
				if params, ok := itemParams.([]interface{}); ok {
					schemaItem, iErr := makeSchemaItem(itemName, params, restore)
					if iErr != 0 {
						return nil, iErr
					}
					schemaItem.dataIndex = i
					//schemaItem.rawParams = params
					s[itemName] = schemaItem
					i++
				} else {
					return nil, helpers.ErrorSchemaInvalidFormat
				}
			}
		} else {
			return nil, helpers.ErrorSchemaInvalidFormat
		}
	} else if schemaInput, ok := schema.([]interface{}); ok {
		// use []interface{} like []schemaConfigItem
		posTrack := make([]bool, len(schemaInput))
		for _, v := range schemaInput {
			if si, ok := v.(map[string]interface{}); ok {
				var siName string
				var ok bool
				// Get Name and check format
				if siName, ok = si["Name"].(string); !ok {
					return nil, helpers.ErrorSchemaInvalidItemParameters
				}
				if len(siName) == 0 || strings.ContainsAny(siName, ".*\n\t\r") {
					return nil, helpers.ErrorSchemaInvalidItemName
				}
				// Get Position and verify integrity
				var siPos uint32
				if siPos, ok = makeUint32(si["Position"]); !ok {
					return nil, helpers.ErrorSchemaInvalidItemParameters
				}
				if int(siPos) >= len(posTrack) || siPos < 0 || posTrack[int(siPos)] {
					return nil, helpers.ErrorSchemaInvalidItemPosition
				}
				// Get DataType
				var siDataType []interface{}
				if siDataType, ok = si["DataType"].([]interface{}); !ok {
					return nil, helpers.ErrorSchemaInvalidItemParameters
				}
				// Make SchemaItem from data
				var err int
				var schemaItem SchemaItem
				if schemaItem, err = makeSchemaItem(siName, siDataType, restore); err != 0 {
					return nil, err
				}
				schemaItem.dataIndex = siPos
				//schemaItem.rawParams = siDataType
				s[siName] = schemaItem
				posTrack[int(siPos)] = true
			} else {
				return nil, helpers.ErrorSchemaInvalidFormat
			}
		}
	} else {
		return nil, helpers.ErrorSchemaInvalidFormat
	}
	//
	return s, 0
}

func makeSchemaItem(name string, params []interface{}, restore bool) (SchemaItem, int) {
	if len(params) <= 1 {
		// Invalid format - requires at least a length of 2 for any item data type
		return SchemaItem{}, helpers.ErrorSchemaInvalidItemParameters
	}

	// Get data type
	if t, ok := params[0].(string); ok {
		if !checkTypeFormat(t)(params[1:]) {
			return SchemaItem{}, helpers.ErrorSchemaInvalidItemParameters
		}
		// Execute create for the type
		si := SchemaItem{name: name, typeName: t, rawParams: params}
		switch t {
		case ItemTypeBool:
			si.iType = BoolItem{defaultValue: params[1].(bool)}

			return si, 0

		case ItemTypeInt8:
			si.iType = Int8Item{defaultValue: int8(params[1].(float64)), min: int8(params[2].(float64)), max: int8(params[3].(float64)), abs: params[4].(bool), required: params[5].(bool), unique: params[6].(bool)}
			return si, 0

		case ItemTypeInt16:
			si.iType = Int16Item{defaultValue: int16(params[1].(float64)), min: int16(params[2].(float64)), max: int16(params[3].(float64)), abs: params[4].(bool), required: params[5].(bool), unique: params[6].(bool)}
			return si, 0

		case ItemTypeInt32:
			si.iType = Int32Item{defaultValue: int32(params[1].(float64)), min: int32(params[2].(float64)), max: int32(params[3].(float64)), abs: params[4].(bool), required: params[5].(bool), unique: params[6].(bool)}
			return si, 0

		case ItemTypeInt64:
			si.iType = Int64Item{defaultValue: int64(params[1].(float64)), min: int64(params[2].(float64)), max: int64(params[3].(float64)), abs: params[4].(bool), required: params[5].(bool), unique: params[6].(bool)}
			return si, 0

		case ItemTypeUint8:
			si.iType = Uint8Item{defaultValue: uint8(params[1].(float64)), min: uint8(params[2].(float64)), max: uint8(params[3].(float64)), required: params[4].(bool), unique: params[5].(bool)}
			return si, 0

		case ItemTypeUint16:
			si.iType = Uint16Item{defaultValue: uint16(params[1].(float64)), min: uint16(params[2].(float64)), max: uint16(params[3].(float64)), required: params[4].(bool), unique: params[5].(bool)}
			return si, 0

		case ItemTypeUint32:
			si.iType = Uint32Item{defaultValue: uint32(params[1].(float64)), min: uint32(params[2].(float64)), max: uint32(params[3].(float64)), required: params[4].(bool), unique: params[5].(bool)}
			return si, 0

		case ItemTypeUint64:
			si.iType = Uint64Item{defaultValue: uint64(params[1].(float64)), min: uint64(params[2].(float64)), max: uint64(params[3].(float64)), required: params[4].(bool), unique: params[5].(bool)}
			return si, 0

		case ItemTypeFloat32:
			si.iType = Float32Item{defaultValue: float32(params[1].(float64)), min: float32(params[2].(float64)), max: float32(params[3].(float64)), abs: params[4].(bool), required: params[5].(bool), unique: params[6].(bool)}
			return si, 0

		case ItemTypeFloat64:
			si.iType = Float64Item{defaultValue: params[1].(float64), min: params[2].(float64), max: params[3].(float64), abs: params[4].(bool), required: params[5].(bool), unique: params[6].(bool)}
			return si, 0

		case ItemTypeString:
			si.iType = StringItem{defaultValue: params[1].(string), maxChars: uint32(params[2].(float64)), required: params[3].(bool), unique: params[4].(bool)}
			return si, 0

		case ItemTypeArray:
			schemaItem, iErr := makeSchemaItem(name, params[1].([]interface{}), restore)
			if iErr != 0 {
				return SchemaItem{}, iErr
			}
			si.iType = ArrayItem{dataType: schemaItem, maxItems: uint32(params[2].(float64))}
			return si, 0

		case ItemTypeMap:
			schemaItem, iErr := makeSchemaItem(name, params[1].([]interface{}), restore)
			if iErr != 0 {
				return SchemaItem{}, iErr
			}
			si.iType = MapItem{dataType: schemaItem, maxItems: uint32(params[2].(float64))}
			return si, 0

		case ItemTypeObject:
			// Creating new schema from query...
			schema, schemaErr := New(params[1], restore)
			if schemaErr != 0 {
				return SchemaItem{}, schemaErr
			}
			si.iType = ObjectItem{schema: schema}
			return si, 0

		case ItemTypeTime:
			var format string = timeFormatInitializor[params[1].(string)]
			if format == "" {
				return SchemaItem{}, helpers.ErrorSchemaInvalidTimeFormat
			}
			si.iType = TimeItem{format: format, required: params[2].(bool)}
			return si, 0

		default:
			return SchemaItem{}, helpers.ErrorUnexpected
		}
	}
	return SchemaItem{}, helpers.ErrorSchemaInvalidFormat
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Restoring a Schema   ///////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Restore restores a schema from a config file with it's Schema array
func Restore(schema []SchemaConfigItem) (Schema, int) {
	s := make(Schema)
	var iErr int
	var schemaItem SchemaItem
	posTrack := make([]bool, len(schema))
	for _, schemaConfItem := range schema {
		if int(schemaConfItem.Position) >= len(posTrack) || schemaConfItem.Position < 0 || posTrack[int(schemaConfItem.Position)] {
			return nil, helpers.ErrorSchemaInvalidItemPosition
		} else if len(schemaConfItem.Name) == 0 || strings.ContainsAny(schemaConfItem.Name, ".*\t\n\r") {
			return nil, helpers.ErrorSchemaInvalidItemName
		}
		schemaItem, iErr = makeSchemaItem(schemaConfItem.Name, schemaConfItem.DataType, true)
		if iErr != 0 {
			return nil, iErr
		}
		schemaItem.dataIndex = schemaConfItem.Position
		//schemaItem.rawParams = schemaConfItem.DataType
		s[schemaConfItem.Name] = schemaItem
		posTrack[int(schemaConfItem.Position)] = true
	}
	return s, 0
}

// MakeConfig makes a Schema for a config file
func (s Schema) MakeConfig() []SchemaConfigItem {
	var sc []SchemaConfigItem = make([]SchemaConfigItem, len(s))
	i := 0
	for _, v := range s {
		sc[i] = SchemaConfigItem{
			Position: v.dataIndex,
			Name:     v.name,
			DataType: v.makeConfigDataType(),
		}
		i++
	}
	return sc
}

func (si SchemaItem) makeConfigDataType() []interface{} {
	switch si.typeName {
	case ItemTypeObject:
		si.rawParams[1] = si.iType.(ObjectItem).schema.MakeConfig()

	case ItemTypeArray:
		// Check inner item type
		si.rawParams[1] = si.iType.(ArrayItem).dataType.makeConfigDataType()

	case ItemTypeMap:
		// Check inner item type
		si.rawParams[1] = si.iType.(MapItem).dataType.makeConfigDataType()
	}
	return si.rawParams
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Schema checks   ////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Validate returns true if Schema is valid
func (s Schema) Validate() bool {
	if len(s) == 0 {
		return false
	}
	for _, v := range s {
		if !v.Validate() {
			return false
		}
	}
	return true
}

// Validate returns true if SchemaItem is valid
func (si SchemaItem) Validate() bool {
	if si.name == "" || si.typeName == "" || si.dataIndex < 0 {
		return false
	}

	switch reflect.TypeOf(si.iType) {
	case itemTypeRefBool, itemTypeRefInt8, itemTypeRefInt16, itemTypeRefInt32,
		itemTypeRefInt64, itemTypeRefUint8, itemTypeRefUint16, itemTypeRefUint32,
		itemTypeRefUint64, itemTypeRefFloat32, itemTypeRefFloat64, itemTypeRefString,
		itemTypeRefArray, itemTypeRefMap, itemTypeRefObject, itemTypeRefTime:
		return true
	}

	return false
}

// QuickValidate returns true if SchemaItem is "valid"
func (si SchemaItem) QuickValidate() bool {
	return si.name != ""
}

// DataIndex gets the SchemaItem data index (table specific).
func (si SchemaItem) DataIndex() uint32 {
	return si.dataIndex
}

// TypeName gets the type name of the SchemaItem.
func (si SchemaItem) TypeName() string {
	return si.typeName
}

// Unique returns true if the SchemaItem is unique.
func (si SchemaItem) Unique() bool {
	switch si.typeName {
	case ItemTypeInt8:
		return si.iType.(Int8Item).unique
	case ItemTypeInt16:
		return si.iType.(Int16Item).unique
	case ItemTypeInt32:
		return si.iType.(Int32Item).unique
	case ItemTypeInt64:
		return si.iType.(Int64Item).unique
	case ItemTypeUint8:
		return si.iType.(Uint8Item).unique
	case ItemTypeUint16:
		return si.iType.(Uint16Item).unique
	case ItemTypeUint32:
		return si.iType.(Uint32Item).unique
	case ItemTypeUint64:
		return si.iType.(Uint64Item).unique
	case ItemTypeFloat32:
		return si.iType.(Float32Item).unique
	case ItemTypeFloat64:
		return si.iType.(Float64Item).unique
	case ItemTypeString:
		return si.iType.(StringItem).unique
	}
	return false
}

func (si SchemaItem) IsNumeric() bool {
	switch si.typeName {
	case ItemTypeInt8, ItemTypeInt16, ItemTypeInt32, ItemTypeInt64,
		ItemTypeUint8, ItemTypeUint16, ItemTypeUint32, ItemTypeUint64,
		ItemTypeFloat64, ItemTypeFloat32:
		return true
	default:
		return false
	}
}

func (si SchemaItem) IsFloat() bool {
	switch si.typeName {
	case ItemTypeFloat64, ItemTypeFloat32:
		return true
	default:
		return false
	}
}

// Unique returns true if the SchemaItem is unique.
func (si SchemaItem) Required() bool {
	switch si.typeName {
	case ItemTypeInt8:
		return si.iType.(Int8Item).required
	case ItemTypeInt16:
		return si.iType.(Int16Item).required
	case ItemTypeInt32:
		return si.iType.(Int32Item).required
	case ItemTypeInt64:
		return si.iType.(Int64Item).required
	case ItemTypeUint8:
		return si.iType.(Uint8Item).required
	case ItemTypeUint16:
		return si.iType.(Uint16Item).required
	case ItemTypeUint32:
		return si.iType.(Uint32Item).required
	case ItemTypeUint64:
		return si.iType.(Uint64Item).required
	case ItemTypeFloat32:
		return si.iType.(Float32Item).required
	case ItemTypeFloat64:
		return si.iType.(Float64Item).required
	case ItemTypeString:
		return si.iType.(StringItem).required
	case ItemTypeArray:
		return si.iType.(ArrayItem).required
	case ItemTypeObject:
		return false
	case ItemTypeMap:
		return si.iType.(MapItem).required
	case ItemTypeTime:
		return si.iType.(TimeItem).required
	}
	return false
}
