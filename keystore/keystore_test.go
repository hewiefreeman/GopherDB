package keystore

import (
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/keystore"
	"testing"
	"errors"
	"encoding/json"
	"strconv"
	//"fmt"
)

var (
	setupComplete bool
	table *keystore.Keystore
)

// TO TEST:
// go test authtable_test.go -bench=.

func setup() error {
	if(!setupComplete) {
		// Set-up
		newTableJson := "{\"NewKeystore\": [\"storage\",{\"email\": [\"String\", \"\", 0, true, true],\"friends\": [\"Array\", [\"Object\", {\"name\": [\"String\", \"\", 0, true, true],\"status\": [\"Uint8\", 0, 0, 2, false, false]}, false], 50, false],\"vCode\": [\"String\", \"\", 0, true, false],\"verified\": [\"Bool\", false], \"mmr\": [\"Uint16\", 1500, 1100, 2250, false, false], \"testMap\": [\"Map\", [\"Map\", [\"Int16\", 100, 0, 0, true, true, true], 0, false], 0, false], \"timeStamp\": [\"Time\", \"Kitchen\", false]}, 0, 0, 0, 0]}"
		v := make(map[string]interface{})
		err := json.Unmarshal([]byte(newTableJson), &v)
		if err != nil {
			return err
		}

		// Get the schema object from the query
		s := v["NewKeystore"].([]interface{})[1].(map[string]interface{})

		// Make a schema with the query's schema object
		schemaObj, schemaErr := schema.New(s)
		if schemaErr != 0 {
			return errors.New("Schema error: " + strconv.Itoa(schemaErr))
		}

		// Make a new Keystore with the schema
		var tableErr int
		table, tableErr = keystore.New("storage", schemaObj, 0, false, true)
		if tableErr != 0 {
			return errors.New("Table create error: " + strconv.Itoa(tableErr))
		} else if table == nil {
			return errors.New("Nil table?")
		}

		//
		setupComplete = true
	}
	return nil
}

func BenchmarkInsert(b *testing.B) {
	b.ReportAllocs()
	if sErr := setup(); sErr != nil {
		b.Errorf(sErr.Error())
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		is := strconv.Itoa(i)
		if iErr := table.Insert("guest"+is, map[string]interface{}{"email": "dinospumoni"+is+"@yahoo.com", "mmr": 1674, "vCode": "06AJ3T9"}); iErr != 0 && iErr != helpers.ErrorKeyInUse {
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
		is := strconv.Itoa(i)
		if iErr := table.UpdateData("guest"+is, map[string]interface{}{"mmr.*add.*mul": []interface{}{6, 0.9}}); iErr != 0 && iErr != helpers.ErrorNoEntryFound {
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
		is := strconv.Itoa(i)
		if _, iErr := table.GetData("guest"+is, []string{"mmr"}); iErr != 0 && iErr != helpers.ErrorNoEntryFound {
			b.Errorf("Get error (%v): %v", i, iErr)
			return
		}
	}
}