package keystore

import (
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/keystore"
	"testing"
	"errors"
	"encoding/json"
	"strconv"
	"time"
	"fmt"
)

var (
	setupComplete bool
	table *keystore.Keystore
)

// TO TEST:
// go test -v keystore_test.go -bench=.
//
// Use -v to display fmt output

func setup() (bool, error) {
	if(!setupComplete) {
		// Set-up

		// Try to restore table & find out how long it took
		var tableErr int
		now := time.Now()
		table, tableErr = keystore.Restore("test")
		if tableErr == 0 {
			since := time.Since(now).Seconds()
			fmt.Printf("Restore success! Took %v seconds.", since)
			setupComplete = true
			return true, nil
		}

		newTableJson := "{\"NewKeystore\": [\"test\", {\"mmr\": [\"Uint16\", 0, 0, 0, false, false], \"email\": [\"String\", \"\", 0, false, false], \"subbed\": [\"Time\", \"RFC3339\", false]}, 0, 0, 0, 0]}"
		v := make(map[string]interface{})
		err := json.Unmarshal([]byte(newTableJson), &v)
		if err != nil {
			return false, err
		}

		// Get the schema object from the query
		s := v["NewKeystore"].([]interface{})[1].(map[string]interface{})

		// Make a schema with the query's schema object
		schemaObj, schemaErr := schema.New(s)
		if schemaErr != 0 {
			return false, errors.New("Schema error: " + strconv.Itoa(schemaErr))
		}

		// Make a new Keystore with the schema
		table, tableErr = keystore.New("test", nil, schemaObj, 0, true, false)
		if tableErr != 0 {
			return false, errors.New("Table create error: " + strconv.Itoa(tableErr))
		}

		//
		setupComplete = true
	}
	return false, nil
}

func BenchmarkInsert(b *testing.B) {
	b.ReportAllocs()
	var restored bool
	var sErr error
	if restored, sErr = setup(); sErr != nil {
		b.Errorf(sErr.Error())
		return
	}
	if restored {
		b.Errorf("Restored table... Skipping BenchmarkInsert()!")
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		is := strconv.Itoa(i)
		if _, iErr := table.InsertKey("guest"+is, map[string]interface{}{"email": "dinospumoni"+is+"@yahoo.com", "mmr": 1674}); iErr != 0 && iErr != helpers.ErrorKeyInUse {
			b.Errorf("Insert error (%v): %v", i, iErr)
			return
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	b.ReportAllocs()
	if table == nil {
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 240 vs 1 to test file update efficiency (200 near default max partition)
		if iErr := table.UpdateKey("guest240", map[string]interface{}{"mmr.*add.*mul": []interface{}{6, 0.9}}); iErr != 0 && iErr != helpers.ErrorNoEntryFound {
			b.Errorf("Update error (%v): %v", i, iErr)
			return
		}
	}
}

func BenchmarkGet(b *testing.B) {
	b.ReportAllocs()
	if table == nil {
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 240 vs 1 to test file read efficiency (200 near default max partition)
		if _, iErr := table.GetKeyData("guest240", []string{"mmr"}); iErr != 0 && iErr != helpers.ErrorNoEntryFound {
			b.Errorf("Get error (%v): %v", i, iErr)
			return
		}
	}
}