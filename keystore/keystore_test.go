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

const (
	// Test settings
	tableName string = "test"
	tablePartitionMax uint16 = 250
	tableMaxEntries uint64 = 1000000
	tableEncryptionCost int = 4
)

var (
	// Test variables
	setupComplete bool
	table *keystore.Keystore
)

// TO TEST:
// go test -v keystore_test.go
//
// Use -v to display fmt output

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

	// Set fileOpenTime to 3 seconds (needs 3 sec or lower for testing file closing at end)
	storage.SetFileOpenTime(3 * time.Second)

	// Set maxOpenFiles to 2 (to test auto file close when max has been reached) - Uncomment for testing
	//storage.SetMaxOpenFiles(2)

	// Set max partition file size
	err := table.SetPartitionMax(tablePartitionMax)
	if err != 0 {
		t.Errorf("Error while setting max partition file size: %v", err)
		return
	}

	// Set max table entries
	err = table.SetMaxEntries(tableMaxEntries)
	if err != 0 {
		t.Errorf("Error while setting max table entries: %v", err)
		return
	}

	// Set table encryption cost
	err = table.SetEncryptionCost(tableEncryptionCost)
	if err != 0 {
		t.Errorf("Error while setting table encryption cost: %v", err)
		return
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
	_, err := table.InsertKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"email": "guest" + strconv.Itoa(table.Size()) + "@gmail.com"})
	if (err != helpers.ErrorKeyInUse) {
		t.Errorf("InsertDuplicateUniqueTableValue expected error %v, but got: %v", helpers.ErrorKeyInUse, err)
		return
	}
	// Test duplicate key
	_, err = table.InsertKey("guest" + strconv.Itoa(table.Size() + 1), map[string]interface{}{"email": "guest" + strconv.Itoa(table.Size()) + "@gmail.com"})
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
	err := table.UpdateKey("Vokome", map[string]interface{}{"friends.*append": []interface{}{[]interface{}{map[string]interface{}{"login": "Sir Smackem", "status": 0, "labels": map[string]interface{}{"nickname":"Oni","friendNum":666}}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendDuplicateUniqueNestedValueArray expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
		return
	}
	// Test deeper nesting...
	err = table.UpdateKey("Vokome", map[string]interface{}{"friends.0.labels.nickname": "H"})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendDuplicateUniqueNestedValueArray expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
		return
	}
	// Testing Int8...
	err = table.UpdateKey("Vokome", map[string]interface{}{"friends.2.labels.friendNum": 0})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendDuplicateUniqueNestedValueArray expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
}

func TestInsertWithUniqueValueDuplicatesArray(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	var guestName string = "guest" + strconv.Itoa(table.Size() + 1);
	_, err := table.InsertKey(guestName, map[string]interface{}{"mmr": 1337, "email": guestName + "@gmail.com", "friends": []interface{}{map[string]interface{}{"login": "Vokome", "status": 0, "labels": map[string]interface{}{"nickname":"Oni","friendNum":666}}, map[string]interface{}{"login": "Vokome", "status": 1, "labels": map[string]interface{}{"nickname":"rawrrr","friendNum":432}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestInsertWithUniqueValueDuplicatesArray expected error %v but got: %v", helpers.ErrorUniqueValueDuplicate, err)
		return
	}
	// Test nested unique Object item
	_, err = table.InsertKey(guestName, map[string]interface{}{"mmr": 1337, "email": guestName + "@gmail.com", "friends": []interface{}{map[string]interface{}{"login": "Moe", "status": 0, "labels": map[string]interface{}{"nickname":"Moe","friendNum":27}}, map[string]interface{}{"login": "Bob", "status": 1, "labels": map[string]interface{}{"nickname":"Moe","friendNum":27}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestInsertWithUniqueValueDuplicatesArray expected error %v but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
}

func TestAppendWithUniqueValueDuplicatesArray(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.*append": []interface{}{[]interface{}{map[string]interface{}{"login": "Vokome", "status": 0, "labels": map[string]interface{}{"nickname":"Oni","friendNum":666}}, map[string]interface{}{"login": "Vokome", "status": 1, "labels": map[string]interface{}{"nickname":"rawrrr","friendNum":432}}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendWithUniqueValueDuplicatesArray expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
		return
	}
	// Test nested unique Object item
	err = table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.*append": []interface{}{[]interface{}{map[string]interface{}{"login": "Moe", "status": 0, "labels": map[string]interface{}{"nickname":"Moe","friendNum":27}}, map[string]interface{}{"login": "Bob", "status": 1, "labels": map[string]interface{}{"nickname":"Moe","friendNum":27}}}}})
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
	err := table.UpdateKey("Vokome", map[string]interface{}{"actions.*append": []interface{}{map[string]interface{}{"yo": map[string]interface{}{"type": "greeting", "id": 1}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendDuplicateUniqueNestedValueMap expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
		return
	}
	// Testing Uint16...
	err = table.UpdateKey("Vokome", map[string]interface{}{"actions.*append": []interface{}{map[string]interface{}{"fek off": map[string]interface{}{"type": "insult", "id": 0}}}})
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
	err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*append": []interface{}{map[string]interface{}{"hi": map[string]interface{}{"type": "greeting", "id": 0}, "yo": map[string]interface{}{"type": "greeting", "id": 1}}}})
	if (err != helpers.ErrorUniqueValueDuplicate) {
		t.Errorf("TestAppendWithUniqueValueDuplicatesMap expected error %v, but got: %v", helpers.ErrorUniqueValueDuplicate, err)
	}
}

func TestGet(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	data, err := table.GetKeyData("Harry Potter", nil)
	if err != 0 {
		t.Errorf("TestGet error: %v", err)
	} else if data["mmr"] != float64(1674) {
		t.Errorf("TestGet expected 1674, but got: %v\n", data["mmr"])
		t.Errorf("TestGet recieved: %v\n", data)
	}
}

func TestGetArrayLength(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	data, err := table.GetKeyData("Mary", map[string]interface{}{"friends.*len": []interface{}{}})
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
	data, err := table.GetKeyData("Vokome", map[string]interface{}{"actions.*len": []interface{}{}})
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
		err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.*append": []interface{}{[]interface{}{map[string]interface{}{"login": "guest133" + strconv.Itoa(7+i), "status": 0, "labels": map[string]interface{}{"nickname": "G" + strconv.Itoa(7+i), "friendNum": i}}}}})
		if (err != 0) {
			t.Errorf("TestAppendArray error: %v", err)
			return
		}
	}
}

func TestChangeValueInnerArray(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	for i := 0; i < 3; i++ {
		err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends." + strconv.Itoa(i) + ".status": 2})
		if (err != 0) {
			t.Errorf("TestAppendArray error: %v", err)
			return
		}
	}
}

func TestAppendToMap(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*append": []interface{}{map[string]interface{}{"fek off": map[string]interface{}{"type": "insult", "id": 1}}}})
	if (err != 0) {
		t.Errorf("TestAppendMap error: %v", err)
		return
	}
	err = table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*append": []interface{}{map[string]interface{}{"hallo": map[string]interface{}{"type": "greeting", "id": 0}}}})
	if (err != 0) {
		t.Errorf("TestAppendMap error: %v", err)
		return
	}
	err = table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*append": []interface{}{map[string]interface{}{"peace": map[string]interface{}{"type": "farewell", "id": 2}}}})
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
		return
	}
	data, _ := table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"mmr": nil})
	if data["mmr"] != float64(5) {
		t.Errorf("TestArithmetic expected 5, but got: %v", data["mmr"])
		return
	}
}

// Testing number and string comparisons
func TestComparisons(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	// Number comparisons
	data, err := table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"mmr.*add.*eq": []interface{}{10, 15}})
	if err != 0 {
		t.Errorf("TestComparisons error: %v", err)
		return
	}
	if data["mmr.*add.*eq"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["mmr.*add.*eq"])
		return
	}
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"mmr.*add.*eq": []interface{}{10, 43}})
	if data["mmr.*add.*eq"] != false {
		t.Errorf("TestComparisons expected false, but got: %v", data["mmr.*add.*eq"])
		return
	}
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"mmr.*add.*lt": []interface{}{10, 15}})
	if data["mmr.*add.*lt"] != false {
		t.Errorf("TestComparisons expected false, but got: %v", data["mmr.*add.*lt"])
		return
	}
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"mmr.*add.*lte": []interface{}{10, 15}})
	if data["mmr.*add.*lte"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["mmr.*add.*lte"])
		return
	}
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"mmr.*add.*gt": []interface{}{10, 14}})
	if data["mmr.*add.*gt"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["mmr.*add.*gt"])
		return
	}
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"mmr.*add.*gte": []interface{}{10, 15}})
	if data["mmr.*add.*gte"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["mmr.*add.*gte"])
		return
	}
	// combining array methods with number comparisons
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.*len.*gte": []interface{}{5}})
	if data["friends.*len.*gte"] != false {
		t.Errorf("TestComparisons expected false, but got: %v", data["friends.*len.*gte"])
		return
	}
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.*len.*gte": []interface{}{2}})
	if data["friends.*len.*gte"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["friends.*len.*gte"])
		return
	}
	// combining map methods with number comparisons
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*len.*gte": []interface{}{5}})
	if data["actions.*len.*gte"] != false {
		t.Errorf("TestComparisons expected false, but got: %v", data["actions.*len.*gte"])
		return
	}
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*len.*gte": []interface{}{2}})
	if data["actions.*len.*gte"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["actions.*len.*gte"])
		return
	}
	// String comparisons
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"email.*eq": []interface{}{"guest" + strconv.Itoa(table.Size()) + "@gmail.com"}})
	if data["email.*eq"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["email.*eq"])
		return
	}
	data, _ = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"email.*len.*gt": []interface{}{30}})
	if data["email.*len.*gt"] != false {
		t.Errorf("TestComparisons expected false, but got: %v", data["email.*len.*gt"])
		return
	}
	// friends.0.login = "guest1337"
	data, err = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.0.login.*len.*eq": []interface{}{9}})
	if err != 0 {
		t.Errorf("TestComparisons error: %v", err)
		return
	}
	if data["friends.0.login.*len.*eq"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["friends.0.login.*len.*eq"])
		return
	}
	// actions.fek off.type = "insult"
	data, err = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.fek off.type.*len.*add.*eq": []interface{}{3, 9}})
	if err != 0 {
		t.Errorf("TestComparisons error: %v", err)
	}
	if data["actions.fek off.type.*len.*add.*eq"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["actions.fek off.type.*len.*add.*eq"])
		return
	}
	data, err = table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.fek off.id.*gte": []interface{}{1}})
	if err != 0 {
		t.Errorf("TestComparisons error: %v", err)
		return
	}
	if data["actions.fek off.id.*gte"] != true {
		t.Errorf("TestComparisons expected true, but got: %v", data["actions.fek off.id.*gte"])
	}
}

