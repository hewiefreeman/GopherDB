package helpers

import (
	"reflect"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   interface{} verification   /////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func IsHashable(v interface{}) bool {
	t := reflect.TypeOf(v).Kind()
	return (t < reflect.Complex64 && t > 0) || t == reflect.String
}

func IsArray(v interface{}) bool {
	t := reflect.TypeOf(v).Kind()
	return t == reflect.Array || t == reflect.Slice
}

func IsMap(v interface{}) bool {
	t := reflect.TypeOf(v).Kind()
	return t == reflect.Map
}
