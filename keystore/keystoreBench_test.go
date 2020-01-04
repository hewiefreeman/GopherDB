package keystore

import (
	"errors"
	"fmt"
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/keystore"
	"github.com/hewiefreeman/GopherDB/storage"
	"strconv"
	"testing"
	"time"
)

const (
	// Test settings
	tableName string = "test"
)

var (
	// Test variables
	setupComplete bool
	table         *keystore.Keystore
)

// TO TEST:
// go test -v keystoreBench_test.go -bench=.
//
// Use -v to display fmt output

// NOTE: All .gdbs files must have a new line at the end of the file to work properly.
//
// WARNING: This benchmark could potentially permanently damage SSD/flash space!
//          Please use with caution, review the code, and run these benchmarks on disk storage.
//          GopherDB and it's creators are not liable for any damages these benchmarks may cause.

func restore() (bool, error) {
	// Initialize storage engine
	storage.Init()

	//
	if !setupComplete {
		var tableErr int
		now := time.Now()
		table, tableErr = keystore.Restore(tableName)
		since := time.Since(now).Seconds()
		if tableErr != 0 {
			return false, errors.New("Fatal restore error: " + strconv.Itoa(tableErr))
		} else if table.Size() == 0 {
			return false, errors.New("Restored 0 entries! Skipping benchmarks...")
		}
		fmt.Printf("Restore success! Took %v seconds to restore %v keys.\n", since, table.Size())
		setupComplete = true
		return true, nil
	}
	return false, nil
}

/*func BenchmarkInsert(b *testing.B) {
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
}*/

func BenchmarkUpdateValue(b *testing.B) {
	var err error
	if !setupComplete {
		if setupComplete, err = restore(); err != nil {
			b.Errorf("Error restoring: %v", err)
			return
		}
	}
	b.ReportAllocs()
	if table == nil {
		b.Errorf("Update error: table is nil!")
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Both Vokome's get appended because no checks are done on the input array!
		if iErr := table.UpdateKey("Harry Potter", map[string]interface{}{"mmr": 1002}); iErr != 0 && iErr != helpers.ErrorNoEntryFound {
			b.Errorf("Update error (%v): %v", i, iErr)
			return
		}
	}
}

func BenchmarkRunMethodOnValue(b *testing.B) {
	b.ReportAllocs()
	if table == nil {
		b.Errorf("Update error: table is nil!")
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Both Vokome's get appended because no checks are done on the input array!
		if iErr := table.UpdateKey("Harry Potter", map[string]interface{}{"mmr.*add": []interface{}{1}}); iErr != 0 && iErr != helpers.ErrorNoEntryFound {
			b.Errorf("Update error (%v): %v", i, iErr)
			return
		}
	}
}

func BenchmarkGetValue(b *testing.B) {
	b.ReportAllocs()
	if table == nil || !setupComplete {
		b.Skip()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 240 vs 1 to test file read efficiency (200 is near default max partition)
		if _, iErr := table.GetKeyData("Harry Potter", map[string]interface{}{"mmr": nil}); iErr != 0 && iErr != helpers.ErrorNoEntryFound {
			b.Errorf("Get error (%v): %v", i, iErr)
			return
		}
	}
	table.Close(true)
}
