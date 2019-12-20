package schema

import (

)

// GetUniqueItems gets the item name (string, eg: ["login", "email"]) of all unique table items and appends them to destination.
func GetUniqueItems(schema Schema, destination *[]string, outerItems string) {
	// Loop through schema & find unique value names
	for itemName, schemaItem := range schema {
		if schemaItem.typeName == ItemTypeObject {
			// Top-level Objects can hold items unique to the table
			if outerItems != "" {
				outerItems = outerItems + "." + itemName
			} else {
				outerItems = itemName
			}
			GetUniqueItems(schemaItem.iType.(ObjectItem).schema, destination, outerItems)
		} else if schemaItem.Unique() {
			if outerItems != "" {
				outerItems = outerItems + "." + itemName
			} else {
				outerItems = itemName
			}
			*destination = append(*destination, outerItems)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Query checks   /////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Check if current item is either a top-level table item to search whole database after filter,
// or nested object to search locally within entry now.
func uniqueCheck(filter *Filter) bool {
	// Get parent index
	parentIndex := -1
	for i := len(filter.schemaItems)-1; i >= 0; i-- {
		if filter.schemaItems[i].typeName == ItemTypeArray || filter.schemaItems[i].typeName == ItemTypeMap {
			parentIndex = i
			break
		}
	}
	if parentIndex == -1 {
		// No valid parent, get name for table check
		name := filter.schemaItems[0].name
		for i := 1; i < len(filter.schemaItems)-1; i++ {
				name = name+"."+filter.schemaItems[i].name
		}
		// Add to uniqueVals to be checked after filter
		(*(filter.uniqueVals))[name] = filter.item

		return false
	} else if filter.innerData[parentIndex] == nil {
		return false
	}

	if filter.schemaItems[parentIndex].typeName == ItemTypeMap {
		// Check Map
		for _, item := range filter.innerData[parentIndex].(map[string]interface{}) {
			if getInnerUnique(filter, (parentIndex+1), item) == filter.item {
				return true
			}
		}
	} else {
		// Check Array
		for _, item := range filter.innerData[parentIndex].([]interface{}) {
			if getInnerUnique(filter, (parentIndex+1), item) == filter.item {
				return true
			}
		}
	}
	//
	return false
}

// Get nested entry items for unique check
func getInnerUnique(filter *Filter, indexOn int, item interface{}) interface{} {
	switch filter.schemaItems[indexOn].typeName {
		case ItemTypeString:
			return item
		case ItemTypeInt8, ItemTypeInt16, ItemTypeInt32, ItemTypeInt64,
			ItemTypeUint8, ItemTypeUint16, ItemTypeUint32, ItemTypeUint64,
			ItemTypeFloat32, ItemTypeFloat64:
			filter.item, _ = makeFloat64(filter.item)
			item, _ := makeFloat64(item)
			return item

		case ItemTypeObject:
			// get item
			innerItem := item.(map[string]interface{})[filter.schemaItems[indexOn+1].name]
			return getInnerUnique(filter, (indexOn+1), innerItem)
	}
	return false
}