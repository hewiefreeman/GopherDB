package schema

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"strconv"
	"strings"
	"time"
)

// Method names
const (
	MethodOperatorAdd = "*add"
	MethodOperatorSub = "*sub"
	MethodOperatorMul = "*mul"
	MethodOperatorDiv = "*div"
	MethodOperatorMod = "*mod"
	MethodLength      = "*length" // TO-DO
	MethodEquals      = "*equals" // TO-DO
	MethodContains    = "*contains" // TO-DO
	MethodIndexOf     = "*indexOf" // TO-DO
	MethodAppend      = "*append"
	MethodAppendAt    = "*append["
	MethodAppendAtFin = "]"
	MethodPrepend     = "*prepend"
	MethodDelete      = "*delete"
	MethodSince       = "*since"
	MethodUntil       = "*until"
	MethodDay         = "*day"
	MethodHour        = "*hr"
	MethodMinute      = "*min"
	MethodSecond      = "*sec"
	MethodMillisecond = "*mil"
)

// Filter for queries
type Filter struct {
	restore bool
	get bool // when true, output is for get queries
	item interface{} // The item data to insert/get
	destination *interface{} // Pointer to where the filtered/retrieved data must go
	methods []string // Method list
	innerData []interface{} // Data hierarchy holder for entry on database (used for unique value search in insert/updates)
	schemaItems []*SchemaItem // Schema hierarchy holder (used for unique value search in insert/updates)
	uniqueVals *map[string]interface{} // Pointer to map storing unique values to check & set
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   QUERY FILTER   /////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ItemFilter filters an item in a query against it's cooresponding SchemaItem.
func ItemFilter(item interface{}, methods []string, destination *interface{}, innerData interface{}, schemaItem *SchemaItem, uniqueVals *map[string]interface{}, get bool, restore bool) int {
	filter := Filter{
		restore: restore,
		get: get,
		item: item,
		methods: methods,
		destination: destination,
		innerData: []interface{}{innerData},
		schemaItems: []*SchemaItem{schemaItem},
		uniqueVals: uniqueVals,
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
	if (filter.get && len(filter.methods) == 0) || (!filter.get && len(filter.schemaItems) == 1) {
		(*(*filter).destination) = filter.item
	}
	return 0
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   UNIQUE CHECKS   ////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
	}
	if filter.schemaItems[parentIndex].typeName == ItemTypeMap {
		// Check Map
		for _, item := range filter.innerData[parentIndex].(map[string]interface{}) {
			if getInnerUnique(filter, parentIndex+1, item) == filter.item {
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

func getInnerUnique(filter *Filter, indexOn int, item interface{}) interface{} {
	switch filter.schemaItems[indexOn].typeName {
		case ItemTypeString, ItemTypeInt8, ItemTypeInt16, ItemTypeInt32, ItemTypeInt64,
			ItemTypeUint8, ItemTypeUint16, ItemTypeUint32, ItemTypeUint64,
			ItemTypeFloat32, ItemTypeFloat64:
			return item

		case ItemTypeObject:
			// get item
			innerItem := item.(map[string]interface{})[filter.schemaItems[indexOn+1].name]
			return getInnerUnique(filter, (indexOn+1), innerItem)
	}
	return false
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ITEM TYPE METHODS   ////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetQueryItemMethods checks query item names for methods and returns the item name and the list of methods.
func GetQueryItemMethods(itemName string) (string, []string) {
	if strings.Contains(itemName, ".") {
		ml := strings.Split(itemName, ".")
		return ml[0], ml[1:]
	}
	return itemName, nil
}

func applyNumberMethods(numbs []interface{}, methods []string, dbEntryData interface{}) (float64, int) {
	var entryData float64
	if cNumb, ok := makeFloat(dbEntryData); ok {
		entryData = cNumb
	}
	// Must have same amount of numbers in array as methods to use on them
	if len(numbs) != len(methods) {
		return 0, helpers.ErrorInvalidMethodParameters
	}
	for i, numb := range numbs {
		// Check numb type
		if cNumb, ok := makeFloat(numb); ok {
			op := methods[i]
			switch op {
			case MethodOperatorAdd:
				entryData = entryData + cNumb

			case MethodOperatorSub:
				entryData = entryData - cNumb

			case MethodOperatorMul:
				entryData = entryData * cNumb

			case MethodOperatorDiv:
				entryData = entryData / cNumb

			case MethodOperatorMod:
				entryData = float64(int(entryData) % int(cNumb))

			default:
				return 0, helpers.ErrorInvalidMethod
			}
		} else {
			return 0, helpers.ErrorInvalidMethodParameters
		}
	}
	return entryData, 0
}

func applyStringMethods(strs []interface{}, methods []string, dbEntryData string) (string, int) {
	// Must have same amount of strings in array as methods to use on them
	if len(strs) != len(methods) {
		return "", helpers.ErrorInvalidMethodParameters
	}
	for i, str := range strs {
		// Check string type
		if cStr, ok := str.(string); ok {
			op := methods[i]
			switch op {
			case MethodOperatorAdd, MethodAppend:
				dbEntryData = dbEntryData + cStr
				continue

			case MethodPrepend:
				dbEntryData = cStr + dbEntryData
				continue
			}
			// Check for append at index method
			if len(methods[i]) >= 10 && methods[i][:8] == MethodAppendAt && methods[i][len(methods[i])-1:len(methods[i])] == MethodAppendAtFin {
				// Convert the text inside brackets to int
				j, jErr := strconv.Atoi(methods[i][8 : len(methods[i])-1])
				if jErr != nil {
					return "", helpers.ErrorInvalidMethod
				}
				// Prevent out of range error
				if j < 0 {
					j = 0
				} else if j > len(dbEntryData)-1 {
					j = len(dbEntryData) - 1
				}
				// Merge slices (could possibly be done better?) !!!
				entryStart := dbEntryData[:j]
				entryStart = entryStart + cStr
				dbEntryData = entryStart + dbEntryData[j:]
				continue
			}
			return "", helpers.ErrorInvalidMethod
		} else {
			return "", helpers.ErrorInvalidMethodParameters
		}
	}
	return dbEntryData, 0
}

func applyArrayMethods(filter *Filter) int {
	method := filter.methods[0]
	var dbEntryData []interface{}
	if filter.get {
		dbEntryData = filter.item.([]interface{})
	} else {
		dbEntryData = filter.innerData[len(filter.innerData)-1].([]interface{})
	}
	if item, ok := filter.item.([]interface{}); ok && !filter.get {
		// Basic array methods
		switch method {
		case MethodAppend:
			filter.methods = []string{}
			filter.innerData = append(filter.innerData, nil)
			filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem).dataType.(*SchemaItem))
			var index int
			for index, filter.item = range item {
				iTypeErr := queryItemFilter(filter)
				if iTypeErr != 0 {
					return iTypeErr
				}
				item[index] = filter.item
			}
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
			// Safely delete pointer in list for GC
			safeSchemaItemDelete(&filter.schemaItems)
			filter.item = append(dbEntryData, item...)
			return 0

		case MethodPrepend:
			filter.methods = []string{}
			filter.innerData = append(filter.innerData, nil)
			filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem).dataType.(*SchemaItem))
			var index int
			for index, filter.item = range item {
				iTypeErr := queryItemFilter(filter)
				if iTypeErr != 0 {
					return iTypeErr
				}
				item[index] = filter.item
			}
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
			safeSchemaItemDelete(&filter.schemaItems)
			filter.item = append(item, dbEntryData...)
			return 0

		case MethodDelete:
			// Item numbers to delete must be in order of greatest to least
			var lastNum int = len(dbEntryData)
			for _, numb := range item {
				if i, ok := makeInt(numb); ok {
					if i >= lastNum {
						return helpers.ErrorInvalidMethodParameters
					} else if i >= 0 {
						dbEntryData = append(dbEntryData[:i], dbEntryData[i+1:]...)
					}
					lastNum = i
				} else {
					return helpers.ErrorInvalidMethodParameters
				}
			}
			filter.methods = []string{}
			filter.item = dbEntryData
			return 0
		}

		// Check for append at index method
		if len(method) >= 10 && method[:8] == MethodAppendAt && method[len(method)-1:len(method)] == MethodAppendAtFin {
			// Convert the text inside brackets to int
			i, iErr := strconv.Atoi(method[8 : len(method)-1])
			if iErr != nil {
				return helpers.ErrorInvalidMethod
			}
			// Prevent out of range error
			if i < 0 {
				i = 0
			} else if i > len(dbEntryData)-1 {
				i = len(dbEntryData) - 1
			}
			filter.methods = []string{}
			filter.innerData = append(filter.innerData, nil)
			filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem).dataType.(*SchemaItem))
			var index int
			for index, filter.item = range item {
				iTypeErr := queryItemFilter(filter)
				if iTypeErr != 0 {
					return iTypeErr
				}
				item[index] = filter.item
			}
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
			safeSchemaItemDelete(&filter.schemaItems)
			// Merge slices (could possibly be done better?) !!!
			entryStart := append([]interface{}{}, dbEntryData[:i]...)
			entryStart = append(entryStart, item...)
			filter.item = append(entryStart, dbEntryData[i:]...)
			return 0
		}
	}

	// Try to convert methods[0] to int for index method
	i, iErr := strconv.Atoi(method)
	if iErr != nil {
		return helpers.ErrorInvalidMethod
	}
	// Prevent out of range error
	if len(dbEntryData) == 0 {
		return helpers.ErrorArrayEmpty
	} else if i < 0 {
		i = 0
	} else if i > len(dbEntryData)-1 {
		i = len(dbEntryData) - 1
	}
	// Check for more methods & filter
	filter.methods = filter.methods[1:]
	filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem).dataType.(*SchemaItem))
	if !filter.get {
		filter.innerData = append(filter.innerData, dbEntryData[i])
		iTypeErr := queryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		filter.innerData = filter.innerData[:len(filter.innerData)-1]
		safeSchemaItemDelete(&filter.schemaItems)
		dbEntryData[i] = filter.item
		filter.item = dbEntryData
	} else {
		filter.item = dbEntryData[i]
		iTypeErr := queryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		safeSchemaItemDelete(&filter.schemaItems)
	}
	return 0
}

