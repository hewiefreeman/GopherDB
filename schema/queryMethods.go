package schema

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"time"
	"strconv"
	"strings"
)

// Method names
const (
	MethodOperatorAdd = "*add"
	MethodOperatorSub = "*sub"
	MethodOperatorMul = "*mul"
	MethodOperatorDiv = "*div"
	MethodOperatorMod = "*mod"
	MethodLength      = "*len"
	MethodEquals      = "*eq" // TO-DO
	MethodGreater     = "*gt" // TO-DO
	MethodLess        = "*lt" // TO-DO
	MethodGreaterOE   = "*gte" // TO-DO
	MethodLessOE      = "*lte" // TO-DO
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

// GetQueryItemMethods checks query item names for methods and returns the item name and the list of methods.
func GetQueryItemMethods(itemName string) (string, []string) {
	if strings.Contains(itemName, ".") {
		ml := strings.Split(itemName, ".")
		return ml[0], ml[1:]
	}
	return itemName, nil
}

// Run methods on numer type item
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

// Run methods on String item
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

// Run get methods on String item
func applyStringGetMethod(method string, dbEntryData string, filter *Filter) int {
	switch method {
	case MethodLength:
		filter.item = len(dbEntryData)
	default:
		return helpers.ErrorInvalidMethod
	}
	return 0
}

// Run methods on Array item collection
func applyArrayMethods(filter *Filter) int {
	method := filter.methods[0]
	var dbEntryData []interface{}
	if filter.get {
		dbEntryData = filter.item.([]interface{})
		switch method {
		case MethodLength:
			filter.methods = []string{}
			filter.item = len(dbEntryData)
			return 0
		}
	} else {
		dbEntryData = filter.innerData[len(filter.innerData)-1].([]interface{})
		if item, ok := filter.item.([]interface{}); ok {
			// Basic array methods
			switch method {
			case MethodAppend:
				if err := filterArrayMethodItems(filter, &item); err != 0 {
					return err
				}
				filter.item = append(dbEntryData, item...)
				return 0
			case MethodPrepend:
				if err := filterArrayMethodItems(filter, &item); err != 0 {
					return err
				}
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
				if err := filterArrayMethodItems(filter, &item); err != 0 {
					return err
				}
				// Merge slices (could possibly be done better?) !!!
				entryStart := append([]interface{}{}, dbEntryData[:i]...)
				entryStart = append(entryStart, item...)
				filter.item = append(entryStart, dbEntryData[i:]...)
				return 0
			}
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
	filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem).dataType)
	if !filter.get {
		filter.innerData = append(filter.innerData, dbEntryData[i])
		iTypeErr := queryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		filter.innerData = filter.innerData[:len(filter.innerData)-1]
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
		dbEntryData[i] = filter.item
		filter.item = dbEntryData
	} else {
		filter.item = dbEntryData[i]
		iTypeErr := queryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
	}
	return 0
}

// Run filter on Array method item collection
func filterArrayMethodItems(filter *Filter, item *[]interface{}) int {
	filter.methods = []string{}
	filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem).dataType)
	var index int
	for index, filter.item = range *item {
		iTypeErr := queryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		(*item)[index] = filter.item
		filter.innerData[len(filter.innerData)-1] = append(filter.innerData[len(filter.innerData)-1].([]interface{}), filter.item)
	}
	filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
	return 0
}

// Run methods on Map item collection
func applyMapMethods(filter *Filter) int {
	method := filter.methods[0]
	var dbEntryData map[string]interface{}
	if filter.get {
		dbEntryData = filter.item.(map[string]interface{})
		switch method {
		case MethodLength:
			filter.methods = []string{}
			filter.item = len(dbEntryData)
			return 0
		}
	} else {
		dbEntryData = filter.innerData[len(filter.innerData)-1].(map[string]interface{})
		if item, ok := filter.item.([]interface{}); ok && method == MethodDelete {
			// Delete method - eg: ["Mary", "Joe", "Vokome"]
			for _, n := range item {
				if itemName, ok := n.(string); ok {
					delete(dbEntryData, itemName)
					continue
				}
				return helpers.ErrorInvalidMethodParameters
			}
			filter.methods = []string{}
			filter.item = dbEntryData
			return 0
		} else if item, ok := filter.item.(map[string]interface{}); ok && method == MethodAppend {
			// Append method - eg: {"x": 27, "y": 43}
			filter.methods = []string{}
			//filter.innerData = append(filter.innerData, nil)
			filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(MapItem).dataType)
			var itemName string
			for itemName, filter.item = range item {
				if iTypeErr := queryItemFilter(filter); iTypeErr != 0 {
					return iTypeErr
				}
				dbEntryData[itemName] = filter.item
				filter.innerData[len(filter.innerData)-1].(map[string]interface{})[itemName] = filter.item
			}
			//filter.innerData = filter.innerData[:len(filter.innerData)-1]
			filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
			filter.item = dbEntryData
			return 0
		}
	}

	// Checking for item with the name method[0] (Items with * not accepted)
	if !strings.Contains(method, "*") {
		filter.methods = filter.methods[1:]
		filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(MapItem).dataType)

		if !filter.get {
			filter.innerData = append(filter.innerData, dbEntryData[method])
			if iTypeErr := queryItemFilter(filter); iTypeErr != 0 {
				return iTypeErr
			}
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
			filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
			dbEntryData[method] = filter.item
			filter.item = dbEntryData
			return 0
		}

		filter.item = dbEntryData[method]
		if iTypeErr := queryItemFilter(filter); iTypeErr != 0 {
			return iTypeErr
		}
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
		return 0
	}
	return helpers.ErrorInvalidMethod
}

// Run methods on Object item
func applyObjectMethods(filter *Filter) int {
	method := filter.methods[0]
	// Check if valid object item
	si := (filter.schemaItems[len(filter.schemaItems)-1].iType.(ObjectItem).schema)[method]
	if !si.Validate() {
		return helpers.ErrorInvalidMethod
	}
	// Remove this method and add new schemaItem
	filter.methods = filter.methods[1:]
	filter.schemaItems = append(filter.schemaItems, si)
	// Run method on item
	if filter.get {
		switch method {
		case MethodLength:
			filter.item = len(filter.item.(map[string]interface{}))
			return 0
		}
		filter.item = filter.item.(map[string]interface{})[method]
		iTypeErr := queryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
		return 0
	}
	dbEntryData := filter.innerData[len(filter.innerData)-1].(map[string]interface{})
	filter.innerData = append(filter.innerData, dbEntryData[method])
	iTypeErr := queryItemFilter(filter)
	if iTypeErr != 0 {
		return iTypeErr
	}
	filter.innerData = filter.innerData[:len(filter.innerData)-1]
	filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
	dbEntryData[method] = filter.item
	filter.item = dbEntryData
	return 0
}

// Run methods on Time item
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