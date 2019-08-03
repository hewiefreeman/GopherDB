package authtable

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"strconv"
	"encoding/json"
	"fmt"
)

const (
	JsonEntryName     = "n"
	JsonEntryPassword = "p"
	JsonEntryData     = "d"
)

func makeJsonBytes(name string, password []byte, data []interface{}, jBytes *[]byte) int {
	var jErr error
	*jBytes, jErr = json.Marshal(map[string]interface{}{
		JsonEntryName: name,
		JsonEntryPassword: password,
		JsonEntryData: data,
	})
	if jErr != nil {
		return helpers.ErrorJsonEncoding
	}
	return 0
}

// Example JSON for new user query:
//
//     {"NewUser": {"table": "tableName", "query": ["userName", "password", { *items that match schema* }]}}
//

// NewUser creates a new AuthTableEntry in the AuthTable
func (t *AuthTable) NewUser(name string, password string, insertObj map[string]interface{}) int {
	// Name and password are required
	if len(name) == 0 {
		return helpers.ErrorNameRequired
	} else if len(password) < int(t.minPassword.Load().(uint8)) {
		return helpers.ErrorPasswordLength
	}

	// Create entry
	ute := AuthTableEntry{
		data: make([]interface{}, len(*(t.schema)), len(*(t.schema))),
	}

	// Alternative login name
	altLogin := ""
	altLoginItem := t.altLoginItem.Load().(string)
	uniqueVals := make(map[string]interface{})

	// Fill entry data with insertObj - Loop through schema to also check for required items
	for itemName, schemaItem := range *(t.schema) {
		// Item filter
		err := schema.ItemFilter(insertObj[itemName], nil, &ute.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, false)
		if err != 0 {
			return err
		}

		if itemName == altLoginItem {
			altLogin = ute.data[schemaItem.DataIndex()].(string)
		}
	}

	// Encrypt password and store in entry
	ePass, ePassErr := helpers.EncryptString(password, t.encryptCost.Load().(int))
	if ePassErr != nil {
		return helpers.ErrorPasswordEncryption
	}
	ute.password.Store(ePass)

	// Make JSON []byte for entry
	var jBytes []byte
	if jErr := makeJsonBytes(name, ePass, ute.data, &jBytes); jErr != 0 {
		return jErr
	}

	// Lock table, check for duplicate entry
	maxEntries := t.maxEntries.Load().(uint64)
	t.eMux.Lock()
	if t.entries[name] != nil {
		t.eMux.Unlock()
		return helpers.ErrorEntryNameInUse
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
	lineOn, aErr := storage.Insert(t.persistName + "/" + strconv.Itoa(int(t.fileOn)) + storage.FileTypeStorage, jBytes)
	if aErr != 0 {
		t.uMux.Unlock()
		t.eMux.Unlock()
		return aErr
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
	ute.persistIndex = lineOn
	ute.persistFile = t.fileOn

	// Increase fileOn when the index has reached or surpassed partitionMax
	if ute.persistIndex >= t.partitionMax.Load().(uint16) {
		t.fileOn++
	}

	// Remove data from memory if dataOnDrive is true
	if t.dataOnDrive {
		ute.data = nil
	}

	// Apply altLogin
	if altLogin != "" {
		t.altLogins[altLogin] = &ute
	}

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
func (t *AuthTable) GetUserData(userName string, password string, items []string) (map[string]interface{}, int) {
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

func (t *AuthTable) dataFromDrive(file string, index uint16) ([]interface{}, int) {
	// Read bytes from file
	bytes, rErr := storage.Read(file, index)
	if rErr != 0 {
		return nil, rErr
	}
	jMap := make(map[string]interface{})
	jErr := json.Unmarshal(bytes, &jMap)
	if jErr != nil {
		fmt.Println("got data from '"+file+"': '"+string(bytes)+"'")
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
func (t *AuthTable) UpdateUserData(userName string, password string, updateObj map[string]interface{}) int {
	if updateObj == nil || len(updateObj) == 0 {
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
	var altLoginBefore string
	var altLoginAfter string
	altLoginItem := t.altLoginItem.Load().(string)
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

		//
		if updateName == altLoginItem {
			altLoginBefore = itemBefore.(string)
			altLoginAfter = data[schemaItem.DataIndex()].(string)
		}
	}

	// Make JSON []byte for entry
	var jBytes []byte
	if jErr := makeJsonBytes(userName, e.password.Load().([]byte), data, &jBytes); jErr != 0 {
		return jErr
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
	uErr := storage.Update(t.persistName + "/" + strconv.Itoa(int(e.persistFile)) + storage.FileTypeStorage, e.persistIndex, jBytes)
	if uErr != 0 {
		t.uMux.Unlock()
		e.mux.Unlock()
		return uErr
	}

	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if t.uniqueVals[itemName] == nil {
			t.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		t.uniqueVals[itemName][itemVal] = true
	}
	t.uMux.Unlock()

	// Apply alt login and remove old if changed
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

func (t *AuthTable) ChangePassword(userName string, password string, newPassword string) int {
	if len(newPassword) < int(t.minPassword.Load().(uint8)) {
		return helpers.ErrorPasswordLength
	}

	ue, err := t.Get(userName, password)
	if err != 0 {
		return err
	}

	// Encrypt new password
	ePass, eErr := helpers.EncryptString(newPassword, t.encryptCost.Load().(int))
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
	} else {
		ue.mux.Lock()
		data = append([]interface{}{}, ue.data...)
		ue.mux.Unlock()
	}

	// Make JSON []byte for entry
	var jBytes []byte
	if jErr := makeJsonBytes(userName, ePass, data, &jBytes); jErr != 0 {
		return jErr
	}

	// Update entry on disk with jBytes
	uErr := storage.Update(t.persistName + "/" + strconv.Itoa(int(ue.persistFile)) + storage.FileTypeStorage, ue.persistIndex, jBytes)
	if uErr != 0 {
		return uErr
	}

	ue.password.Store(ePass)

	//
	return 0
}

func (t *AuthTable) ResetPassword(userName string) int {
	// Name and password are required
	if len(userName) == 0 {
		return helpers.ErrorNameRequired
	} else if t.emailItem.Load().(string) == "" {
		// Database shouldn't change password without sending an email to the user
		return helpers.ErrorNoEmailItem
	}

	// Generate new password
	newPass, pErr := helpers.GenerateRandomBytes(int(t.passResetLen.Load().(uint8)))
	if pErr != nil {
		return helpers.ErrorPasswordEncryption
	}

	// Send newPass to emailItem, do not proceed unless the email was a success !!!

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

	if ue == nil && t.altLoginItem.Load().(string) != "" {
		ue = t.altLogins[userName]
	}

	//
	if ue == nil {
		return helpers.ErrorInvalidNameOrPassword
	}
	// Change password
	ePass, eErr := helpers.EncryptString(string(newPass), t.encryptCost.Load().(int))
	if eErr != nil {
		return helpers.ErrorPasswordEncryption
	}

	// Delete auto-login hashes !!!

	// Make JSON []byte for entry
	var jBytes []byte
	if jErr := makeJsonBytes(userName, ePass, data, &jBytes); jErr != 0 {
		return jErr
	}

	// Update entry on disk with jBytes
	uErr := storage.Update(t.persistName + "/" + strconv.Itoa(int(ue.persistFile)) + storage.FileTypeStorage, ue.persistIndex, jBytes)
	if uErr != 0 {
		return uErr
	}

	ue.password.Store(ePass)

	//
	return 0
}

func (t *AuthTable) DeleteUser(userName string, password string) int {
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
	altLoginItem := t.altLoginItem.Load().(string)
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
		if itemName == altLoginItem {
			t.eMux.Lock()
			delete(t.altLogins, i.(string))
			t.eMux.Unlock()
		}
		delete(t.uniqueVals[itemName], i)
	}
	ue.mux.Unlock()
	t.uMux.Unlock()

	// Update entry on disk with []byte{}
	uErr := storage.Update(t.persistName + "/" + strconv.Itoa(int(ue.persistFile)) + storage.FileTypeStorage, ue.persistIndex, []byte{})
	if uErr != 0 {
		return uErr
	}

	t.eMux.Lock()
	// Delete entry
	delete(t.entries, userName)
	t.eMux.Unlock()

	//
	return 0
}