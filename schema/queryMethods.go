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
	MethodEquals      = "*eq" // Testing...
	MethodGreater     = "*gt" // Testing...
	MethodLess        = "*lt" // Testing...
	MethodGreaterOE   = "*gte" // Testing...
	MethodLessOE      = "*lte" // Testing...
	MethodContains    = "*contains" // Testing strings... TO-DO: arrays, maps
	MethodIndexOf     = "*indexOf" // Testing strings... TO-DO: arrays, maps
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

// Run methods on number type item
func applyNumberMethods(filter *Filter) int {
	var entryData float64
	var err int
	var brk bool
	entryData, _ = makeFloat64(filter.innerData[len(filter.innerData)-1])
	if fList, ok := filter.item.([]interface{}); ok {
		for _, methodParam := range fList {
			// Check methodParam type
			if cNumb, ok := makeFloat64(methodParam); ok {
				brk, err = getNumberMethodResult(filter, &entryData, cNumb)
				if err != 0 {
					return err
				}
			} else {
				return helpers.ErrorInvalidMethodParameters
			}
			// Break when requested
			if brk {
				break
			}
			// Remove this method
			filter.methods = filter.methods[1:]
		}
	} else {
		return helpers.ErrorInvalidMethodParameters
	}
	filter.methods = []string{}
	if !brk {
		filter.item = entryData
	}
	return 0
}

func getNumberMethodResult(filter *Filter, entryData *float64, num float64) (bool, int) {
	if len(filter.methods) == 0 {
		return false, helpers.ErrorTooManyMethodParameters
	}
	method := filter.methods[0]
	var brk bool
	if filter.get {
		switch method {
			case MethodEquals:
				filter.item = (*entryData == num)
				brk = true

			case MethodGreater:
				filter.item = (*entryData > num)
				brk = true

			case MethodGreaterOE:
				filter.item = (*entryData >= num)
				brk = true

			case MethodLess:
				filter.item = (*entryData < num)
				brk = true

			case MethodLessOE:
				filter.item = (*entryData <= num)
				brk = true

			default:
				if err := checkGeneralNumberMethods(method, entryData, num); err != 0 {
					return false, err
				}
		}
	} else {
		if err := checkGeneralNumberMethods(method, entryData, num); err != 0 {
			return false, err
		}
	}
	return brk, 0
}

func checkGeneralNumberMethods(method string, entryData *float64, num float64) int {
	switch method {
		case MethodOperatorAdd:
			*entryData = *entryData + num

		case MethodOperatorSub:
			*entryData = *entryData - num

		case MethodOperatorMul:
			*entryData = *entryData * num

		case MethodOperatorDiv:
			*entryData = *entryData / num

		case MethodOperatorMod:
			*entryData = float64(int(*entryData) % int(num))

		default:
			return helpers.ErrorInvalidMethod
	}
	return 0
}

func applyStringMethods(filter *Filter) int {
	var entryData string
	var err int
	var brk bool
	var typeName string
	entryData, _ = filter.innerData[len(filter.innerData)-1].(string)
	if fList, ok := filter.item.([]interface{}); ok {
		for _, methodParam := range fList {
			// Check methodParam type
			if cString, ok := methodParam.(string); ok {
				brk, typeName, err = getStringMethodResult(filter, &entryData, cString)
				if err != 0 {
					return err
				}
			} else {
				brk, typeName, err = getStringMethodResult(filter, &entryData, "")
				if err != 0 {
					return err
				}
			}
			filter.methods = filter.methods[1:]
			if brk {
				break
			}
		}
	} else {
		return helpers.ErrorInvalidMethodParameters
	}
	if brk {
		if len(filter.methods) > 0 {
			// More methods to run...
			switch typeName {
				case ItemTypeFloat64:
					if err = applyNumberMethods(filter); err != 0 {
						return err
					}
			}
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
		} else {
			filter.item = filter.innerData[len(filter.innerData)-1]
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
		}
	} else {
		filter.methods = []string{}
		filter.item = entryData
	}
	return 0
}

