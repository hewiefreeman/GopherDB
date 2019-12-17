package keystore

import (
	"github.com/hewiefreeman/GopherDB/keystore"
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/storage"
	"testing"
	"errors"
	"strconv"
	"time"
	"fmt"
)

var (
	// Test settings
	tableName string = "test"
	tablePartitionMax uint16 = 250
	tableMaxEntries uint64 = 1000000
	tableEncryptionCost int = 4;

	// Test variables
	setupComplete bool
	table *keystore.Keystore
)

// TO TEST:
// go test -v keystore_test.go
//
// Use -v to display fmt output

// NOTE: KS-test/0.gdbs must contain at least the first "Vokome" entry.
// NOTE: All .gdbs files must have a new line at the end of the file to work properly.

func restore() (bool, error) {
	// Initialize storage engine
	storage.Init()

	if(!setupComplete) {
		var tableErr int
		now := time.Now()
		table, tableErr = keystore.Restore(tableName)
		since := time.Since(now).Seconds()
		if tableErr != 0 {
			return false, errors.New("Fatal restore error: " + strconv.Itoa(tableErr))
		} else if table.Size() == 0 {
			return false, errors.New("Restored 0 entries! Skipping tests...")
		}
		fmt.Printf("Restore success! Took %v seconds to restore %v keys.\n", since, table.Size())
		setupComplete = true
		return true, nil
	}
	return false, nil
}

func TestRestore(t *testing.T) {
	// Run restore() to restore "test" table
	if ok, err := restore(); !ok {
		t.Errorf("Error while restoring '%v' table: %v", tableName, err)
	}
}

func TestChangeSettings(t *testing.T) {
	if (!setupComplete) {
		t.Skip("Skipping tests...")
	}

	storage.SetFileOpenTime(3 * time.Second)

	// Set max partition file size
	err := table.SetPartitionMax(tablePartitionMax)
	if err != 0 {
		t.Errorf("Error while setting max partition file size: %v", err)
	}

	// Set max table entries
	err = table.SetMaxEntries(tableMaxEntries)
	if err != 0 {
		t.Errorf("Error while setting max table entries: %v", err)
	}

	// Set table encryption cost
	err = table.SetEncryptionCost(tableEncryptionCost)
	if err != 0 {
		t.Errorf("Error while setting table encryption cost: %v", err)
	}
}

func TestInsert(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	var guestName string = "guest" + strconv.Itoa(table.Size() + 1);
	_, err := table.InsertKey(guestName, map[string]interface{}{"mmr": 1337, "email": guestName + "@gmail.com"})
	if (err != 0) {
		t.Errorf("Error while inserting '%v': %v", guestName, err)
	}
}

func TestInsertMissingRequiredItem(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	_, err := table.InsertKey("guest" + strconv.Itoa(table.Size() + 1), map[string]interface{}{"mmr": 1337})
	if (err != helpers.ErrorMissingRequiredItem) {
		t.Errorf("InsertDuplicateUniqueTableValue expected error %v, but got: %v", helpers.ErrorMissingRequiredItem, err)
	}
}

