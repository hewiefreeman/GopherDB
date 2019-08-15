package keystore

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"strconv"
	"encoding/json"
)

const (
	JsonEntryKey     = "k"
	JsonEntryData     = "d"
)

func makeJsonBytes(key string, data []interface{}, jBytes *[]byte) int {
	var jErr error
	*jBytes, jErr = json.Marshal(map[string]interface{}{
		JsonEntryKey: key,
		JsonEntryData: data,
	})
	if jErr != nil {
		return helpers.ErrorJsonEncoding
	}
	return 0
}

// Example JSON for new key query:
//
//     {"InsertKey": {"table": "tableName", "query": ["userName", "password", { *items that match schema* }]}}
//

// Insert creates a new KeystoreEntry in the Keystore, as long as one doesnt already exist
func (t *Keystore) Insert(key string, insertObj map[string]interface{}) int {
	// Name and password are required
	if len(key) == 0 {
		return helpers.ErrorKeyRequired
	}

	// Create entry
	e := KeystoreEntry{
		data: make([]interface{}, len(*(t.schema)), len(*(t.schema))),
	}

	uniqueVals := make(map[string]interface{})

	// Fill entry data with insertObj - Loop through schema to also check for required items
	for itemName, schemaItem := range *(t.schema) {
		// Item filter
		err := schema.ItemFilter(insertObj[itemName], nil, &e.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, false)
		if err != 0 {
			return err
		}
	}

	// Make JSON []byte for entry
	var jBytes []byte
	if !t.memOnly {
		if jErr := makeJsonBytes(key, e.data, &jBytes); jErr != 0 {
			return jErr
		}
	}

	// Lock table, check for duplicate entry
	maxEntries := t.maxEntries.Load().(uint64)
	t.eMux.Lock()
	if t.entries[key] != nil {
		t.eMux.Unlock()
		return helpers.ErrorKeyInUse
	} else if maxEntries > 0 && len(t.entries) >= int(maxEntries) {
		// Table is full
		return helpers.ErrorTableFull
	}
	t.uMux.Lock()
	// Check unique values
	for itemName, itemVal := range uniqueVals {
		if t.uniqueVals[itemName] != nil && t.uniqueVals[itemName][itemVal] {
			t.uMux.Unlock()
			t.eMux.Unlock()
			return helpers.ErrorUniqueValueInUse
		}/* else {
			// DISTRIBUTED CHECKS HERE !!!
		}*/
	}
	// Append jBytes to fileOn and get the persistIndex
	var lineOn uint16
	if !t.memOnly {
		var aErr int
		lineOn, aErr = storage.Insert(t.persistName + "/" + strconv.Itoa(int(t.fileOn)) + storage.FileTypeStorage, jBytes)
		if aErr != 0 {
			t.uMux.Unlock()
			t.eMux.Unlock()
			return aErr
		}
	}

	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if t.uniqueVals[itemName] == nil {
			t.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		t.uniqueVals[itemName][itemVal] = true
	}
	t.uMux.Unlock()

	//
	e.persistIndex = lineOn
	e.persistFile = t.fileOn

	// Increase fileOn when the index has reached or surpassed partitionMax
	if e.persistIndex >= t.partitionMax.Load().(uint16) {
		t.fileOn++
	}

	// Remove data from memory if dataOnDrive is true
	if t.dataOnDrive {
		e.data = nil
	}

	// Insert item
	t.entries[key] = &e
	t.eMux.Unlock()

	return 0
}

// Example JSON for get query:
//
//     {"GetUserData": {"table": "tableName", "query": ["userName", "password"]}}
//

// Get
func (t *Keystore) GetData(key string, items []string) (map[string]interface{}, int) {
	e, err := t.Get(key)
	if err != 0 {
		return nil, err
	}

	var data []interface{}

	// Get entry data
	if t.dataOnDrive {
		var dErr int
		data, dErr = t.dataFromDrive(t.persistName + "/" + strconv.Itoa(int(e.persistFile)) + storage.FileTypeStorage, e.persistIndex)
		if dErr != 0 {
			return nil, dErr
		}
	} else {
		e.mux.Lock()
		data = append([]interface{}{}, e.data...)
		e.mux.Unlock()
	}

	// Check for specific items to get
	m := make(map[string]interface{})
	if items != nil {
		for _, itemName := range items {
			siName, itemMethods := schema.GetQueryItemMethods(itemName)
			//
			si := (*(t.schema))[siName]
			if si == nil {
				return nil, helpers.ErrorInvalidItem
			}
			// Item filter
			var i interface{}
			err := schema.ItemFilter(data[si.DataIndex()], itemMethods, &i, nil, si, nil, true)
			if err != 0 {
				return nil, err
			}
			m[itemName] = i
		}
	} else {
		for itemName, si := range *(t.schema) {
			// Item filter
			var i interface{}
			err := schema.ItemFilter(data[si.DataIndex()], nil, &i, nil, si, nil, true)
			if err != 0 {
				return nil, err
			}
			m[itemName] = i
		}
	}
	return m, 0
}

func (t *Keystore) dataFromDrive(file string, index uint16) ([]interface{}, int) {
	// Read bytes from file
	bytes, rErr := storage.Read(file, index)
	if rErr != 0 {
		return nil, rErr
	}
	jMap := make(map[string]interface{})
	jErr := json.Unmarshal(bytes, &jMap)
	if jErr != nil {
		return nil, helpers.ErrorJsonDecoding
	}
	if jMap[JsonEntryData] == nil || len(jMap[JsonEntryData].([]interface{})) != len(*(t.schema)) {
		return nil, helpers.ErrorJsonDataFormat
	}
	return jMap[JsonEntryData].([]interface{}), 0
}

