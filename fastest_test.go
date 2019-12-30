package main

import (
	"testing"
	//"strconv"
	//"encoding/json"
	"sort"
)

// run with:
// go test *filename* -bench=.

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
//  MAP VS SWITCH  //////////////////////////////////////////////////////////////////////
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

/*type thing struct {
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
}*/

/////////////////////////////////////////////////////////////////////////////////////////
//  Function w/ pointer call vs Pointer method call  ////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////

// Result: No difference!

/*type thing struct {
	A string
	B string
	C string
	D string
}

func BenchmarkFuncWithPointer(b *testing.B) {
	t := thing{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pointerFunc(&t)
	}
}

func BenchmarkPointerMethod(b *testing.B) {
	t := thing{}
	tp := &t
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tp.pointerMethod()
	}
}

func pointerFunc(p *thing) {
	p.A = "hello"
	p.B = "yo"
	p.C = "wassup"
	p.D = "woah man"
}

func (p *thing) pointerMethod() {
	p.A = "hello"
	p.B = "yo"
	p.C = "wassup"
	p.D = "woah man"
}*/

/////////////////////////////////////////////////////////////////////////////////////////
//  Non stack limited Quick-Sort vs No-Copy-Sort vs Go's stack limit Quick-Sort  ////////
/////////////////////////////////////////////////////////////////////////////////////////

// -1 allocs/op for each test to compensate for copying the sort array

/*
BenchmarkQuicksort-8                     5000000               244 ns/op             232 B/op          9 allocs/op
BenchmarkIdeksort-8                     30000000                53.3 ns/op            80 B/op          1 allocs/op
BenchmarkSortPackage-8                  10000000               224 ns/op             112 B/op          2 allocs/op
BenchmarkQuicksortReverse-8              5000000               329 ns/op             392 B/op          9 allocs/op
BenchmarkIdeksortReverse-8              30000000                51.8 ns/op            80 B/op          1 allocs/op
BenchmarkSortPackageReverse-8           10000000               207 ns/op             112 B/op          2 allocs/op
BenchmarkQuicksortLarge-8                 100000             16225 ns/op           23696 B/op        301 allocs/op
BenchmarkIdeksortLarge-8                   50000             39872 ns/op            2688 B/op          1 allocs/op
BenchmarkSortPackageLarge-8               100000             18999 ns/op            2720 B/op          2 allocs/op
BenchmarkQuicksortLargeReverse-8           20000             99400 ns/op          384248 B/op        300 allocs/op
BenchmarkIdeksortLargeReverse-8           100000             22359 ns/op            2688 B/op          1 allocs/op
BenchmarkSortPackageLargeReverse-8        100000             12477 ns/op            2720 B/op          2 allocs/op
*/