func getStringMethodResult(filter *Filter, entryData *string, str string) (bool, string, int) {
	if len(filter.methods) == 0 {
		return false, "", helpers.ErrorTooManyMethodParameters
	}
	method := filter.methods[0]
	var brk bool
	var typeName string
	if filter.get {
		switch method {
			case MethodLength:
				filter.innerData = append(filter.innerData, float64(len(*entryData)))
				typeName = ItemTypeFloat64
				brk = true

			case MethodIndexOf:
				var indexOf float64 = -1
				for i := 0; i < len(*entryData) - (len(str) - 1); i++ {
					if (*entryData)[i:i + len(str)] == str {
						indexOf = float64(i)
						break
					}
				}
				filter.item = filter.item.([]interface{})[1:]
				filter.innerData = append(filter.innerData, indexOf)
				typeName = ItemTypeFloat64
				brk = true

			case MethodContains:
				var contains bool
				for i := 0; i < len(*entryData) - (len(str) - 1); i++ {
					if (*entryData)[i:i + len(str)] == str {
						contains = true
						break
					}
				}
				filter.innerData = append(filter.innerData, contains)
				typeName = ItemTypeBool
				brk = true

			case MethodEquals:
				filter.innerData = append(filter.innerData, (*entryData == str))
				typeName = ItemTypeBool
				brk = true

			default:
				if err := checkGeneralStringMethods(filter, method, entryData, str); err != 0 {
					return false, "", err
				}
		}
	} else {
		if err := checkGeneralStringMethods(filter, method, entryData, str); err != 0 {
			return false, "", err
		}
	}
	return brk, typeName, 0
}

func checkGeneralStringMethods(filter *Filter, method string, entryData *string, str string) int {
	switch method {
		case MethodOperatorAdd, MethodAppend:
			*entryData = *entryData + str

		case MethodPrepend:
			*entryData = str + *entryData

		default:
			// Check for append at index method
			if len(method) >= 10 && method[:8] == MethodAppendAt && method[len(method)-1:len(method)] == MethodAppendAtFin {
				// Convert the text inside brackets to int
				j, jErr := strconv.Atoi(method[8 : len(method)-1])
				if jErr != nil {
					return helpers.ErrorInvalidMethod
				}
				// Prevent out of range error
				if j < 0 {
					j = 0
				} else if j > len(*entryData) - 1 {
					j = len(*entryData) - 1
				}
				// Merge slices (could possibly be done better?) !!!
				entryStart := (*entryData)[:j]
				entryStart = entryStart + str
				*entryData = entryStart + (*entryData)[j:]
			} else {
				return helpers.ErrorInvalidMethod
			}
	}
	filter.item = filter.item.([]interface{})[1:]
	return 0
}

// Run methods on Array item collection
func applyArrayMethods(filter *Filter) int {
	method := filter.methods[0]
	dbEntryData := filter.innerData[len(filter.innerData)-1].([]interface{})
	if filter.get {
		switch method {
		case MethodLength:
			filter.methods = filter.methods[1:]
			if len(filter.methods) > 0 {
				// run methods as float64
				filter.innerData = append(filter.innerData, float64(len(dbEntryData)))
				if err := applyNumberMethods(filter); err != 0 {
					return err
				}
				filter.innerData = filter.innerData[:len(filter.innerData) - 1]
			} else {
				filter.item = len(dbEntryData)
			}

		case MethodIndexOf:
			// TO-DO
			filter.item = -1

		case MethodContains:
			// TO-DO
			filter.item = false
		}
		return 0
	} else {
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
	dbEntryData := filter.innerData[len(filter.innerData)-1].(map[string]interface{})
	if filter.get {
		switch method {
		case MethodLength:
			filter.methods = filter.methods[1:]
			if len(filter.methods) > 0 {
				// run methods as float64
				filter.innerData = append(filter.innerData, float64(len(dbEntryData)))
				if err := applyNumberMethods(filter); err != 0 {
					return err
				}
				filter.innerData = filter.innerData[:len(filter.innerData) - 1]
			} else {
				filter.item = len(dbEntryData)
			}

		case MethodIndexOf:
			// TO-DO
			filter.item = -1

		case MethodContains:
			// TO-DO
			filter.item = false
		}
		return 0
	} else {
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
	dbEntryData := filter.innerData[len(filter.innerData)-1].(map[string]interface{})
	filter.schemaItems = append(filter.schemaItems, si)
	filter.innerData = append(filter.innerData, dbEntryData[method])
	iTypeErr := queryItemFilter(filter)
	if iTypeErr != 0 {
		return iTypeErr
	}
	filter.innerData = filter.innerData[:len(filter.innerData)-1]
	filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
	if !filter.get {
		dbEntryData[method] = filter.item
		filter.item = dbEntryData
	}
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