// Example JSON for update query:
//
//  Changing a string:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"email": "differentemail@yahoo.com"}]}}
//
//  Arithmetic on a number type:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"mmr.*add": [0.5]}]}} // can also be "*sub", "*mul", "*div", or "*mod"
//
//  Updating an item inside an Array:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"friends.0": {"name": "Joe", "status": 1}}]}}
//
//  Append item(s) to an Array or Map:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"friends.*append": [{"name": "Joe", "status": 1}]}]}}
//
//  Prepend item(s) to an Array:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"friends.*prepend": [{"name": "Joe", "status": 1}]}]}}
//
//  Append item(s) to certain position in an Array:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"friends.*append[3]": [{"name": "Joe", "status": 1}]}]}}
//
//  Delete item(s) in an Array or Map:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"friends.*delete": [0]}]}}
//
//  Changing an item in an Object (at an Array index or Map item):
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"friends.0.status": 2}]}}
//
//  Set Time item to current database time:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"timeStamp": "*now"}]}}
//
//  Set Time item manually:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"timeStamp": "5:23AM"}]}}
//

// Update
func (t *Keystore) UpdateData(key string, updateObj map[string]interface{}) int {
	if updateObj == nil || len(updateObj) == 0 {
		return helpers.ErrorQueryInvalidFormat
	}

	e, err := t.Get(key)
	if err != 0 {
		return err
	}

	var data []interface{}

	// Get entry data
	if t.dataOnDrive {
		var dErr int
		data, dErr = t.dataFromDrive(t.persistName + "/" + strconv.Itoa(int(e.persistFile)) + storage.FileTypeStorage, e.persistIndex)
		if dErr != 0 {
			return dErr
		}
		e.mux.Lock()
	} else {
		e.mux.Lock()
		data = append([]interface{}{}, e.data...)
	}

	uniqueVals := make(map[string]interface{})
	// Iterate through updateObj
	for updateName, updateItem := range updateObj {
		var itemMethods []string
		updateName, itemMethods = schema.GetQueryItemMethods(updateName)

		// Check if valid schema item
		schemaItem := (*(*t).schema)[updateName]
		if schemaItem == nil {
			e.mux.Unlock()
			return helpers.ErrorSchemaInvalid
		}

		itemBefore := data[schemaItem.DataIndex()]

		// Item filter
		err := schema.ItemFilter(updateItem, itemMethods, &data[schemaItem.DataIndex()], itemBefore, schemaItem, &uniqueVals, false)
		if err != 0 {
			e.mux.Unlock()
			return err
		}
	}

	// Make JSON []byte for entry
	var jBytes []byte
	if !t.memOnly {
		if jErr := makeJsonBytes(key, data, &jBytes); jErr != 0 {
			return jErr
		}
	}
	t.uMux.Lock()
	// Check unique values
	for itemName, itemVal := range uniqueVals {
		// Local unique check
		if t.uniqueVals[itemName] != nil && t.uniqueVals[itemName][itemVal] {
			t.uMux.Unlock()
			e.mux.Unlock()
			return helpers.ErrorUniqueValueInUse
		}

		// DISTRIBUTED UNIQUE CHECKS HERE !!!
	}

	// Update entry on disk with jBytes
	if !t.memOnly {
		uErr := storage.Update(t.persistName + "/" + strconv.Itoa(int(e.persistFile)) + storage.FileTypeStorage, e.persistIndex, jBytes)
		if uErr != 0 {
			t.uMux.Unlock()
			e.mux.Unlock()
			return uErr
		}
	}

	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if t.uniqueVals[itemName] == nil {
			t.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		t.uniqueVals[itemName][itemVal] = true
	}
	t.uMux.Unlock()

	//
	if !t.dataOnDrive {
		e.data = data
	}
	e.mux.Unlock()

	return 0
}

// Delete
func (t *Keystore) Delete(key string) int {
	ue, err := t.Get(key)
	if err != 0 {
		return err
	}

	var data []interface{}

	// Get entry data
	if t.dataOnDrive {
		var dErr int
		data, dErr = t.dataFromDrive(t.persistName + "/" + strconv.Itoa(int(ue.persistFile)) + storage.FileTypeStorage, ue.persistIndex)
		if dErr != 0 {
			return dErr
		}
		ue.mux.Lock()
	} else {
		ue.mux.Lock()
		data = append([]interface{}{}, ue.data...)
	}

	t.uMux.Lock()
	uItems := []string{}
	schema.GetUniqueItems(t.schema, &uItems, "")
	for _, itemName := range uItems {
		// Get entry's unique value for this unique item
		siName, itemMethods := schema.GetQueryItemMethods(itemName)
		//
		si := (*(t.schema))[siName]
		if si == nil {
			ue.mux.Unlock()
			t.uMux.Unlock()
			return helpers.ErrorUnexpected
		}
		// Make filter
		var i interface{}
		err := schema.ItemFilter(data[si.DataIndex()], itemMethods, &i, nil, si, nil, true)
		if err != 0 {
			ue.mux.Unlock()
			t.uMux.Unlock()
			return helpers.ErrorUnexpected
		}
		delete(t.uniqueVals[itemName], i)
	}
	ue.mux.Unlock()
	t.uMux.Unlock()

	// Update entry on disk with []byte{}
	if !t.memOnly {
		uErr := storage.Update(t.persistName + "/" + strconv.Itoa(int(ue.persistFile)) + storage.FileTypeStorage, ue.persistIndex, []byte{})
		if uErr != 0 {
			return uErr
		}
	}

	t.eMux.Lock()
	// Delete entry
	delete(t.entries, key)
	t.eMux.Unlock()

	//
	return 0
}