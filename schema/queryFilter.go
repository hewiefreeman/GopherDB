package schema

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"time"
	//"fmt"
)

// Filter for queries
type Filter struct {
	restore bool
	get bool // when true, output is for get queries
	item interface{} // The item data to insert/get
	destination *interface{} // Pointer to where the filtered/retrieved data must go
	methods []string // Method list
	innerData []interface{} // Data hierarchy holder for entry on database (used for unique value search in insert/updates)
	schemaItems []SchemaItem // Schema hierarchy holder (used for unique value search in insert/updates)
	uniqueVals *map[string]interface{} // Pointer to map storing unique values to check & set
}

// ItemFilter filters an item in a query against it's cooresponding SchemaItem.
func ItemFilter(item interface{}, methods []string, destination *interface{}, innerData interface{}, schemaItem SchemaItem, uniqueVals *map[string]interface{}, get bool, restore bool) int {
	filter := Filter{
		restore: restore,
		get: get,
		item: item,
		methods: methods,
		destination: destination,
		innerData: []interface{}{},
		schemaItems: []SchemaItem{schemaItem},
		uniqueVals: uniqueVals,
	}
	if innerData != nil {
		filter.innerData = []interface{}{innerData}
	}
	return queryItemFilter(&filter)
}

// queryItemFilter takes in an item from a query, and filters/checks it for format/completion against the cooresponding SchemaItem data type.
func queryItemFilter(filter *Filter) int {
	if filter.item == nil {
		// No methods allowed on a nil item
		if len(filter.methods) > 0 {
			return helpers.ErrorInvalidMethodParameters
		}
		// Get default value
		defaultVal, defaultErr := defaultVal(filter.schemaItems[len(filter.schemaItems)-1])
		if defaultErr != 0 {
			return defaultErr
		}
		if len(filter.schemaItems) == 1 {
			(*(*filter).destination) = defaultVal
		} else {
			filter.item = defaultVal
		}
		return 0
	}

	// Run type filter
	iTypeErr := getTypeFilter(filter.schemaItems[len(filter.schemaItems)-1].typeName)(filter)
	if iTypeErr != 0 {
		return iTypeErr
	}
	// Check if this is the last filter itteration, and apply item to destination
	if (filter.get && len(filter.methods) == 0) || (!filter.get && len(filter.schemaItems) == 1) {
		(*(*filter).destination) = filter.item
	}
	return 0
}

