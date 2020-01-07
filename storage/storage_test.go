package storage

import (
	"github.com/hewiefreeman/GopherDB/storage"
	"testing"
	"fmt"
)

func TestStorageInsert(t *testing.T) {
	storage.Init()
	var b uint16
	var err int
	if b, err = storage.Insert("0.gdbs", []byte("123geegee")); err != 0 {
		t.Errorf("Error inserting to file: %v", err)
	}
	fmt.Println("Inserted to line: ", b)
}

func TestStorageGet(t *testing.T) {
	var b []byte
	var err int
	if b, err = storage.Read("0.gdbs", 1); err != 0 {
		t.Errorf("Error reading file: %v", err)
	}
	fmt.Println(string(b))
}