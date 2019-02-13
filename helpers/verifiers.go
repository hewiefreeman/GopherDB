package helpers

import (
	"reflect"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   interface{} verification   /////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func isHashable(v interface{}) bool {
	t := reflect.TypeOf(v).Kind()
	return (t < reflect.Complex64 && t > 0) || t == reflect.String
}

func isArray(v interface{}) bool {
	t := reflect.TypeOf(v).Kind()
	return t == reflect.Array || t == reflect.Slice
}

func isMap(v interface{}) bool {
	t := reflect.TypeOf(v).Kind()
	return t == reflect.Map
}