func applyMapMethods(filter *Filter) int {
	method := filter.methods[0]
	var dbEntryData map[string]interface{}
	if filter.get {
		dbEntryData = filter.item.(map[string]interface{})
	} else {
		dbEntryData = filter.innerData[len(filter.innerData)-1].(map[string]interface{})
	}

	if !filter.get {
		if item, ok := filter.item.([]interface{}); ok && method == MethodDelete {
			// Delete method - eg: ["Mary", "Joe", "Vokome"]
			for _, n := range item {
				if itemName, ok := n.(string); ok {
					delete(dbEntryData, itemName)
				} else {
					return helpers.ErrorInvalidMethodParameters
				}
			}
			filter.methods = []string{}
			filter.item = dbEntryData
			return 0
		} else if item, ok := filter.item.(map[string]interface{}); ok && method == MethodAppend {
			// Append method - eg: {"x": 27, "y": 43}
			filter.methods = []string{}
			filter.innerData = append(filter.innerData, nil)
			filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(MapItem).dataType.(*SchemaItem))
			var itemName string
			for itemName, filter.item = range item {
				iTypeErr := queryItemFilter(filter)
				if iTypeErr != 0 {
					return iTypeErr
				}
				dbEntryData[itemName] = filter.item
			}
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
			safeSchemaItemDelete(&filter.schemaItems)
			filter.item = dbEntryData
			return 0
		}
	}

	// Checking for item with the name method[0] (Items with * not accepted)
	if !strings.Contains(method, "*") {
		filter.methods = filter.methods[1:]
		filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(MapItem).dataType.(*SchemaItem))
		if !filter.get {
			filter.innerData = append(filter.innerData, dbEntryData[method])
			iTypeErr := queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
			safeSchemaItemDelete(&filter.schemaItems)

			dbEntryData[method] = filter.item
			filter.item = dbEntryData
		} else {
			filter.item = dbEntryData[method]
			iTypeErr := queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			safeSchemaItemDelete(&filter.schemaItems)
		}
		return 0
	}
	return helpers.ErrorInvalidMethod
}

