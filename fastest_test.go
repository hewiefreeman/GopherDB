package main

import (
	"testing"
	//"strconv"
	//"encoding/json"
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
//  Pass Pointer Vs. Return Value  //////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

/*
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
*/

/////////////////////////////////////////////////////////////////////////////////////////
//  Itteration Map vs Slice  ////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

/*
func BenchmarkIterMap(b *testing.B) {
	b.ReportAllocs()
	a := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
		"e": "e",
		"f": "f",
		"g": "g",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for n, _ := range a {
			a[n] = "hello";
		}
	}
}

// Winner; about 65x faster
func BenchmarkIterSlice(b *testing.B) {
	b.ReportAllocs()
	a := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"f",
		"g",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i, _ := range a {
			a[i] = "hello";
		}
	}
}
*/

/////////////////////////////////////////////////////////////////////////////////////////
//  Lookup Map vs Slice  ////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

/*
func BenchmarkLookupMap(b *testing.B) {
	b.ReportAllocs()
	a := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
		"e": "e",
		"f": "f",
		"g": "g",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a["g"]
	}
}

// Winner; about 45x faster
func BenchmarkLookupSlice(b *testing.B) {
	b.ReportAllocs()
	a := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"f",
		"g",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a[6]
	}
}
*/

/////////////////////////////////////////////////////////////////////////////////////////
//  Assign Map vs Slice  ////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

/*
func BenchmarkAssignMap(b *testing.B) {
	b.ReportAllocs()
	a := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
		"e": "e",
		"f": "f",
		"g": "g",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a["g"] = "h"
	}
}

// Winner; about 57x faster
func BenchmarkAssignSlice(b *testing.B) {
	b.ReportAllocs()
	a := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"f",
		"g",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a[6] = "h"
	}
}
*/

/////////////////////////////////////////////////////////////////////////////////////////
//  Pass Many vs Pass One  //////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

// STALEMATE

/*
func BenchmarkPassOne(b *testing.B) {
	b.ReportAllocs()
	a := passOne{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		onePassed(&a)
	}
}

func BenchmarkPassMany(b *testing.B) {
	b.ReportAllocs()
	a := 2
	f := "Up-B"
	c := "ok!"
	d := "ok!"
	e := "ok!" // - SSB Nes
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manyPassed(&a, &f, &c, &d, &e)
	}
}

func manyPassed(a *int, b *string, c *string, d *string, e *string) {
	*a = 12
	*b = "hello"
	*c = "hey"
	*d = "sup"
	*e = "yo"
}

type passOne struct {
	a int
	b string
	c string
	d string
	e string
}

func onePassed(one *passOne) {
	one.a = 12
	one.b = "hello"
	one.c = "hey"
	one.d = "sup"
	one.e = "yo"
}
*/

/////////////////////////////////////////////////////////////////////////////////////////
//  map[int] vs map[string]  ////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

/*
func BenchmarkMapString(b *testing.B) {
	b.ReportAllocs()
	a := map[string]string{
		"a": "hello",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a["a"]
	}
}

// Winner by a tiny margin
func BenchmarkMapInt(b *testing.B) {
	b.ReportAllocs()
	a := map[int]string{
		0: "hello",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a[0]
	}
}*/

/////////////////////////////////////////////////////////////////////////////////////////
//  JSON Marshal/Unmarshal map vs struct  ///////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

/*
type testStruct struct {
	Name string
	ID   int
	Data []interface{}
	Save bool
}

var (
	mData map[string]interface{} = map[string]interface{}{
		"Name": "billy",
		"ID": 2,
		"Data": []interface{}{1, 2, 3},
		"Save": false,
	}
	sData testStruct = testStruct{
		Name: "billy",
		ID: 2,
		Data: []interface{}{1, 2, 3},
		Save: false,
	}

	marshalMapData []byte
	marshalStructData []byte
	jErr error
)

func BenchmarkMarshalMap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Test json storing with map
		marshalMapData, jErr = json.Marshal(mData)
		if jErr != nil {
			b.Errorf("JSON error: %v", jErr)
			return
		}
	}
}

// More than 2x faster, 15:2 (map:struct) alloc/op
func BenchmarkMarshalStruct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Test json storing with struct
		marshalStructData, jErr = json.Marshal(sData)
		if jErr != nil {
			b.Errorf("JSON error: %v", jErr)
			return
		}
	}
}

func BenchmarkUnmarshalMap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Test json unmarshal with map
		jErr := json.Unmarshal(marshalMapData, &mData)
		if jErr != nil {
			b.Errorf("JSON error: %v", jErr)
			return
		}
	}
}

// About 900ns faster, 29:7 (map:struct) alloc/op
func BenchmarkUnmarshalStruct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Test json unmarshal with struct
		jErr := json.Unmarshal(marshalStructData, &sData)
		if jErr != nil {
			b.Errorf("JSON error: %v", jErr)
			return
		}
	}
}
*/

/////////////////////////////////////////////////////////////////////////////////////////
//  Map vs Struct alloc  ////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

type thing struct {
	A string
	B string
	C string
	D thing2
}

type thing2 struct {
	A string
	B string
	C string
}

func BenchmarkAllocStruct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = thing {
			A: "hello",
			B: "yo",
			C: "sup",
			D: thing2 {
				A: "goodbye",
				B: "peace",
				C: "later",
			},
		}
	}
}

func BenchmarkAllocMap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = map[string]interface{}{
			"A": "hello",
			"B": "yo",
			"C": "sup",
			"D": map[string]interface{}{
				"A": "goodbye",
				"B": "peace",
				"C": "later",
			},
		}
	}
}

// run with:
// go test *filename* -bench=.