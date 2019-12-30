package keystore

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"strconv"
	"strings"
	"encoding/json"
)

type jsonEntry struct {
	K string
	D []interface{}
}

func makeJsonBytes(key string, data []interface{}, jBytes *[]byte) int {
	var jErr error
	*jBytes, jErr = json.Marshal(jsonEntry{
		K: key,
		D: data,
	})
	if jErr != nil {
		return helpers.ErrorJsonEncoding
	}
	return 0
}

// Examples of nested Get queries
/*

 - Get index for deleting specific friend
	["Update", "test", "Mary", {"friends.*delete.*get": [["test", "Mary", {"friends.*indexOf": ["Vokome"]}]]}]

*/

// Example JSON for new key query:
//
//     ["Insert", "tableName", "key", { *items that match schema* }]
//

// Insert creates a new keystoreEntry in the Keystore, as long as one doesnt already exist
func (k *Keystore) InsertKey(key string, insertObj map[string]interface{}) (*keystoreEntry, int) {
	// Key is required
	if len(key) == 0 {
		return nil, helpers.ErrorKeyRequired
	} else if strings.ContainsAny(key, ".*\t\n\r"){
		return nil, helpers.ErrorInvalidKeyCharacters
	}

	// Create entry
	e := keystoreEntry{
		data: make([]interface{}, len(k.schema), len(k.schema)),
	}

	uniqueVals := make(map[string]interface{})

	// Fill entry data with insertObj - Loop through schema to also check for required items
	for itemName, schemaItem := range k.schema {
		// Item filter
		err := schema.ItemFilter(insertObj[itemName], nil, &e.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, false, false)
		if err != 0 {
			return nil, err
		}
	}

	// Make JSON []byte for entry
	var jBytes []byte
	if !k.memOnly {
		if jErr := makeJsonBytes(key, e.data, &jBytes); jErr != 0 {
			return nil, jErr
		}
	}

	// Lock table, check for duplicate entry
	maxEntries := k.maxEntries.Load().(uint64)
	k.eMux.Lock()
	if k.entries[key] != nil {
		k.eMux.Unlock()
		return nil, helpers.ErrorKeyInUse
	} else if maxEntries > 0 && len(k.entries) >= int(maxEntries) {
		// Table is full
		return nil, helpers.ErrorTableFull
	}
	k.uMux.Lock()
	// Check unique values
	for itemName, itemVal := range uniqueVals {
		if k.uniqueVals[itemName] != nil && k.uniqueVals[itemName][itemVal] {
			k.uMux.Unlock()
			k.eMux.Unlock()
			return nil, helpers.ErrorUniqueValueDuplicate
		}/* else {
			// DISTRIBUTED CHECKS HERE !!!
		}*/
	}
	// Append jBytes to fileOn and get the persistIndex
	var lineOn uint16
	if !k.memOnly {
		var aErr int
		lineOn, aErr = storage.Insert(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(k.fileOn)) + helpers.FileTypeStorage, jBytes)
		if aErr != 0 {
			k.uMux.Unlock()
			k.eMux.Unlock()
			return nil, aErr
		}
	}

	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if k.uniqueVals[itemName] == nil {
			k.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		k.uniqueVals[itemName][itemVal] = true
	}
	k.uMux.Unlock()

	//
	e.persistIndex = lineOn
	e.persistFile = k.fileOn

	// Increase fileOn when the index has reached or surpassed partitionMax
	if e.persistIndex >= k.partitionMax.Load().(uint16) {
		k.fileOn++
		writeConfigFile(k.configFile, keystoreConfig{
			Name: k.name,
			Schema: k.schema.MakeConfig(),
			FileOn: k.fileOn,
			DataOnDrive: k.dataOnDrive,
			MemOnly: k.memOnly,
			PartitionMax: k.partitionMax.Load().(uint16),
			EncryptCost: k.encryptCost.Load().(int),
			MaxEntries: k.maxEntries.Load().(uint64),
		})
	}

	// Remove data from memory if dataOnDrive is true
	if k.dataOnDrive {
		e.data = nil
	}

	// Insert item
	k.entries[key] = &e
	k.eMux.Unlock()

	return &e, 0
}