// Testing Array *indexOf and *contains
func TestArrayIndexOfAndContains(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	// Array of String
	data, err := table.GetKeyData("Vokome", map[string]interface{}{"testStringArray.*indexOf": []interface{}{"c"}, "testStringArray.*contains": []interface{}{"h"}})
	if err != 0 {
		t.Errorf("TestArrayIndexOfAndContains error: %v", err)
		return
	}
	if data["testStringArray.*indexOf"] != float64(2) {
		t.Errorf("TestArrayIndexOfAndContains expected 2, but got: %v", data["testStringArray.*indexOf"])
		return
	}
	if data["testStringArray.*contains"] != false {
		t.Errorf("TestArrayIndexOfAndContains expected false, but got: %v", data["testStringArray.*contains"])
		return
	}
	// Array of Float32
	data, err = table.GetKeyData("Vokome", map[string]interface{}{"testFloatArray.*indexOf": []interface{}{5.5}, "testFloatArray.*contains": []interface{}{45}})
	if err != 0 {
		t.Errorf("TestArrayIndexOfAndContains error: %v", err)
		return
	}
	if data["testFloatArray.*indexOf"] != float64(3) {
		t.Errorf("TestArrayIndexOfAndContains expected 3, but got: %v", data["testFloatArray.*indexOf"])
		return
	}
	if data["testFloatArray.*contains"] != false {
		t.Errorf("TestArrayIndexOfAndContains expected false, but got: %v", data["testFloatArray.*contains"])
	}
}