var (
	sortArr []int = []int{8, 5, 1, 7, 3, 9, 2, 4, 6}
	sortArrReverse []int = []int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	sortArrLarge []int = []int{174,37,210,116,28,249,147,99,206,97,291,91,75,182,115,119,279,30,113,136,54,27,273,140,266,29,26,171,224,38,166,6,77,154,226,50,265,55,255,220,205,236,89,61,288,270,126,44,227,123,168,138,215,207,232,267,88,141,69,106,25,78,196,18,149,191,217,129,218,233,222,19,173,103,74,95,143,189,300,23,243,1,282,184,254,176,22,260,132,246,35,295,80,111,169,117,108,32,8,177,192,86,287,201,229,133,3,98,90,92,67,248,269,240,294,101,76,142,162,58,39,299,9,180,278,185,7,209,160,198,107,290,231,135,259,43,2,170,157,14,234,84,4,124,208,34,161,186,203,122,179,153,41,172,15,280,272,244,52,53,104,204,144,245,211,49,297,110,82,42,24,100,48,10,139,47,190,221,258,16,164,202,283,262,235,264,146,263,195,105,275,94,241,293,250,36,5,79,188,228,257,239,223,216,276,137,159,167,68,17,60,33,65,163,155,127,197,45,292,242,93,12,178,120,20,81,134,151,261,46,271,130,238,286,72,114,213,181,175,253,21,102,183,152,64,40,284,96,85,150,63,165,225,247,118,274,296,156,268,51,285,256,214,194,237,66,251,199,230,145,200,193,158,70,131,13,128,277,57,59,121,187,11,73,87,252,71,109,31,212,125,56,219,298,281,83,14,8,289,62,112}
	sortArrLargeReverse []int = []int{300,299,298,297,296,295,294,293,292,291,290,289,288,287,286,285,284,283,282,281,280,279,278,277,276,275,274,273,272,271,270,269,268,267,266,265,264,263,262,261,260,259,258,257,256,255,254,253,252,251,250,249,248,247,246,245,244,243,242,241,240,239,238,237,236,235,234,233,232,231,230,229,228,227,226,225,224,223,222,221,220,219,218,217,216,215,214,213,212,211,210,209,208,207,206,205,204,203,202,201,200,199,198,197,196,195,194,193,192,191,190,189,188,187,186,185,184,183,182,181,180,179,178,177,176,175,174,173,172,171,170,169,168,167,166,165,164,163,162,161,160,159,158,157,156,155,154,153,152,151,150,149,148,147,146,145,144,143,142,141,140,139,138,137,136,135,134,133,132,131,130,129,128,127,126,125,124,123,122,121,120,119,118,117,116,115,114,113,112,111,110,109,108,107,106,105,104,103,102,101,100,99,98,97,96,95,94,93,92,91,90,89,88,87,86,85,84,83,82,81,80,79,78,77,76,75,74,73,72,71,70,69,68,67,66,65,64,63,62,61,60,59,58,57,56,55,54,53,52,51,50,49,48,47,46,45,44,43,42,41,40,39,38,37,36,35,34,33,32,31,30,29,28,27,26,25,24,23,22,21,20,19,18,17,16,15,14,13,12,11,10,9,8,7,6,5,4,3,2,1}
)

// Compare with random small array
func BenchmarkQuicksort(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		quicksort(append([]int{}, sortArr...))
	}
}

func BenchmarkNoCopySort(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		noCopySort(append([]int{}, sortArr...))
	}
}

func BenchmarkSortPackage(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Ints(append([]int{}, sortArr...))
	}
}

// Compare with reverse small array
func BenchmarkQuicksortReverse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		quicksort(append([]int{}, sortArrReverse...))
	}
}


func BenchmarkNoCopySortReverse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		noCopySort(append([]int{}, sortArrReverse...))
	}
}

func BenchmarkSortPackageReverse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Ints(append([]int{}, sortArrReverse...))
	}
}

// Compare with random large array
func BenchmarkQuicksortLarge(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		quicksort(append([]int{}, sortArrLarge[:150]...))
	}
}


func BenchmarkNoCopySortLarge(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		noCopySort(append([]int{}, sortArrLarge[:150]...))
	}
}

func BenchmarkSortPackageLarge(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Ints(append([]int{}, sortArrLarge[:150]...))
	}
}

// Compare with reversed large array
func BenchmarkQuicksortLargeReverse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		quicksort(append([]int{}, sortArrLargeReverse[:150]...))
	}
}


func BenchmarkNoCopySortLargeReverse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		noCopySort(append([]int{}, sortArrLargeReverse[:150]...))
	}
}

func BenchmarkSortPackageLargeReverse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Ints(append([]int{}, sortArrLargeReverse[:150]...))
	}
}

// Quicksort
func quicksort(sa []int) {
	if len(sa) < 2 {
		return
	}
	var i int = -1;
	var j int
	var t int // temp for swaps
	for j = 0; j < (len(sa) - 1); j++ {
		if sa[j] < sa[len(sa) - 1] {
			i++
			// swap i and j
			t = sa[i]
			sa[i] = sa[j]
			sa[j] = t
		}
	}
	part1 := append([]int{}, sa[:i + 1]...)
	part2 := append([]int{}, sa[i + 1:len(sa) - 1]...)
	quicksort(part1)
	quicksort(part2)
	i = -1
	for j, t = range part1 {
		sa[j] = t
		i++
	}
	sa[i + 1] = sa[len(sa) - 1]
	i = i + 2
	for j = 0; j < len(part2); j++ {
		sa[i] = part2[j]
		i++
	}
}

// NoCopySort
func noCopySort(sa []int) {
	var t int
	for i := 0; i < len(sa) - 1; i++ {
		for j := len(sa) - 1; j > i; j-- {
			if sa[i] > sa[j] {
				t = sa[i]
				sa[i] = sa[j]
				sa[j] = t
			}
		}
	}
}