// Example JSON for get query:
//
//     ["Get", "tableName", "key", { *items that match schema* }]
//

// Get
func (k *Keystore) GetKeyData(key string, items map[string]interface{}) (map[string]interface{}, int) {
	// Get entry
	e, err := k.Get(key)
	if err != 0 {
		return nil, err
	}

	var data []interface{}

	// Get entry data
	if k.dataOnDrive {
		var dErr int
		data, dErr = k.dataFromDrive(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage, e.persistIndex)
		if dErr != 0 {
			return nil, dErr
		}
	} else {
		e.mux.Lock()
		data = append([]interface{}{}, e.data...)
		e.mux.Unlock()
	}

	// Check for specific items to get
	if items != nil && len(items) > 0 {
		for itemName, methodParams := range items {
			siName, itemMethods := schema.GetQueryItemMethods(itemName)
			//
			si := (k.schema)[siName]
			if !si.QuickValidate() {
				return nil, helpers.ErrorInvalidItem
			}
			// Item filter
			var i interface{}
			err = schema.ItemFilter(methodParams, itemMethods, &i, data[si.DataIndex()], si, nil, true, false)
			if err != 0 {
				return nil, err
			}
			items[itemName] = i
		}
		return items, 0
	} else {
		items = make(map[string]interface{})
		for itemName, si := range k.schema {
			// Item filter
			var i interface{}
			err = schema.ItemFilter(nil, nil, &i, data[si.DataIndex()], si, nil, true, false)
			if err != 0 {
				return nil, err
			}
			items[itemName] = i
		}

	}
	return items, 0
}