func getTypeFilter(typeName string) func(*Filter)(int) {
	switch typeName {
	case ItemTypeBool: return boolFilter
	case ItemTypeInt8: return int8Filter
	case ItemTypeInt16: return int16Filter
	case ItemTypeInt32: return int32Filter
	case ItemTypeInt64: return int64Filter
	case ItemTypeUint8: return uint8Filter
	case ItemTypeUint16: return uint16Filter
	case ItemTypeUint32: return uint32Filter
	case ItemTypeUint64: return uint64Filter
	case ItemTypeFloat32: return float32Filter
	case ItemTypeFloat64: return float64Filter
	case ItemTypeString: return stringFilter
	case ItemTypeArray: return arrayFilter
	case ItemTypeMap: return mapFilter
	case ItemTypeObject: return objectFilter
	case ItemTypeTime: return timeFilter
	default: return nil
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Item type filters   ////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func boolFilter(filter *Filter) int {
	if filter.get {
		filter.methods = []string{}
		filter.item = filter.innerData[len(filter.innerData)-1]
		return 0
	} else if i, ok := filter.item.(bool); ok {
		filter.item = i
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func int8Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic int8
	if i, ok := makeInt(filter.item); ok {
		ic = int8(i)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Int8Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic * (-1)
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func int16Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic int16
	if i, ok := makeFloat(filter.item); ok {
		ic = int16(i)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Int16Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic * (-1)
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func int32Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic int32
	if i, ok := makeFloat(filter.item); ok {
		ic = int32(i)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Int32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic * (-1)
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func int64Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic int64
	if i, ok := makeFloat(filter.item); ok {
		ic = int64(i)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Int64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic * (-1)
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func uint8Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic uint8
	if i, ok := makeFloat(filter.item); ok {
		ic = uint8(i)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Uint8Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func uint16Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic uint16
	if i, ok := makeFloat(filter.item); ok {
		ic = uint16(i)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Uint16Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func uint32Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic uint32
	if i, ok := makeFloat(filter.item); ok {
		ic = uint32(i)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Uint32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func uint64Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic uint64
	if i, ok := makeFloat(filter.item); ok {
		ic = uint64(i)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Uint64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func float32Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic float32
	if i, ok := makeFloat(filter.item); ok {
		ic = float32(i)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Float32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic * (-1)
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func float64Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyNumberMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic float64
	var ok bool
	if ic, ok = makeFloat(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Float64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic * (-1)
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func stringFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply string methods
		mErr := applyStringMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if filter.get {
		return 0
	}
	var ic string
	var ok bool
	if ic, ok = filter.item.(string); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(StringItem)
	l := uint32(len(ic))
	// Check length and if required
	if it.maxChars > 0 && l > it.maxChars {
		return helpers.ErrorStringTooLarge
	} else if it.required && l == 0 {
		return helpers.ErrorStringRequired
	}
	// Check if unique
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func arrayFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		mErr := applyArrayMethods(filter)
		if mErr != 0 {
			return mErr
		}
		return 0
	} else if filter.get {
		return 0
	} else if i, ok := filter.item.([]interface{}); ok {
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem)
		var iTypeErr int
		// Check inner item type
		if (len(filter.schemaItems) == 1 && len(filter.innerData) == 0) || len(filter.schemaItems) > 1 {
			filter.innerData = append(filter.innerData, make([]interface{}, 0))
		}
		filter.schemaItems = append(filter.schemaItems, it.dataType)
		var index int
		for index, filter.item = range i {
			iTypeErr = queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			i[index] = filter.item
			filter.innerData[len(filter.innerData)-1] = append(filter.innerData[len(filter.innerData)-1].([]interface{}), filter.item)
		}
		filter.innerData = filter.innerData[:len(filter.innerData)-1]
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
		if it.required && len(i) == 0 {
			return helpers.ErrorArrayItemsRequired
		}
		filter.item = i
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func mapFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		mErr := applyMapMethods(filter)
		if mErr != 0 {
			return mErr
		}
		return 0
	} else if filter.get {
		return 0
	} else if i, ok := filter.item.(map[string]interface{}); ok {
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(MapItem)
		var iTypeErr int
		// Check inner item type
		if (len(filter.schemaItems) == 1 && len(filter.innerData) == 0) || len(filter.schemaItems) > 1 {
			filter.innerData = append(filter.innerData, make(map[string]interface{}))
		}
		filter.schemaItems = append(filter.schemaItems, it.dataType)
		var itemName string
		for itemName, filter.item = range i {
			iTypeErr = queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			i[itemName] = filter.item
			filter.innerData[len(filter.innerData)-1].(map[string]interface{})[itemName] = filter.item
		}
		filter.innerData = filter.innerData[:len(filter.innerData)-1]
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
		if it.required && len(i) == 0 {
			return helpers.ErrorMapItemsRequired
		}
		filter.item = i
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func objectFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		mErr := applyObjectMethods(filter)
		if mErr != 0 {
			return mErr
		}
		return 0
	} else if filter.get {
		return 0
	} else if i, ok := filter.item.(map[string]interface{}); ok {
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(ObjectItem)
		if (len(filter.schemaItems) == 1 && len(filter.innerData) == 0) || len(filter.schemaItems) > 1 {
			filter.innerData = append(filter.innerData, make(map[string]interface{}))
		}
		filter.schemaItems = append(filter.schemaItems, SchemaItem{})
		var itemName string
		for itemName, filter.schemaItems[len(filter.schemaItems)-1] = range it.schema {
			filter.item = i[itemName]
			filterErr := queryItemFilter(filter)
			if filterErr != 0 {
				return filterErr
			}
			i[itemName] = filter.item
			filter.innerData[len(filter.innerData)-1].(map[string]interface{})[itemName] = filter.item
		}
		filter.innerData = filter.innerData[:len(filter.innerData)-1]
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
		filter.item = i
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func timeFilter(filter *Filter) int {
	if filter.get {
		var t time.Time
		// If the item is a string, was retrieved from disk - convert to time.Time
		if i, ok := filter.item.(string); ok {
			var tErr error
			t, tErr = time.Parse(TimeFormatRFC3339, i) // JSON uses RFC3339
			if tErr != nil {
				return helpers.ErrorInvalidTimeFormat
			}
		} else {
			t = filter.item.(time.Time)
		}
		if len(filter.methods) > 0 {
			if mErr := applyTimeMethods(filter, t); mErr != 0 {
				return mErr
			}
			return 0
		}
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(TimeItem)
		filter.item = t.Format(it.format)
		return 0
	} else if i, ok := filter.item.(string); ok {
		if len(filter.methods) > 0 {
			return helpers.ErrorInvalidMethod
		}
		if i == "*now" {
			// Set to current database time
			filter.item = time.Now()
			return 0
		}
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(TimeItem)
		var t time.Time
		var err error
		if filter.restore {
			// Restoring from JSON file using RFC3339
			t, err = time.Parse(TimeFormatRFC3339, i)
		} else {
			t, err = time.Parse(it.format, i)
		}
		if err != nil {
			return helpers.ErrorInvalidTimeFormat
		}
		filter.item = t
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Data type converters   /////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func makeFloat(i interface{}) (float64, bool) {
	switch t := i.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case float32:
		return float64(t), true
	}
	return 0, false
}

func makeInt(i interface{}) (int, bool) {
	switch t := i.(type) {
	case float64:
		return int(t), true
	case int:
		return t, true
	case int8:
		return int(t), true
	case int16:
		return int(t), true
	case int32:
		return int(t), true
	case int64:
		return int(t), true
	case uint8:
		return int(t), true
	case uint16:
		return int(t), true
	case uint32:
		return int(t), true
	case uint64:
		return int(t), true
	case float32:
		return int(t), true
	}
	return 0, false
}