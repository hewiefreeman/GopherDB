package main

import (
	"github.com/json-iterator/go"
	"encoding/json"
	"testing"
	//"fmt"
)

type Key struct {
	K string
	D []interface{}
}

var (
	fjson = jsoniter.ConfigCompatibleWithStandardLibrary
	cData []byte = []byte("{\"K\":\"Vokome\",\"D\":[1674,\"dinospumoni1@yahoo.com\",\"2019-09-04T21:10:44.9724398-07:00\",[[\"Mary\",2,[0,\"M\"]],[\"Harry Potter\",2,[2,\"H\"]],[\"Sir Smackem\",1,[1,\"S\"]]],{\"hello\":[0,\"greeting\"]},[\"a\",\"b\",\"c\",\"d\",\"e\",\"f\",\"g\"],[1,2.5,4,5.5,7],{\"one\": 1, \"two\": 2, \"three point 45\": 3.45}]}")
	m map[string]interface{}
	mU Key
)

func BenchmarkJsonIterUnmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := fjson.Unmarshal(cData, &mU); err != nil {
			b.Errorf("Error while iter unmarshaling: %v", err)
			return
		}
	}
}

func BenchmarkStdJsonUnmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(cData, &mU); err != nil {
			b.Errorf("Error while std unmarshaling: %v", err)
			return
		}
	}
}

func BenchmarkJsonIterMarshal(b *testing.B) {
	var err error
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if cData, err = fjson.Marshal(&mU); err != nil {
			b.Errorf("Error while iter unmarshaling: %v", err)
			return
		}
	}
}

func BenchmarkStdJsonMarshal(b *testing.B) {
	var err error
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if cData, err = json.Marshal(&mU); err != nil {
			b.Errorf("Error while iter unmarshaling: %v", err)
			return
		}
	}
}