func (k *Keystore) dataFromDrive(file string, index uint16) ([]interface{}, int) {
	// Read bytes from file
	bytes, rErr := storage.Read(file, index)
	if rErr != 0 {
		return nil, rErr
	}
	var jEntry jsonEntry
	jErr := json.Unmarshal(bytes, &jEntry)
	if jErr != nil {
		return nil, helpers.ErrorJsonDecoding
	}
	if jEntry.D == nil || len(jEntry.D) == 0 {
		return nil, helpers.ErrorJsonDecoding
	}
	return jEntry.D, 0
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
func (k *Keystore) UpdateKey(key string, updateObj map[string]interface{}) int {
	if updateObj == nil || len(updateObj) == 0 {
		return helpers.ErrorQueryInvalidFormat
	}

	e, err := k.Get(key)
	if err != 0 {
		return err
	}

	var data []interface{}

	// Get entry data
	if k.dataOnDrive {
		var dErr int
		data, dErr = k.dataFromDrive(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage, e.persistIndex)
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
		schemaItem := k.schema[updateName]
		if !schemaItem.QuickValidate() {
			e.mux.Unlock()
			return helpers.ErrorSchemaInvalid
		}

		//itemBefore := data[schemaItem.DataIndex()]

		// Item filter
		err := schema.ItemFilter(updateItem, itemMethods, &data[schemaItem.DataIndex()], data[schemaItem.DataIndex()], schemaItem, &uniqueVals, false, false)
		if err != 0 {
			e.mux.Unlock()
			return err
		}
	}

	// Make JSON []byte for entry
	var jBytes []byte
	if !k.memOnly {
		if jErr := makeJsonBytes(key, data, &jBytes); jErr != 0 {
			return jErr
		}
	}
	k.uMux.Lock()
	// Check unique values
	for itemName, itemVal := range uniqueVals {
		// Local unique check
		if k.uniqueVals[itemName] != nil && k.uniqueVals[itemName][itemVal] {
			k.uMux.Unlock()
			e.mux.Unlock()
			return helpers.ErrorUniqueValueDuplicate
		}

		// DISTRIBUTED UNIQUE CHECKS HERE !!!
	}

	// Update entry on disk with jBytes
	if !k.memOnly {
		uErr := storage.Update(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage, e.persistIndex, jBytes)
		if uErr != 0 {
			k.uMux.Unlock()
			e.mux.Unlock()
			return uErr
		}
	}

	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if k.uniqueVals[itemName] == nil {
			k.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		k.uniqueVals[itemName][itemVal] = true
	}
	k.uMux.Unlock()

	//
	if !k.dataOnDrive {
		e.data = data
	}
	e.mux.Unlock()

	return 0
}

// UpsertKey
func (k *Keystore) UpsertKey(key string, upsertObj map[string]interface{}) (*keystoreEntry, int) {
	// Key is required
	if len(key) == 0 {
		return nil, helpers.ErrorKeyRequired
	}

	ke, err := k.Get(key)
	if err != 0 {
		// Insert
		return k.InsertKey(key, upsertObj)
	}
	return ke, k.UpdateKey(key, upsertObj)
}

// Delete
func (k *Keystore) DeleteKey(key string) int {
	ue, err := k.Get(key)
	if err != 0 {
		return err
	}

	var data []interface{}

	// Get entry data
	if k.dataOnDrive {
		var dErr int
		data, dErr = k.dataFromDrive(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex)
		if dErr != 0 {
			return dErr
		}
		ue.mux.Lock()
	} else {
		ue.mux.Lock()
		data = append([]interface{}{}, ue.data...)
	}

	k.uMux.Lock()
	uItems := []string{}
	schema.GetUniqueItems(k.schema, &uItems, "")
	for _, itemName := range uItems {
		// Get entry's unique value for this unique item
		siName, itemMethods := schema.GetQueryItemMethods(itemName)
		//
		si := k.schema[siName]
		if !si.QuickValidate() {
			ue.mux.Unlock()
			k.uMux.Unlock()
			return helpers.ErrorUnexpected
		}
		// Make filter
		var i interface{}
		err := schema.ItemFilter(data[si.DataIndex()], itemMethods, &i, nil, si, nil, true, false)
		if err != 0 {
			ue.mux.Unlock()
			k.uMux.Unlock()
			return helpers.ErrorUnexpected
		}
		delete(k.uniqueVals[itemName], i)
	}
	ue.mux.Unlock()
	k.uMux.Unlock()

	// Update entry on disk with []byte{}
	if !k.memOnly {
		uErr := storage.Update(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex, []byte{})
		if uErr != 0 {
			return uErr
		}
	}

	k.eMux.Lock()
	// Delete entry
	delete(k.entries, key)
	k.eMux.Unlock()

	//
	return 0
}

// Restores a key from a config file - NOT concurrently safe on it's own! Must lock Keystore before-hand.
func (k *Keystore) restoreKey(key string, data []interface{}, fileOn uint16, lineOn uint16) int {
	// Check for duplicate entry
	if k.entries[key] != nil {
		return helpers.ErrorKeyInUse
	}

	// Create entry
	e := keystoreEntry{
		data: make([]interface{}, len(k.schema), len(k.schema)),
	}

	uniqueVals := make(map[string]interface{})

	// Fill entry data with data
	for _, schemaItem := range k.schema {
		if int(schemaItem.DataIndex()) > len(data)-1 {
			return helpers.ErrorRestoreItemSchema
		}

		// Item filter
		err := schema.ItemFilter(data[schemaItem.DataIndex()], nil, &e.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, false, true)
		if err != 0 {
			return err
		}
	}

	// Check unique values
	for itemName, itemVal := range uniqueVals {
		// Local unique check
		if k.uniqueVals[itemName] != nil && k.uniqueVals[itemName][itemVal] {
			return helpers.ErrorUniqueValueDuplicate
		}

		// DISTRIBUTED UNIQUE CHECKS HERE !!!
	}

	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if k.uniqueVals[itemName] == nil {
			k.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		k.uniqueVals[itemName][itemVal] = true
	}

	//
	e.persistIndex = lineOn
	e.persistFile = fileOn

	// Remove data from memory if dataOnDrive is true
	if k.dataOnDrive {
		e.data = nil
	}

	// Insert item
	k.entries[key] = &e
	return 0
}