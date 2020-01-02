package schema

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"time"
	"strings"
)

var (

)

// Sorting Arrays for query filters
func sort(filter *Filter, ary []interface{}, by interface{}, asc bool) int {
	itemType := filter.schemaItems[len(filter.schemaItems) - 1].iType.(ArrayItem).dataType
	if itemType.IsNumeric() {
		sortArrayNumeric(ary, &itemType, asc)
		return 0
	}
	switch itemType.typeName {
	case ItemTypeString:
		sortArrayString(ary, asc)
	case ItemTypeTime:
		sortArrayTime(ary, &itemType, asc)
	case ItemTypeObject: // TO-DO
		// Convert "by" to string array
		var byArr []string
		if s, ok := by.(string); ok {
			if s == "" {
				return helpers.ErrorArrayItemNotSortable
			}
			byArr = strings.Split(s, ".")
		} else {
			return helpers.ErrorInvalidMethodParameters
		}
		if len(byArr) == 0 {
			return helpers.ErrorInvalidMethodParameters
		}
		if err := sortArrayByObjectItem(ary, &itemType, byArr, asc); err != 0 {
			return err
		}
	default:
		return helpers.ErrorArrayItemNotSortable
	}
	return 0
}

// Sort numeric Arrays
func sortArrayNumeric(ary []interface{}, itemType *SchemaItem, asc bool) {
	// Convert numeric type to float64
	var fArr []float64 = make([]float64, len(ary), len(ary))
	var tf float64
	var ti interface{}
	for i, v := range ary {
		fArr[i], _ = makeFloat64(v)
	}
	// Sort as float64
	for i := 0; i < len(fArr) - 1; i++ {
		for j := len(fArr) - 1; j > i; j-- {
			if (asc && fArr[i] > fArr[j]) || (!asc && fArr[i] < fArr[j]) {
				tf = fArr[i]
				fArr[i] = fArr[j]
				fArr[j] = tf
				ti = ary[i]
				ary[i] = ary[j]
				ary[j] = ti
			}
		}
	}
}

// Sort string Arrays
func sortArrayString(ary []interface{}, asc bool) {
	var iItem string
	var jItem string
	for i := 0; i < len(ary) - 1; i++ {
		iItem, _ = ary[i].(string)
		for j := len(ary) - 1; j > i; j-- {
			jItem, _ = ary[j].(string)
			if (asc && iItem > jItem) || (!asc && iItem < jItem) {
				ary[i] = jItem
				ary[j] = iItem
				iItem = jItem
			}
		}
	}
}

// Sort time Arrays
func sortArrayTime(ary []interface{}, itemType *SchemaItem, asc bool) {
	var iItem time.Time
	var jItem time.Time
	var t interface{}
	for i := 0; i < len(ary) - 1; i++ {
		iItem, _ = makeTime(ary[i], itemType)
		for j := len(ary) - 1; j > i; j-- {
			jItem, _ = makeTime(ary[j], itemType)
			if (asc && iItem.After(jItem)) || (!asc && iItem.Before(jItem)) {
				t = ary[i]
				ary[i] = ary[j]
				ary[j] = t
				iItem = jItem
			}
		}
	}
}

// Sort Array by inner Object item
func sortArrayByObjectItem(ary []interface{}, itemType *SchemaItem, byArr []string, asc bool) int {
	// Check by item
	var innerSi SchemaItem
	var dataIndexes []int = make([]int, len(byArr), len(byArr))
	var err int
	if innerSi, err = checkSortByItem(itemType.iType.(ObjectItem).schema, byArr, dataIndexes, 0); err != 0 {
		return err
	}
	// Create check array for inner object item values
	checkAry := make([]interface{}, len(ary), len(ary))
	for i, v := range ary {
		checkAry[i] = getSortByValue(v.([]interface{}), dataIndexes, 0)
	}
	// Sort by inner type
	if innerSi.IsNumeric() {
		// Sort as float64
		var iItem float64
		var jItem float64
		var t interface{}
		for i := 0; i < len(checkAry) - 1; i++ {
			iItem, _ = makeFloat64(checkAry[i])
			for j := len(checkAry) - 1; j > i; j-- {
				jItem, _ = makeFloat64(checkAry[j])
				if (asc && iItem > jItem) || (!asc && iItem < jItem) {
					// Swap both ary and checkAry
					t = ary[i]
					ary[i] = ary[j]
					ary[j] = t
					checkAry[i] = jItem
					checkAry[j] = iItem
					iItem = jItem
				}
			}
		}
		return 0
	}
	switch innerSi.typeName {
	case ItemTypeString:
		// Sort as string
		var iItem string
		var jItem string
		var t interface{}
		for i := 0; i < len(checkAry) - 1; i++ {
			iItem, _ = checkAry[i].(string)
			for j := len(checkAry) - 1; j > i; j-- {
				jItem, _ = checkAry[j].(string) // could change with inner loop
				if (asc && iItem > jItem) || (!asc && iItem < jItem) {
					// Swap both ary and checkAry
					t = ary[i]
					ary[i] = ary[j]
					ary[j] = t
					checkAry[i] = jItem
					checkAry[j] = iItem
					iItem = jItem
				}
			}
		}
		return 0
	case ItemTypeTime:
		// Sort as Time
		var iItem time.Time
		var jItem time.Time
		var t interface{}
		for i := 0; i < len(checkAry) - 1; i++ {
			iItem, _ = makeTime(checkAry[i], &innerSi)
			for j := len(checkAry) - 1; j > i; j-- {
				jItem, _ = makeTime(checkAry[j], &innerSi)
				if (asc && iItem.After(jItem)) || (!asc && iItem.Before(jItem)) {
					// Swap both ary and checkAry
					t = ary[i]
					ary[i] = ary[j]
					ary[j] = t
					checkAry[i] = jItem
					checkAry[j] = iItem
					iItem = jItem
				}
			}
		}
	}
	return 0
}

// Check validity of a sort-by query parameter
func checkSortByItem(schema Schema, byArr []string, dataIndexes []int, byOn int) (SchemaItem, int) {
	si := schema[byArr[byOn]]
	if si.QuickValidate() {
		if !si.IsNumeric() {
			switch si.typeName {
			case ItemTypeTime, ItemTypeString:
				// valid item
			case ItemTypeObject:
				if len(byArr) == byOn + 1 {
					return SchemaItem{}, helpers.ErrorArrayItemNotSortable
				}
				dataIndexes[byOn] = int(schema[byArr[byOn]].dataIndex)
				return checkSortByItem(si.iType.(ObjectItem).schema, byArr, dataIndexes, byOn + 1)

			default:
				return SchemaItem{}, helpers.ErrorArrayItemNotSortable
			}
		}
		if len(byArr) != byOn + 1 {
			return SchemaItem{}, helpers.ErrorInvalidMethodParameters
		}
		dataIndexes[byOn] = int(schema[byArr[byOn]].dataIndex)
		return si, 0
	}
	return SchemaItem{}, helpers.ErrorInvalidMethodParameters
}

// Get the value of inner Objects for a sort-by query
func getSortByValue(i []interface{}, dataIndexes []int, iOn int) interface{} {
	if iOn < len(dataIndexes) - 1 {
		return getSortByValue(i[dataIndexes[iOn]].([]interface{}), dataIndexes, iOn + 1)
	} else {
		r := i[dataIndexes[iOn]]
		return r
	}
}