// Testing Array *indexOf and *contains
func TestMapKeyOfAndContains(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	// Map of Float32
	data, err := table.GetKeyData("Vokome", map[string]interface{}{"testFloatMap.*keyOf": []interface{}{3.45}, "testFloatMap.*contains": []interface{}{7}})
	if err != 0 {
		t.Errorf("TestMapKeyOfAndContains error: %v", err)
		return
	}
	if data["testFloatMap.*keyOf"] != "three point 45" {
		t.Errorf("TestMapKeyOfAndContains expected 'three point 45', but got: '%v'", data["testFloatMap.*keyOf"])
		return
	}
	if data["testFloatMap.*contains"] != false {
		t.Errorf("TestMapKeyOfAndContains expected false, but got: %v", data["testFloatMap.*contains"])
		return
	}
}

// Testing multi-layered array methods
func TestMultiArrayUpdateMethod(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"friends.*delete.*prepend.*append": []interface{}{[]interface{}{2,0}, []interface{}{map[string]interface{}{"login": "guest1227", "status": 0, "labels": map[string]interface{}{"nickname": "G7", "friendNum": 0}}}, []interface{}{map[string]interface{}{"login": "guest1229", "status": 0, "labels": map[string]interface{}{"nickname": "G9", "friendNum": 2}}}}})
	if (err != 0) {
		t.Errorf("TestAppendArray error: %v", err)
		return
	}
}

// Testing multi-layered map methods
func TestMultiMapUpdateMethod(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	err := table.UpdateKey("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"actions.*delete.*append": []interface{}{[]interface{}{"fek off"}, map[string]interface{}{"bloke": map[string]interface{}{"type": "insult", "id": 1}}}})
	if (err != 0) {
		t.Errorf("TestMultiMapMethod error: %v", err)
		return
	}
}



// Testing nested get queries
/*func TestUpdateWithNestedGetQuery(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	data, err := table.GetKeyData("guest" + strconv.Itoa(table.Size()), map[string]interface{}{"mmr.*get": []interface{}{map[string]interface{}}})
	if err != 0 {
		t.Errorf("TestComparisons error: %v", err)
		return
	}
}*/

// Must be last test!!
func TestLetFilesClose(t *testing.T) {
	if (!setupComplete) {
		t.Skip()
	}
	time.Sleep(4 * time.Second)
	if storage.GetNumOpenFiles() != 0 {
		t.Errorf("Error: Storage files did not close properly")
	}
}