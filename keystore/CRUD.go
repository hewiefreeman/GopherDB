package keystore

import (
	"encoding/json"
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"strconv"
	"strings"
)

type jsonEntry struct {
	K string
	D []interface{}
}

func makeJsonBytes(key string, data []interface{}, jBytes *[]byte) int {
	var jErr error
	if *jBytes, jErr = helpers.Fjson.Marshal(jsonEntry{
		K: key,
		D: data,
	}); jErr != nil {
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
func (k *Keystore) InsertKey(key string, insertObj map[string]interface{}) (*keystoreEntry, helpers.Error) {
	// Key is required
	if len(key) == 0 {
		return nil, helpers.+(helpers.ErrorKeyRequired, k.name)
	} else if strings.ContainsAny(key, ".*\t\n\r") {
		return nil, helpers.NewError(helpers.ErrorInvalidKeyCharacters, key)
	}

	// Create entry
	e := keystoreEntry{
		data: make([]interface{}, len(k.schema), len(k.schema)),
	}

	uniqueVals := make(map[string]interface{})

	// Fill entry data with insertObj - Loop through schema to also check for required items
	for itemName, schemaItem := range k.schema {
		// Item filter
		err := schema.ItemFilter(insertObj[itemName], nil, &e.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, k.EncryptCost(), false, false)
		if err != 0 {
			return nil, helpers.NewError(err, itemName)
		}
	}

	// Make JSON []byte for entry
	var jBytes []byte
	if !k.memOnly {
		if jErr := makeJsonBytes(key, e.data, &jBytes); jErr != 0 {
			return nil, helpers.NewError(jErr, key)
		}
	}

	// Lock table, check for duplicate entry
	maxEntries := k.maxEntries.Load().(uint64)
	k.eMux.Lock()
	if k.entries[key] != nil {
		k.eMux.Unlock()
		return nil, helpers.NewError(helpers.ErrorKeyInUse, key)
	} else if maxEntries > 0 && len(k.entries) >= int(maxEntries) {
		// Table is full
		return nil, helpers.NewError(helpers.ErrorTableFull, k.name)
	}
	k.uMux.Lock()
	// Check unique values
	for itemName, itemVal := range uniqueVals {
		if k.uniqueVals[itemName] != nil && k.uniqueVals[itemName][itemVal] {
			k.uMux.Unlock()
			k.eMux.Unlock()
			return nil, helpers.NewError(helpers.ErrorUniqueValueDuplicate, itemName)
		} /* else {
			// DISTRIBUTED CHECKS HERE !!!
		}*/
	}
	// Append jBytes to fileOn and get the persistIndex
	var lineOn uint16
	if !k.memOnly {
		var aErr int
		lineOn, aErr = storage.Insert(dataFolderPrefix+k.name+"/"+strconv.Itoa(int(k.fileOn))+helpers.FileTypeStorage, jBytes)
		if aErr != 0 {
			k.uMux.Unlock()
			k.eMux.Unlock()
			return nil, helpers.NewError(aErr, key)
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
			Name:         k.name,
			Schema:       k.schema.MakeConfig(),
			FileOn:       k.fileOn,
			DataOnDrive:  k.dataOnDrive,
			MemOnly:      k.memOnly,
			PartitionMax: k.partitionMax.Load().(uint16),
			EncryptCost:  k.encryptCost.Load().(int),
			MaxEntries:   k.maxEntries.Load().(uint64),
		})
	}

	// Remove data from memory if dataOnDrive is true
	if k.dataOnDrive {
		e.data = nil
	}

	// Insert item
	k.entries[key] = &e
	k.eMux.Unlock()

	return &e, helpers.Error{}
}

// Example JSON for get query:
//
//     ["Get", "tableName", "key", { *items that match schema* }]
//

// Get
func (k *Keystore) GetKey(key string, items map[string]interface{}) (map[string]interface{}, helpers.Error) {
	// Get entry
	e, err := k.Get(key)
	if err != 0 {
		return nil, helpers.NewError(err, k.name + " > " + key)
	}

	var data []interface{}

	// Get entry data
	if k.dataOnDrive {
		data, err = k.dataFromDrive(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage, e.persistIndex)
		if err != 0 {
			return nil, helpers.NewError(err, dataFolderPrefix + k.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage)
		}
	} else {
		e.mux.Lock()
		data = append([]interface{}{}, e.data...)
		e.mux.Unlock()
	}

	// Check for specific items to get
	if items != nil && len(items) > 0 {
		// Items were found in query, get requested items
		for itemName, methodParams := range items {
			siName, itemMethods := schema.GetQueryItemMethods(itemName)
			//
			si := (k.schema)[siName]
			if !si.QuickValidate() {
				return nil, helpers.NewError(helpers.ErrorInvalidItem, itemName)
			}
			// Item filter
			var i interface{}
			err = schema.ItemFilter(methodParams, itemMethods, &i, data[si.DataIndex()], si, nil, k.EncryptCost(), true, false)
			if err != 0 {
				return nil, helpers.NewError(err, itemName)
			}
			items[itemName] = i
		}
	} else {
		// No items specified, get all items
		items = make(map[string]interface{}, len(k.schema))
		for itemName, si := range k.schema {
			// Item filter
			var i interface{}
			err = schema.ItemFilter(nil, nil, &i, data[si.DataIndex()], si, nil, k.EncryptCost(), true, false)
			if err != 0 {
				return nil, helpers.NewError(err, itemName)
			}
			items[itemName] = i
		}
	}
	return items, helpers.Error{}
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
func (k *Keystore) UpdateKey(key string, updateObj map[string]interface{}) helpers.Error {
	if updateObj == nil || len(updateObj) == 0 {
		return helpers.NewError(helpers.ErrorQueryInvalidFormat, k.name + " > " + key)
	}

	e, err := k.Get(key)
	if err != 0 {
		return helpers.NewError(err, k.name + " > " + key)
	}

	var data []interface{}

	// Get entry data
	if k.dataOnDrive {
		data, err = k.dataFromDrive(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage, e.persistIndex)
		if err != 0 {
			return helpers.NewError(err, dataFolderPrefix + k.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage)
		}
		e.mux.Lock()
	} else {
		e.mux.Lock()
		data = append([]interface{}{}, e.data...)
	}

	uniqueVals := make(map[string]interface{})
	uniqueValsBefore := make(map[string]interface{})
	// Iterate through updateObj
	for updateName, updateItem := range updateObj {
		var itemMethods []string
		var uName string
		uName, itemMethods = schema.GetQueryItemMethods(updateName)

		// Check if valid schema item
		schemaItem := k.schema[uName]
		if !schemaItem.QuickValidate() {
			e.mux.Unlock()
			return helpers.NewError(helpers.ErrorSchemaInvalid, updateName)
		}

		itemBefore := data[schemaItem.DataIndex()]

		// Item filter
		err = schema.ItemFilter(updateItem, itemMethods, &data[schemaItem.DataIndex()], itemBefore, schemaItem, &uniqueVals, k.EncryptCost(), false, false)
		if err != 0 {
			e.mux.Unlock()
			return helpers.NewError(err, updateName)
		}
		// Check for changed unique value to remove old value from table's uniqueVals
		if uniqueVals[uName] != nil && data[schemaItem.DataIndex()] != itemBefore {
			uniqueValsBefore[uName] = itemBefore
		}
	}

	// Make JSON []byte for entry
	var jBytes []byte
	if !k.memOnly {
		if jErr := makeJsonBytes(key, data, &jBytes); jErr != 0 {
			return helpers.NewError(jErr, k.name + " > " + key)
		}
	}
	k.uMux.Lock()
	// Check unique values
	for itemName, itemVal := range uniqueVals {
		// Local unique check
		if k.uniqueVals[itemName] != nil && k.uniqueVals[itemName][itemVal] {
			k.uMux.Unlock()
			e.mux.Unlock()
			return helpers.NewError(helpers.ErrorUniqueValueDuplicate, itemName)
		}
		// DISTRIBUTED UNIQUE CHECKS HERE !!!
	}

	// Update entry on disk with jBytes
	if !k.memOnly {
		err = storage.Update(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage, e.persistIndex, jBytes)
		if err != 0 {
			k.uMux.Unlock()
			e.mux.Unlock()
			return helpers.NewError(err, dataFolderPrefix + k.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage)
		}
	}

	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if k.uniqueVals[itemName] == nil {
			k.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		k.uniqueVals[itemName][itemVal] = true

		// Remove old unique values
		if uniqueValsBefore[itemName] != nil {
			delete(k.uniqueVals[itemName], uniqueValsBefore[itemName])
		}
	}
	k.uMux.Unlock()

	//
	if !k.dataOnDrive {
		e.data = data
	}
	e.mux.Unlock()

	return helpers.Error{}
}

// UpsertKey
func (k *Keystore) UpsertKey(key string, upsertObj map[string]interface{}) (*keystoreEntry, helpers.Error) {
	// Key is required
	if len(key) == 0 {
		return nil, helpers.NewError(helpers.ErrorKeyRequired, k.name + " > " + key)
	}

	ke, err := k.Get(key)
	if err != 0 {
		// Insert
		return k.InsertKey(key, upsertObj)
	}
	return ke, k.UpdateKey(key, upsertObj)
}

// Delete
func (k *Keystore) DeleteKey(key string) helpers.Error {
	ue, err := k.Get(key)
	if err != 0 {
		return helpers.NewError(err, k.name + " > " + key)
	}

	var data []interface{}

	// Get entry data
	if k.dataOnDrive {
		data, err = k.dataFromDrive(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex)
		if err != 0 {
			return helpers.NewError(err, dataFolderPrefix + k.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage)
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
			return helpers.NewError(helpers.ErrorUnexpected, k.name + ": Invalid schema while deleting Keystore")
		}
		// Make get filter
		var i interface{}
		err := schema.ItemFilter(nil, itemMethods, &i, data[si.DataIndex()], si, nil, k.EncryptCost(), true, false)
		if err != 0 {
			ue.mux.Unlock()
			k.uMux.Unlock()
			return helpers.NewError(helpers.ErrorUnexpected, k.name ": Item filter failed while deleting Keystore")
		}
		delete(k.uniqueVals[itemName], i)
	}
	ue.mux.Unlock()
	k.uMux.Unlock()

	// Update entry on disk with []byte{}
	if !k.memOnly {
		err = storage.Update(dataFolderPrefix + k.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex, []byte{})
		if err != 0 {
			return helpers.NewError(err, dataFolderPrefix + k.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage)
		}
	}

	k.eMux.Lock()
	// Delete entry
	delete(k.entries, key)
	k.eMux.Unlock()

	//
	return helpers.Error{}
}

// Restores a key from a config file - NOT concurrently safe on it's own! Must lock Keystore before-hand.
func (k *Keystore) restoreKey(key string, data []interface{}, fileOn uint32, lineOn uint16) int {
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
		err := schema.ItemFilter(data[schemaItem.DataIndex()], nil, &e.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, 0, false, true)
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
