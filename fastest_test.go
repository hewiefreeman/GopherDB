package main

import (
	"testing"
	//"sync"
	"strconv"
	"encoding/json"
)

/////////////////////////////////////////////////////////////////////////////////////////
//  MAP ALLOCATION BENCHMARKS  //////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

/*
func BenchmarkMapAlloc1(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		allocTest1()
	}
}

func BenchmarkMapAlloc2(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		allocTest2()
	}
}

func BenchmarkMapAlloc3(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		allocTest3()
	}
}

func allocTest1() map[string]interface{} {
	m := make(map[string]interface{})
	m["g"] = make(map[string]interface{})
	m["g"].(map[string]interface{})["g"] = true
	m["g"].(map[string]interface{})["h"] = true
	m["g"].(map[string]interface{})["i"] = true
	return m
}

func allocTest2() map[string]map[string]interface{} {
	m := make(map[string]map[string]interface{})
	m["g"] = make(map[string]interface{})
	m["g"]["g"] = true
	m["g"]["h"] = true
	m["g"]["i"] = true
	return m
}

// Winner by ~20ns
func allocTest3() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"g": map[string]interface{}{
			"g": true,
			"h": true,
			"i": true,
		},
	}
}
*/

/////////////////////////////////////////////////////////////////////////////////////////
//  MAP VS SWITCH BENCHMARKS  ///////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

/*
const (
	Option1 string = "option1"
	Option2 string = "option2"
	Option3 string = "option3"
	Option4 string = "option4"
	Option5 string = "option5"
	Option6 string = "option6"
	Option7 string = "option7"
	Option8 string = "option8"
	Option9 string = "option9"
	Option10 string = "option10"
	Option11 string = "option11"
	Option12 string = "option12"
	Option13 string = "option13"
	Option14 string = "option14"
	Option15 string = "option15"
	Option16 string = "option16"
	Option17 string = "option17"
	Option18 string = "option18"
	Option19 string = "option19"
	Option20 string = "option20"
)

var sMap map[string]func()(bool) = map[string]func()(bool){
	Option1: option1,
	Option2: option2,
	Option3: option3,
	Option4: option4,
	Option5: option5,
	Option6: option4,
	Option7: option3,
	Option8: option2,
	Option9: option1,
	Option10: option2,
	Option11: option3,
	Option12: option4,
	Option13: option5,
	Option14: option4,
	Option15: option3,
	Option16: option2,
	Option17: option1,
	Option18: option2,
	Option19: option3,
	Option20: option4,
}

// WINNER by ~320ns
func BenchmarkSwitch(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 20; i++ {
			switchTest("option"+strconv.Itoa(i+1))
		}
	}
}

func BenchmarkMap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 20; i++ {
			mapTest("option"+strconv.Itoa(i+1))
		}
	}
}

func switchTest(o string) bool {
	switch o {
	case Option1:
		return option1()
	case Option2:
		return option2()
	case Option3:
		return option3()
	case Option4:
		return option4()
	case Option5:
		return option5()
	case Option6:
		return option4()
	case Option7:
		return option3()
	case Option8:
		return option2()
	case Option9:
		return option1()
	case Option10:
		return option2()
	case Option11:
		return option3()
	case Option12:
		return option4()
	case Option13:
		return option5()
	case Option14:
		return option4()
	case Option15:
		return option3()
	case Option16:
		return option2()
	case Option17:
		return option1()
	case Option18:
		return option2()
	case Option19:
		return option3()
	case Option20:
		return option4()
	}
	return false
}

func mapTest(o string) bool {
	return sMap[o]()
}

func option1() bool {
	return true
}

func option2() bool {
	return true
}

func option3() bool {
	return true
}

func option4() bool {
	return true
}

func option5() bool {
	return true
}
*/

/////////////////////////////////////////////////////////////////////////////////////////
//  Mutex Lock Order  ///////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

/*
var (
	mux sync.Mutex
	get []interface{} = []interface{}{
		"hello",
		25.12,
		[]int{1, 2, 3},
	}
	getP *[]interface{} = &get
)

func BenchmarkLockOnce(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mux.Lock()
			get[0] = get[0].(string)+"wtf"
			get[1] = get[1].(float64)+13.37
			get[2] = append(get[2].([]int), []int{4, 5, 6}...)
			mux.Unlock()
		}
	})
}

// Takes twice as long (as expected) and +1 alloc/op
func BenchmarkLockTwice(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mux.Lock()
			g := append([]interface{}{}, *getP...)
			mux.Unlock()
			g[0] = g[0].(string)+"wtf"
			g[1] = g[1].(float64)+13.37
			g[2] = append(g[2].([]int), []int{4, 5, 6}...)
			mux.Lock()
			*getP = g
			mux.Unlock()
		}
	})
}
*/

/////////////////////////////////////////////////////////////////////////////////////////
//  Pass Pointer Vs. Return Value  //////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

func makeJsonBytes(name string, password []byte, data []interface{}, jBytes *[]byte) int {
	var jErr error
	*jBytes, jErr = json.Marshal(map[string]interface{}{
		"n": name,
		"p": password,
		"d": data,
	})
	if jErr != nil {
		return 1
	}
	return 0
}

func makeJsonBytes2(name string, password []byte, data []interface{}) ([]byte, int) {
	jBytes, jErr := json.Marshal(map[string]interface{}{
		"n": name,
		"p": password,
		"d": data,
	})
	if jErr != nil {
		return nil, 1
	}
	return jBytes, 0
}

// Winner by ~10ns
func BenchmarkPassPointer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var jBytes []byte
		if jErr := makeJsonBytes("someName", []byte("somePass"), []interface{}{1, 2, 3}, &jBytes); jErr != 0 {
			b.Errorf("Error: %v", jErr)
		}
	}
}


func BenchmarkReturnValue(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, jErr := makeJsonBytes2("someName", []byte("somePass"), []interface{}{1, 2, 3}); jErr != 0 {
			b.Errorf("Error: %v", jErr)
		}
	}
}

// run with:
// go test *filename* -bench=.