func applyObjectMethods(filter *Filter) int {
	method := filter.methods[0]
	var dbEntryData map[string]interface{}
	if filter.get {
		dbEntryData = filter.item.(map[string]interface{})
	} else {
		dbEntryData = filter.innerData[len(filter.innerData)-1].(map[string]interface{})
	}
	si := (*(filter.schemaItems[len(filter.schemaItems)-1].iType.(ObjectItem).schema))[method]
	if si == nil {
		return helpers.ErrorInvalidMethod
	}
	// Run method on item
	filter.methods = filter.methods[1:]
	filter.schemaItems = append(filter.schemaItems, si)
	if !filter.get {
		filter.innerData = append(filter.innerData, dbEntryData[method])
		iTypeErr := queryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		filter.innerData = filter.innerData[:len(filter.innerData)-1]
		safeSchemaItemDelete(&filter.schemaItems)
		dbEntryData[method] = filter.item
		filter.item = dbEntryData
	} else {
		filter.item = dbEntryData[method]
		iTypeErr := queryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		safeSchemaItemDelete(&filter.schemaItems)
	}
	return 0
}

func applyTimeMethods(filter *Filter, t time.Time) int {
	// Get method duration
	var d time.Duration
	switch filter.methods[0] {
	case MethodSince:
		d = time.Since(t)

	case MethodUntil:
		d = time.Until(t)

	default:
		return helpers.ErrorInvalidMethod
	}

	// Get method format
	format := MethodSecond
	if len(filter.methods) > 1 {
		format = filter.methods[1]
	}

	switch format {
	case MethodMillisecond:
		filter.item = d.Seconds() * 1000

	case MethodSecond:
		filter.item = d.Seconds()

	case MethodMinute:
		filter.item = d.Minutes()

	case MethodHour:
		filter.item = d.Hours()

	case MethodDay:
		filter.item = d.Hours() / 24

	default:
		return helpers.ErrorInvalidMethod
	}

	//
	filter.methods = []string{}
	return 0
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ITEM TYPE FILTERS   ////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func boolFilter(filter *Filter) int {
	if filter.get {
		return 0
	} else if i, ok := filter.item.(bool); ok {
		filter.item = i
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func int8Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic int8
	if i, ok := makeFloat(filter.item); ok {
		ic = int8(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = int8(mRes)
		filter.methods = []string{}
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func int16Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic int16
	if i, ok := makeFloat(filter.item); ok {
		ic = int16(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = int16(mRes)
		filter.methods = []string{}
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func int32Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic int32
	if i, ok := makeFloat(filter.item); ok {
		ic = int32(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = int32(mRes)
		filter.methods = []string{}
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func int64Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic int64
	if i, ok := makeFloat(filter.item); ok {
		ic = int64(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = int64(mRes)
		filter.methods = []string{}
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func uint8Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic uint8
	if i, ok := makeFloat(filter.item); ok {
		ic = uint8(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = uint8(mRes)
		filter.methods = []string{}
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func uint16Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic uint16
	if i, ok := makeFloat(filter.item); ok {
		ic = uint16(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = uint16(mRes)
		filter.methods = []string{}
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func uint32Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic uint32
	if i, ok := makeFloat(filter.item); ok {
		ic = uint32(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = uint32(mRes)
		filter.methods = []string{}
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func uint64Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic uint64
	if i, ok := makeFloat(filter.item); ok {
		ic = uint64(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = uint64(mRes)
		filter.methods = []string{}
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func float32Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic float32
	if i, ok := makeFloat(filter.item); ok {
		ic = float32(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = float32(mRes)
		filter.methods = []string{}
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func float64Filter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic float64
	if i, ok := makeFloat(filter.item); ok {
		ic = i
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1])
		if mErr != 0 {
			return mErr
		}
		ic = mRes
		filter.methods = []string{}
	} else {
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
		return helpers.ErrorUniqueValueInUse
	}
	return 0
}

func stringFilter(filter *Filter) int {
	if filter.get {
		return 0
	}
	var ic string
	if i, ok := filter.item.(string); ok {
		ic = i
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply string methods
		mRes, mErr := applyStringMethods(i, filter.methods, filter.innerData[len(filter.innerData)-1].(string))
		if mErr != 0 {
			return mErr
		}
		ic = mRes
		filter.methods = []string{}
	} else {
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
		return helpers.ErrorUniqueValueInUse
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
		filter.schemaItems = append(filter.schemaItems, it.dataType.(*SchemaItem))
		var index int
		for index, filter.item = range i {
			iTypeErr = queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			i[index] = filter.item
		}
		safeSchemaItemDelete(&filter.schemaItems)
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
		filter.schemaItems = append(filter.schemaItems, it.dataType.(*SchemaItem))
		var itemName string
		for itemName, filter.item = range i {
			iTypeErr = queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			i[itemName] = filter.item
		}
		safeSchemaItemDelete(&filter.schemaItems)
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
		filter.schemaItems = append(filter.schemaItems, &SchemaItem{})
		var itemName string
		for itemName, filter.schemaItems[len(filter.schemaItems)-1] = range *(it.schema) {
			filter.item = i[itemName]
			filterErr := queryItemFilter(filter)
			if filterErr != 0 {
				return filterErr
			}
			i[itemName] = filter.item
		}
		safeSchemaItemDelete(&filter.schemaItems)
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
			// Restoring from JSON file
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
//   ITEM TYPE CONVERTERS   /////////////////////////////////////////////////////////////////////////////////////
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

	case float64:
		return int(t), true
	}
	return 0, false
}

/////////////////// SAFE POINTER DELETE ////////////////////////////////////

func safeSchemaItemDelete(sil *[]*SchemaItem) {
	// Safely delete pointer in list for GC
	(*sil)[len(*sil)-1] = nil
	*sil = (*sil)[:len(*sil)-1]
}