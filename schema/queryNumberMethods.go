package schema

import (
	"github.com/hewiefreeman/GopherDB/helpers"
)

// Makes temporary Int64 items for query methods which create ints
func tempInt64Method(filter *Filter, i int64) int {
	filter.schemaItems = append(filter.schemaItems, SchemaItem{typeName: ItemTypeInt64})
	filter.innerData = append(filter.innerData, i)
	if err := applyIntMethods(filter); err != 0 {
		return err
	}
	filter.innerData = filter.innerData[:len(filter.innerData) - 1]
	filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
	return 0
}

// Run methods on IntXX type item
func applyIntMethods(filter *Filter) int {
	var entryData int64
	var err int
	var brk bool
	entryData, _ = makeInt64(filter.innerData[len(filter.innerData)-1])
	if fList, ok := filter.item.([]interface{}); ok {
		for _, methodParam := range fList {
			// Check methodParam type
			if cNumb, ok := makeInt64(methodParam); ok {
				brk, err = getIntMethodResult(filter, &entryData, cNumb)
				if err != 0 {
					return err
				}
			} else {
				return helpers.ErrorInvalidMethodParameters
			}
			// Break when requested (when entrydata would no longer be a number type)
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
		if filter.get {
			// Convert int item back to OG int type
			filter.item, _ = makeTypeLiteral(entryData, &filter.schemaItems[len(filter.schemaItems) - 1])
		} else {
			filter.item = entryData
		}
	}
	return 0
}

func getIntMethodResult(filter *Filter, entryData *int64, num int64) (bool, int) {
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
				if err := checkGeneralIntMethods(method, entryData, num); err != 0 {
					return false, err
				}
		}
	} else {
		if err := checkGeneralIntMethods(method, entryData, num); err != 0 {
			return false, err
		}
	}
	return brk, 0
}

func checkGeneralIntMethods(method string, entryData *int64, num int64) int {
	switch method {
		case MethodOperatorAdd:
			*entryData += num

		case MethodOperatorSub:
			*entryData -= num

		case MethodOperatorMul:
			*entryData *= num

		case MethodOperatorDiv:
			*entryData /= num

		case MethodOperatorMod:
			*entryData %= num

		default:
			return helpers.ErrorInvalidMethod
	}
	return 0
}

// Run methods on UintXX type item
func applyUintMethods(filter *Filter) int {
	var entryData uint64
	var err int
	var brk bool
	entryData, _ = makeUint64(filter.innerData[len(filter.innerData)-1])
	if fList, ok := filter.item.([]interface{}); ok {
		for _, methodParam := range fList {
			// Check methodParam type
			if cNumb, ok := makeUint64(methodParam); ok {
				brk, err = getUintMethodResult(filter, &entryData, cNumb)
				if err != 0 {
					return err
				}
			} else {
				return helpers.ErrorInvalidMethodParameters
			}
			// Break when requested (when entrydata would no longer be a number type)
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
		if filter.get {
			// Convert uint item back to OG int type
			filter.item, _ = makeTypeLiteral(entryData, &filter.schemaItems[len(filter.schemaItems) - 1])
		} else {
			filter.item = entryData
		}
	}
	return 0
}

func getUintMethodResult(filter *Filter, entryData *uint64, num uint64) (bool, int) {
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
				if err := checkGeneralUintMethods(method, entryData, num); err != 0 {
					return false, err
				}
		}
	} else {
		if err := checkGeneralUintMethods(method, entryData, num); err != 0 {
			return false, err
		}
	}
	return brk, 0
}

func checkGeneralUintMethods(method string, entryData *uint64, num uint64) int {
	switch method {
		case MethodOperatorAdd:
			*entryData += num

		case MethodOperatorSub:
			*entryData -= num

		case MethodOperatorMul:
			*entryData *= num

		case MethodOperatorDiv:
			*entryData /= num

		case MethodOperatorMod:
			*entryData %= num

		default:
			return helpers.ErrorInvalidMethod
	}
	return 0
}

// Run methods on FloatXX type item
func applyFloatMethods(filter *Filter) int {
	var entryData float64
	var err int
	var brk bool
	entryData, _ = makeFloat64(filter.innerData[len(filter.innerData)-1])
	if fList, ok := filter.item.([]interface{}); ok {
		for _, methodParam := range fList {
			// Check methodParam type
			if cNumb, ok := makeFloat64(methodParam); ok {
				brk, err = getFloatMethodResult(filter, &entryData, cNumb)
				if err != 0 {
					return err
				}
			} else {
				return helpers.ErrorInvalidMethodParameters
			}
			// Break when requested (when entrydata would no longer be a number type)
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
		if filter.get {
			// Convert float item back to OG int type
			filter.item, _ = makeTypeLiteral(entryData, &filter.schemaItems[len(filter.schemaItems) - 1])
		} else {
			filter.item = entryData
		}
	}
	return 0
}

func getFloatMethodResult(filter *Filter, entryData *float64, num float64) (bool, int) {
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
				if err := checkGeneralFloatMethods(method, entryData, num); err != 0 {
					return false, err
				}
		}
	} else {
		if err := checkGeneralFloatMethods(method, entryData, num); err != 0 {
			return false, err
		}
	}
	return brk, 0
}

func checkGeneralFloatMethods(method string, entryData *float64, num float64) int {
	switch method {
		case MethodOperatorAdd:
			*entryData += num

		case MethodOperatorSub:
			*entryData -= num

		case MethodOperatorMul:
			*entryData *= num

		case MethodOperatorDiv:
			*entryData /= num

		case MethodOperatorMod:
			*entryData = float64(int(*entryData + 0.5) % int(num + 0.5))

		default:
			return helpers.ErrorInvalidMethod
	}
	return 0
}