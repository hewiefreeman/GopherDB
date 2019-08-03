package schema

import (
	"github.com/hewiefreeman/GopherDB/schema"
	"testing"
	"reflect"
)

var (
	itemTypeInitializor map[string][]reflect.Kind = map[string][]reflect.Kind{
		schema.ItemTypeBool:    []reflect.Kind{reflect.Bool},
		schema.ItemTypeInt8:    []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		schema.ItemTypeInt16:   []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		schema.ItemTypeInt32:   []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		schema.ItemTypeInt64:   []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		schema.ItemTypeUint8:   []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		schema.ItemTypeUint16:  []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		schema.ItemTypeUint32:  []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		schema.ItemTypeUint64:  []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool},
		schema.ItemTypeFloat32: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		schema.ItemTypeFloat64: []reflect.Kind{reflect.Float64, reflect.Float64, reflect.Float64, reflect.Bool, reflect.Bool, reflect.Bool},
		schema.ItemTypeString:  []reflect.Kind{reflect.String, reflect.Float64, reflect.Bool, reflect.Bool},
		schema.ItemTypeArray:   []reflect.Kind{reflect.Slice, reflect.Float64, reflect.Bool},
		schema.ItemTypeMap:     []reflect.Kind{reflect.Slice, reflect.Float64, reflect.Bool},
		schema.ItemTypeObject:  []reflect.Kind{reflect.Map, reflect.Bool},
		schema.ItemTypeTime:    []reflect.Kind{reflect.String, reflect.Bool},
	}
)

func BenchmarkItemTypeInit2(b *testing.B) {
	b.ReportAllocs()
	params := []interface{}{"Uint8", float64(5), float64(0), float64(0), false, false}
	for i := 0; i < b.N; i++ {
		if !checkTypeFormat("Uint8")(params[1:]) {
			b.Errorf("Incorrect Item Parameters")
		}
	}
}

func BenchmarkItemTypeInit1(b *testing.B) {
	b.ReportAllocs()
	params := []interface{}{"Uint8", float64(5), float64(0), float64(0), false, false}
	for i := 0; i < b.N; i++ {
		dti := itemTypeInitializor["Uint8"]
		// Check for valid params length
		dtiPL := len(dti)
		if dtiPL == 0 {
			b.Errorf("Empty init")
			return
		} else if dtiPL != len(params)-1 {
			b.Errorf("Init doesnt match params")
			return
		}
		// Check for valid parameter data types
		for j := 0; j < dtiPL; j++ {
			if params[j+1] == nil {
				b.Errorf("Nil unacceptable")
				return
			}
			if reflect.TypeOf(params[j+1]).Kind() != dti[j] {
				b.Errorf("Invalid parameter type")
				return
			}
		}
	}
}

func checkTypeFormat(t string) func([]interface{})bool {
	switch t {
	case schema.ItemTypeUint8, schema.ItemTypeUint16,
			schema.ItemTypeUint32, schema.ItemTypeUint64:
		return checkNumericFormat
	case schema.ItemTypeInt8, schema.ItemTypeInt16,
			schema.ItemTypeInt32, schema.ItemTypeInt64,
			schema.ItemTypeFloat32, schema.ItemTypeFloat64:
		return checkNumericPlusFormat
	case schema.ItemTypeString:
		return checkStringFormat
	case schema.ItemTypeArray, schema.ItemTypeMap:
		return checkListFormat
	case schema.ItemTypeObject:
		return checkObjectFormat
	case schema.ItemTypeTime:
		return checkTimeFormat
	default:
		return nil
	}
}

func checkNumericFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 5 {
		return false
	}
	if _, ok := f[0].(float64); !ok {
		return false
	}
	if _, ok := f[1].(float64); !ok {
		return false
	}
	if _, ok := f[2].(float64); !ok {
		return false
	}
	if _, ok := f[3].(bool); !ok {
		return false
	}
	if _, ok := f[4].(bool); !ok {
		return false
	}
	return true
}

func checkNumericPlusFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 6 {
		return false
	}
	if _, ok := f[0].(float64); !ok {
		return false
	}
	if _, ok := f[1].(float64); !ok {
		return false
	}
	if _, ok := f[2].(float64); !ok {
		return false
	}
	if _, ok := f[3].(bool); !ok {
		return false
	}
	if _, ok := f[4].(bool); !ok {
		return false
	}
	if _, ok := f[5].(bool); !ok {
		return false
	}
	return true
}

func checkStringFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 4 {
		return false
	}
	if _, ok := f[0].(string); !ok {
		return false
	}
	if _, ok := f[1].(float64); !ok {
		return false
	}
	if _, ok := f[2].(bool); !ok {
		return false
	}
	if _, ok := f[3].(bool); !ok {
		return false
	}
	return true
}

func checkListFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 4 {
		return false
	}
	if _, ok := f[0].([]interface{}); !ok {
		return false
	}
	if _, ok := f[1].(float64); !ok {
		return false
	}
	if _, ok := f[2].(bool); !ok {
		return false
	}
	return true
}

func checkObjectFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 4 {
		return false
	}
	if _, ok := f[0].(map[string]interface{}); !ok {
		return false
	}
	if _, ok := f[1].(bool); !ok {
		return false
	}
	return true
}

func checkTimeFormat(f []interface{}) bool {
	fLen := len(f)
	if fLen != 4 {
		return false
	}
	if _, ok := f[0].(string); !ok {
		return false
	}
	if _, ok := f[1].(bool); !ok {
		return false
	}
	return true
}