func TestInsertDuplicateUniqueTableValue(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	_, err := table.InsertKey("guest" + strconv.Itoa(table.Size() + 1), map[string]interface{}{"email": "guest" + strconv.Itoa(table.Size()) + "@gmail.com"})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("InsertDuplicateUniqueTableValue expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
}

func TestInsertMissingRequiredNestedItemArray(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	var guestName string = "guest" + strconv.Itoa(table.Size() + 1);
	_, err := table.InsertKey(guestName, map[string]interface{}{"mmr": 1337, "email": guestName + "@gmail.com", "friends": []interface{}{map[string]interface{}{"status": 0}}})
	if (err != helpers.ErrorMissingRequiredItem) {
		t.Errorf("TestInsertMissingRequiredNestedItemArray expected error %v, but got: %v", helpers.ErrorMissingRequiredItem, err)
	}
}

func TestAppendDuplicateUniqueNestedValueArray(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("Vokome", map[string]interface{}{"friends.*append": []interface{}{map[string]interface{}{"login": "Sir Smackem", "status": 0, "labels": map[string]interface{}{"nickname":"Oni","friendNum":666}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendDuplicateUniqueNestedValueArray expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
	// Test deeper nesting...
	err = table.UpdateKey("Vokome", map[string]interface{}{"friends.0.labels.nickname": "H"})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendDuplicateUniqueNestedValueArray expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
	// Testing Int8...
	err = table.UpdateKey("Vokome", map[string]interface{}{"friends.2.labels.friendNum": 0})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendDuplicateUniqueNestedValueArray expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
	time.Sleep(time.Second * 1)
}

func TestInsertWithUniqueValueDuplicatesArray(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	var guestName string = "guest" + strconv.Itoa(table.Size() + 1);
	_, err := table.InsertKey(guestName, map[string]interface{}{"mmr": 1337, "email": guestName + "@gmail.com", "friends": []interface{}{map[string]interface{}{"login": "Vokome", "status": 0, "labels": map[string]interface{}{"nickname":"Oni","friendNum":666}}, map[string]interface{}{"login": "Vokome", "status": 1, "labels": map[string]interface{}{"nickname":"rawrrr","friendNum":432}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestInsertWithUniqueValueDuplicatesArray expected error %v but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
	// Test nested unique Object item
	_, err = table.InsertKey(guestName, map[string]interface{}{"mmr": 1337, "email": guestName + "@gmail.com", "friends": []interface{}{map[string]interface{}{"login": "Moe", "status": 0, "labels": map[string]interface{}{"nickname":"Moe","friendNum":27}}, map[string]interface{}{"login": "Bob", "status": 1, "labels": map[string]interface{}{"nickname":"Moe","friendNum":27}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestInsertWithUniqueValueDuplicatesArray expected error %v but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
	time.Sleep(time.Second * 1)
}

func TestAppendWithUniqueValueDuplicatesArray(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.*append": []interface{}{map[string]interface{}{"login": "Vokome", "status": 0, "labels": map[string]interface{}{"nickname":"Oni","friendNum":666}}, map[string]interface{}{"login": "Vokome", "status": 1, "labels": map[string]interface{}{"nickname":"rawrrr","friendNum":432}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendWithUniqueValueDuplicatesArray expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
	// Test nested unique Object item
	err = table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.*append": []interface{}{map[string]interface{}{"login": "Moe", "status": 0, "labels": map[string]interface{}{"nickname":"Moe","friendNum":27}}, map[string]interface{}{"login": "Bob", "status": 1, "labels": map[string]interface{}{"nickname":"Moe","friendNum":27}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendWithUniqueValueDuplicatesArray expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
}

func TestInsertMissingRequiredNestedItemMap(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	var guestName string = "guest" + strconv.Itoa(table.Size() + 1);
	_, err := table.InsertKey(guestName, map[string]interface{}{"mmr": 1337, "email": guestName + "@gmail.com", "actions": map[string]interface{}{"yo": map[string]interface{}{"type":"greeting"}}})
	if (err != helpers.ErrorMissingRequiredItem) {
		t.Errorf("TestInsertMissingRequiredNestedItemMap expected error %v, but got: %v", helpers.ErrorMissingRequiredItem, err)
	}
}

func TestAppendDuplicateUniqueNestedValueMap(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("Vokome", map[string]interface{}{"actions.*append": map[string]interface{}{"yo": map[string]interface{}{"type": "greeting", "id": 1}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendDuplicateUniqueNestedValueMap expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
	// Testing Uint16...
	err = table.UpdateKey("Vokome", map[string]interface{}{"actions.*append": map[string]interface{}{"fek off": map[string]interface{}{"type": "insult", "id": 0}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendDuplicateUniqueNestedValueMap expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
}

func TestInsertWithUniqueValueDuplicatesMap(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	var guestName string = "guest" + strconv.Itoa(table.Size() + 1);
	_, err := table.InsertKey(guestName, map[string]interface{}{"mmr": 1337, "email": guestName + "@gmail.com", "actions": map[string]interface{}{"hi": map[string]interface{}{"type": "greeting", "id": 0}, "yo": map[string]interface{}{"type": "greeting", "id": 1}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestInsertWithUniqueValueDuplicatesMap expected error %v but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
}

func TestAppendWithUniqueValueDuplicatesMap(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*append": map[string]interface{}{"hi": map[string]interface{}{"type": "greeting", "id": 0}, "yo": map[string]interface{}{"type": "greeting", "id": 1}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendWithUniqueValueDuplicatesMap expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
}

func TestGet(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	data, err := table.GetKeyData("guest" + strconv.Itoa(table.Size()), []string{"mmr"})
	if err != 0 {
		t.Errorf("TestGet error: %v", err)
	} else if data["mmr"] != float64(1337) {
		t.Errorf("TestGet expected 1337, but got: %v", data["mmr"])
	}
	time.Sleep(time.Second * 3)
}

func TestGetArrayLength(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	data, err := table.GetKeyData("Vokome", []string{"friends.*len"})
	if err != 0 {
		t.Errorf("TestGetArrayLength error: %v", err)
	} else if data["friends.*len"] != 3 {
		t.Errorf("TestGetArrayLength expected 3, but got: %v", data["friends.*len"])
	}
}

func TestGetMapLength(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	data, err := table.GetKeyData("Vokome", []string{"actions.*len"})
	if err != 0 {
		t.Errorf("TestGetMapLength error: %v", err)
	} else if data["actions.*len"] != 1 {
		t.Errorf("TestGetMapLength expected 1, but got: %v", data["actions.*len"])
	}
}

func TestAppendToArray(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	for i := 0; i < 3; i++ {
		err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.*append": []interface{}{map[string]interface{}{"login": "guest133" + strconv.Itoa(7+i), "status": 0, "labels": map[string]interface{}{"nickname": "G" + strconv.Itoa(7+i), "friendNum": i}}}})
		if (err != 0) {
			t.Errorf("TestAppendArray error: %v", err)
		}
	}
	time.Sleep(time.Second * 4)
}

func TestAppendToMap(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*append": map[string]interface{}{"fek off": map[string]interface{}{"type": "insult", "id": 1}}})
	if (err != 0) {
		t.Errorf("TestAppendMap error: %v", err)
	}
	err = table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*append": map[string]interface{}{"hallo": map[string]interface{}{"type": "greeting", "id": 0}}})
	if (err != 0) {
		t.Errorf("TestAppendMap error: %v", err)
	}
	err = table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*append": map[string]interface{}{"peace": map[string]interface{}{"type": "farewell", "id": 2}}})
	if (err != 0) {
		t.Errorf("TestAppendMap error: %v", err)
	}
}

func TestArithmetic(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"mmr.*add.*sub.*mul.*div.*mod": []interface{}{10, 7, 2, 3, 8}})
	if err != 0 {
		t.Errorf("TestArithmetic error: %v", err)
	}
	data, _ := table.GetKeyData("guest" + strconv.Itoa(table.Size()), []string{"mmr"})
	if data["mmr"] != float64(5) {
		t.Errorf("TestArithmetic expected 5, but got: %v", data["mmr"])
	}
}

// Must be last test!!
func TestLetFilesClose(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	time.Sleep(4 * time.Second)
	if storage.GetNumOpenFiles() != 0 {
		t.Errorf("Storage files did not close properly!")
	}
}