package ggdb

import (
	"reflect"
)

func tableExists(n string) bool {
	tablesMux.Lock()
	t := tables[n]
	tablesMux.Unlock()
	return t != nil
}

func makeEntryMap(entry []interface{}, schema tableSchema, sel []string) map[string]interface{} {
	entryMap := make(map[string]interface{})
	selLen := len(sel)
	for k, v := range schema.items {
		if selLen > 0 {
			for i := 0; i < selLen; i++ {
				if sel[i] == k {
					entryMap[k] = entry[v]
				}
			}
		} else {
			entryMap[k] = entry[v]
		}
	}
	return entryMap
}

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
