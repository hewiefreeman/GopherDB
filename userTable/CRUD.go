package userTable

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"strconv"
	"encoding/json"
)

const (
	JsonEntryName     = "n"
	JsonEntryPassword = "p"
	JsonEntryData     = "d"
)

func makeJsonBytes(name string, password []byte, data []interface{}) ([]byte, int) {
	jMap := make(map[string]interface{})
	jMap[JsonEntryName] = name
	jMap[JsonEntryPassword] = password
	jMap[JsonEntryData] = data
	jBytes, jErr := json.Marshal(jMap)
	if jErr != nil {
		return nil, helpers.ErrorJsonEncoding
	}
	return jBytes, 0
}

// Example JSON for new user query:
//
//     {"NewUser": {"table": "tableName", "query": ["userName", "password", { *items that match schema* }]}}
//

// NewUser creates a new UserTableEntry in the UserTable
func (t *UserTable) NewUser(name string, password string, insertObj map[string]interface{}) int {
	t.sMux.Lock()
	eCost := t.encryptCost
	minPass := t.minPassword
	maxEntries := t.maxEntries
	t.sMux.Unlock()

	// Name and password are required
	if len(name) == 0 {
		return helpers.ErrorNameRequired
	} else if len(password) < int(minPass) {
		return helpers.ErrorPasswordLength
	} else if maxEntries > 0 && t.Size() >= int(maxEntries) {
		return helpers.ErrorTableFull
	}

	// Get the current file number
	t.pMux.Lock()
	fileOn := t.fileOn
	t.pMux.Unlock()

	// Create entry
	ute := UserTableEntry{
		persistFile:  fileOn,
		data:         make([]interface{}, len(*(t.schema)), len(*(t.schema))),
	}

	// Alternative login name
	altLogin := ""
	uniqueVals := make(map[string]interface{})

	// Fill entry data with insertObj - Loop through schema to also check for required items
	for itemName, schemaItem := range *(t.schema) {
		// Item filter
		err := schema.ItemFilter(insertObj[itemName], nil, &ute.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, false)
		if err != 0 {
			return err
		}

		if itemName == t.altLoginItem {
			altLogin = ute.data[schemaItem.DataIndex()].(string)
		}
	}

	// Encrypt password and store in entry
	ePass, ePassErr := helpers.EncryptString(password, eCost)
	if ePassErr != nil {
		return helpers.ErrorPasswordEncryption
	}
	ute.password = ePass

	// Make JSON []byte for entry
	jBytes, jErr := makeJsonBytes(name, ePass, ute.data)
	if jErr != 0 {
		return jErr
	}

	// Lock table, check for duplicate entry
	t.eMux.Lock()
	if t.entries[name] != nil {
		t.eMux.Unlock()
		return helpers.ErrorEntryNameInUse
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
	retChan := make(chan interface{})
	qErr := storage.QueueFileAction(storage.FileActionInsert, []interface{}{t.persistName + "/" + strconv.Itoa(int(fileOn)) + storage.FileTypeStorage, jBytes}, retChan)
	if qErr != 0 {
		close(retChan)
		t.uMux.Unlock()
		t.eMux.Unlock()
		return qErr
	}
	qRes := <-retChan
	close(retChan)
	if qRes.([]interface{})[1] != nil {
		t.uMux.Unlock()
		t.eMux.Unlock()
		return helpers.ErrorFileAppend
	}
	ute.persistIndex = qRes.([]interface{})[0].(uint16)

	// Increase fileOn when the index has reached or surpassed partitionMax
	if ute.persistIndex >= t.partitionMax {
		t.pMux.Lock()
		t.fileOn++
		t.pMux.Unlock()
	}

	// Remove data from memory if dataOnDrive is true
	if t.dataOnDrive {
		ute.data = nil
	}

	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if t.uniqueVals[itemName] == nil {
			t.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		t.uniqueVals[itemName][itemVal] = true
	}
	if altLogin != "" {
		t.altLogins[altLogin] = &ute
	}
	t.uMux.Unlock()
	// Insert item
	t.entries[name] = &ute
	t.eMux.Unlock()

	return 0
}

// Example JSON for get query:
//
//     {"GetUserData": {"table": "tableName", "query": ["userName", "password"]}}
//

// GetUserData
func (t *UserTable) GetUserData(userName string, password string, items []string) (map[string]interface{}, int) {
	e, err := t.Get(userName, password)
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

	// Check for specific items/methods to get
	m := make(map[string]interface{})
	if len(items) > 0 {
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

func (t *UserTable) dataFromDrive(file string, index uint16) ([]interface{}, int) {
	retChan := make(chan interface{})
	qErr := storage.QueueFileAction(storage.FileActionRead, []interface{}{file, index}, retChan)
	if qErr != 0 {
		close(retChan)
		return nil, qErr
	}
	qRes := <-retChan
	close(retChan)
	if len(qRes.([]byte)) == 0 {
		return nil, helpers.ErrorFileRead
	}
	jMap := make(map[string]interface{})
	jErr := json.Unmarshal(qRes.([]byte), &jMap)
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

// UpdateUserData
func (t *UserTable) UpdateUserData(userName string, password string, updateObj map[string]interface{}) int {
	if len(updateObj) == 0 {
		return helpers.ErrorQueryInvalidFormat
	}

	e, err := t.Get(userName, password)
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
	altLoginBefore := ""
	altLoginAfter := ""
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

		if updateName == t.altLoginItem {
			altLoginBefore = itemBefore.(string)
			altLoginAfter = data[schemaItem.DataIndex()].(string)
		}
	}

	t.uMux.Lock()
	// Check unique values
	for itemName, itemVal := range uniqueVals {
		if t.uniqueVals[itemName] != nil && t.uniqueVals[itemName][itemVal] {
			t.uMux.Unlock()
			e.mux.Unlock()
			return helpers.ErrorUniqueValueInUse
		}/* else {
			// DISTRIBUTED CHECKS HERE !!!
		}*/
	}

	// Make JSON []byte for entry
	jBytes, jErr := makeJsonBytes(userName, e.password, data)
	if jErr != 0 {
		t.uMux.Unlock()
		e.mux.Unlock()
		return helpers.ErrorJsonEncoding
	}

	// Update entry on disk with jBytes
	retChan := make(chan interface{})
	qErr := storage.QueueFileAction(storage.FileActionUpdate, []interface{}{t.persistName + "/" + strconv.Itoa(int(e.persistFile)) + storage.FileTypeStorage, e.persistIndex, jBytes}, retChan)
	if qErr != 0 {
		close(retChan)
		e.mux.Unlock()
		t.uMux.Unlock()
		return qErr
	}
	qRes := <-retChan
	close(retChan)
	if qRes != nil {
		e.mux.Unlock()
		t.uMux.Unlock()
		return helpers.ErrorFileUpdate
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
	if altLoginBefore != "" {
		t.eMux.Lock()
		delete(t.altLogins, altLoginBefore)
		t.altLogins[altLoginAfter] = e
		t.eMux.Unlock()
	}
	//
	if !t.dataOnDrive {
		e.data = data
	}
	e.mux.Unlock()

	return 0
}

func (t *UserTable) ChangePassword(userName string, password string, newPassword string) int {
	t.sMux.Lock()
	eCost := t.encryptCost
	minPass := t.minPassword
	t.sMux.Unlock()

	if len(newPassword) < int(minPass) {
		return helpers.ErrorPasswordLength
	}

	ue, err := t.Get(userName, password)
	if err != 0 {
		return err
	}

	// Change password
	ePass, eErr := helpers.EncryptString(newPassword, eCost)
	if eErr != nil {
		return helpers.ErrorPasswordEncryption
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

	// Make JSON []byte for entry
	jBytes, jErr := makeJsonBytes(userName, ePass, data)
	if jErr != 0 {
		ue.mux.Unlock()
		return helpers.ErrorJsonEncoding
	}

	// Update entry on disk with jBytes
	retChan := make(chan interface{})
	qErr := storage.QueueFileAction(storage.FileActionUpdate, []interface{}{t.persistName + "/" + strconv.Itoa(int(ue.persistFile)) + storage.FileTypeStorage, ue.persistIndex, jBytes}, retChan)
	if qErr != 0 {
		close(retChan)
		ue.mux.Unlock()
		return qErr
	}
	qRes := <-retChan
	close(retChan)
	if qRes != nil {
		ue.mux.Unlock()
		return helpers.ErrorFileUpdate
	}

	ue.password = ePass
	ue.mux.Unlock()

	//
	return 0
}

func (t *UserTable) ResetPassword(userName string) int {
	t.sMux.Lock()
	eCost := t.encryptCost
	passResetLen := t.passResetLen
	t.sMux.Unlock()

	// Name and password are required
	if len(userName) == 0 {
		return helpers.ErrorNameRequired
	} else if t.emailItem == "" {
		return helpers.ErrorNoEmailItem
	}

	// Generate new password
	newPass, pErr := helpers.GenerateRandomBytes(int(passResetLen))
	if pErr != nil {
		return helpers.ErrorPasswordEncryption
	}

	// Send newPass to t.emailItem, do not proceed unless the email was a success !!!

	// Get entry
	t.eMux.Lock()
	ue := t.entries[userName]
	t.eMux.Unlock()

	var data []interface{}

	// Get entry data
	if t.dataOnDrive {
		var dErr int
		data, dErr = t.dataFromDrive(t.persistName + "/" + strconv.Itoa(int(ue.persistFile)) + storage.FileTypeStorage, ue.persistIndex)
		if dErr != 0 {
			return dErr
		}
	} else {
		ue.mux.Lock()
		data = append([]interface{}{}, ue.data...)
		ue.mux.Unlock()
	}

	if ue == nil && t.altLoginItem != "" {
		ue = t.altLogins[userName]
	}

	//
	if ue == nil {
		return helpers.ErrorInvalidNameOrPassword
	}
	// Change password
	ePass, eErr := helpers.EncryptString(string(newPass), eCost)
	if eErr != nil {
		return helpers.ErrorPasswordEncryption
	}

	// Delete auto-login hashes !!!

	// Make JSON []byte for entry
	jBytes, jErr := makeJsonBytes(userName, ePass, data)
	if jErr != 0 {
		return helpers.ErrorJsonEncoding
	}

	// Update entry on disk with jBytes
	retChan := make(chan interface{})
	qErr := storage.QueueFileAction(storage.FileActionUpdate, []interface{}{t.persistName + "/" + strconv.Itoa(int(ue.persistFile)) + storage.FileTypeStorage, ue.persistIndex, jBytes}, retChan)
	if qErr != 0 {
		close(retChan)
		return qErr
	}
	qRes := <-retChan
	close(retChan)
	if qRes != nil {
		return helpers.ErrorFileUpdate
	}

	ue.mux.Lock()
	ue.password = ePass
	ue.mux.Unlock()

	//
	return 0
}

func (t *UserTable) DeleteUser(userName string, password string) int {
	ue, err := t.Get(userName, password)
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
		if itemName == t.altLoginItem {
			t.eMux.Lock()
			delete(t.altLogins, i.(string))
			t.eMux.Unlock()
		}
		delete(t.uniqueVals[itemName], i)
	}
	ue.mux.Unlock()
	t.uMux.Unlock()

	// Update entry on disk with empty []byte
	retChan := make(chan interface{})
	qErr := storage.QueueFileAction(storage.FileActionUpdate, []interface{}{t.persistName + "/" + strconv.Itoa(int(ue.persistFile)) + storage.FileTypeStorage, ue.persistIndex, []byte{}}, retChan)
	if qErr != 0 {
		close(retChan)
		return qErr
	}
	qRes := <-retChan
	close(retChan)
	if qRes != nil {
		return helpers.ErrorFileUpdate
	}

	t.eMux.Lock()
	// Delete entry
	delete(t.entries, userName)
	t.eMux.Unlock()

	//
	return 0
}