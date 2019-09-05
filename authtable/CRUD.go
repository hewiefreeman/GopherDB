package authtable

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"strconv"
	"strings"
	"encoding/json"
	"regexp"
)

type jsonEntry struct {
	N string
	P []byte
	D []interface{}
}

func makeJsonBytes(name string, password []byte, data []interface{}, jBytes *[]byte) int {
	var jErr error
	*jBytes, jErr = json.Marshal(jsonEntry{
		N: name,
		P: password,
		D: data,
	})
	if jErr != nil {
		return helpers.ErrorJsonEncoding
	}
	return 0
}

var (
	emailExp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Example JSON for new user query:
//
//     {"NewUser": {"table": "tableName", "query": ["userName", "password", { *items that match schema* }]}}
//

// NewUser creates a new authTableEntry in the AuthTable
func (t *AuthTable) NewUser(name string, password string, insertObj map[string]interface{}) int {
	// Name and password are required
	if len(name) == 0 {
		return helpers.ErrorNameRequired
	} else if strings.ContainsAny(name, " \t\n\r"){
		return helpers.ErrorInvalidKeyCharacters
	} else if len(password) < int(t.minPassword.Load().(uint8)) {
		return helpers.ErrorPasswordLength
	}

	// Create entry
	ute := authTableEntry{
		data: make([]interface{}, len(t.schema), len(t.schema)),
	}

	// Alternative login name
	var altLogin string
	altLoginItem := t.altLoginItem.Load().(string)
	emailItem := t.emailItem.Load().(string)
	uniqueVals := make(map[string]interface{})

	// Fill entry data with insertObj - Loop through schema to also check for required items
	for itemName, schemaItem := range t.schema {
		// Item filter
		err := schema.ItemFilter(insertObj[itemName], nil, &ute.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, false, false)
		if err != 0 {
			return err
		}

		if itemName == altLoginItem {
			altLogin = ute.data[schemaItem.DataIndex()].(string)
		} else if itemName == emailItem && !emailExp.MatchString(ute.data[schemaItem.DataIndex()].(string)) {
			return helpers.ErrorInvalidEmail
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
	if !t.memOnly {
		if jErr := makeJsonBytes(name, ePass, ute.data, &jBytes); jErr != 0 {
			return jErr
		}
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
	var lineOn uint16
	if !t.memOnly {
		var aErr int
		lineOn, aErr = storage.Insert(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(t.fileOn)) + helpers.FileTypeStorage, jBytes)
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
func (t *AuthTable) GetUserData(userName string, password string, sel []string) (map[string]interface{}, int) {
	e, err := t.Get(userName, password)
	if err != 0 {
		return nil, err
	}

	var data []interface{}

	// Get entry data
	if t.dataOnDrive {
		var dErr int
		data, dErr = t.dataFromDrive(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage, e.persistIndex)
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
	if sel != nil {
		for _, itemName := range sel {
			siName, itemMethods := schema.GetQueryItemMethods(itemName)
			//
			si := t.schema[siName]
			if !si.QuickValidate() {
				return nil, helpers.ErrorInvalidItem
			}
			// Item filter
			var i interface{}
			err := schema.ItemFilter(data[si.DataIndex()], itemMethods, &i, nil, si, nil, true, false)
			if err != 0 {
				return nil, err
			}
			m[itemName] = i
		}
	} else {
		for itemName, si := range t.schema {
			// Item filter
			var i interface{}
			err := schema.ItemFilter(data[si.DataIndex()], nil, &i, nil, si, nil, true, false)
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
		data, dErr = t.dataFromDrive(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage, e.persistIndex)
		if dErr != 0 {
			return dErr
		}
		e.mux.Lock()
	} else {
		e.mux.Lock()
		data = append([]interface{}{}, e.data...)
	}

	altLoginItem := t.altLoginItem.Load().(string)
	emailItem := t.emailItem.Load().(string)
	uniqueVals := make(map[string]interface{})
	uniqueValsBefore := make(map[string]interface{})

	// Iterate through updateObj
	for updateName, updateItem := range updateObj {
		var itemMethods []string
		updateName, itemMethods = schema.GetQueryItemMethods(updateName)

		// Check if valid schema item
		schemaItem := t.schema[updateName]
		if !schemaItem.QuickValidate() {
			e.mux.Unlock()
			return helpers.ErrorSchemaInvalid
		}

		itemBefore := data[schemaItem.DataIndex()]

		// Item filter
		err := schema.ItemFilter(updateItem, itemMethods, &data[schemaItem.DataIndex()], itemBefore, schemaItem, &uniqueVals, false, false)
		if err != 0 {
			e.mux.Unlock()
			return err
		}

		//
		if uniqueVals[updateName] != nil && data[schemaItem.DataIndex()] != itemBefore {
			uniqueValsBefore[updateName] = itemBefore
		}
		if updateName == emailItem && !emailExp.MatchString(data[schemaItem.DataIndex()].(string)) {
			return helpers.ErrorInvalidEmail
		}
	}

	// Make JSON []byte for entry
	var jBytes []byte
	if !t.memOnly {
		if jErr := makeJsonBytes(userName, e.password.Load().([]byte), data, &jBytes); jErr != 0 {
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
		uErr := storage.Update(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(e.persistFile)) + helpers.FileTypeStorage, e.persistIndex, jBytes)
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

		// Remove old unique values   !!! NEEDS TESTING
		if uniqueValsBefore[itemName] != nil {
			delete(t.uniqueVals[itemName], uniqueValsBefore[itemName])
			// Replace altLoginItem as well if that item changed
			if itemName == altLoginItem {
				t.eMux.Lock()
				delete(t.altLogins, uniqueValsBefore[itemName].(string))
				t.altLogins[itemVal.(string)] = e
				t.eMux.Unlock()
			}
		}
	}
	t.uMux.Unlock()
	//
	if !t.dataOnDrive {
		e.data = data
	}
	e.mux.Unlock()

	return 0
}

// ChangePassword
func (t *AuthTable) ChangeUserPassword(userName string, password string, newPassword string) int {
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
		data, dErr = t.dataFromDrive(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex)
		if dErr != 0 {
			return dErr
		}
	} else {
		ue.mux.Lock()
		data = append([]interface{}{}, ue.data...)
		ue.mux.Unlock()
	}

	if !t.memOnly {
		// Make JSON []byte for entry
		var jBytes []byte
		if jErr := makeJsonBytes(userName, ePass, data, &jBytes); jErr != 0 {
			return jErr
		}

		// Update entry on disk with jBytes
		uErr := storage.Update(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex, jBytes)
		if uErr != 0 {
			return uErr
		}
	}

	ue.password.Store(ePass)

	//
	return 0
}

// ResetPassword
func (t *AuthTable) ResetUserPassword(userName string) int {
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
		data, dErr = t.dataFromDrive(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex)
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
		return helpers.ErrorNoEntryFound
	}
	// Change password
	ePass, eErr := helpers.EncryptString(string(newPass), t.encryptCost.Load().(int))
	if eErr != nil {
		return helpers.ErrorPasswordEncryption
	}

	// Delete auto-login hashes !!!

	if !t.memOnly {
		// Make JSON []byte for entry
		var jBytes []byte
		if jErr := makeJsonBytes(userName, ePass, data, &jBytes); jErr != 0 {
			return jErr
		}

		// Update entry on disk with jBytes
		uErr := storage.Update(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex, jBytes)
		if uErr != 0 {
			return uErr
		}
	}

	ue.password.Store(ePass)

	//
	return 0
}

// DeleteUser
func (t *AuthTable) DeleteUser(userName string, password string) int {
	ue, err := t.Get(userName, password)
	if err != 0 {
		return err
	}

	var data []interface{}

	// Get entry data
	if t.dataOnDrive {
		var dErr int
		data, dErr = t.dataFromDrive(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex)
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
		si := t.schema[siName]
		if !si.QuickValidate() {
			t.uMux.Unlock()
			ue.mux.Unlock()
			return helpers.ErrorUnexpected
		}
		// Make filter
		var i interface{}
		err := schema.ItemFilter(data[si.DataIndex()], itemMethods, &i, nil, si, nil, true, false)
		if err != 0 {
			t.uMux.Unlock()
			ue.mux.Unlock()
			return helpers.ErrorUnexpected
		}
		if itemName == altLoginItem {
			t.eMux.Lock()
			delete(t.altLogins, i.(string))
			t.eMux.Unlock()
		}
		delete(t.uniqueVals[itemName], i)
	}
	t.uMux.Unlock()
	ue.mux.Unlock()

	// Delete entry
	t.eMux.Lock()
	delete(t.entries, userName)
	t.eMux.Unlock()

	// Update entry on disk with []byte{}
	if !t.memOnly {
		uErr := storage.Update(dataFolderPrefix + t.name + "/" + strconv.Itoa(int(ue.persistFile)) + helpers.FileTypeStorage, ue.persistIndex, []byte{})
		if uErr != 0 {
			return uErr
		}
	}

	//
	return 0
}

// RestoreUser is NOT concurrently safe! Use authtable.Restore() instead.
func (t *AuthTable) RestoreUser(name string, pass []byte, data []interface{}, fileOn uint16, lineOn uint16, altLoginItem string) int {
	// Check for duplicate entry
	if t.entries[name] != nil {
		return helpers.ErrorKeyInUse
	}

	// Create entry
	e := authTableEntry{
		data: make([]interface{}, len(t.schema), len(t.schema)),
	}

	uniqueVals := make(map[string]interface{})
	altLogin := ""

	// Fill entry data with data
	for itemName, schemaItem := range t.schema {
		// Check for out of range item
		if int(schemaItem.DataIndex()) > len(data)-1 {
			return helpers.ErrorRestoreItemSchema
		}

		// Item filter
		err := schema.ItemFilter(data[schemaItem.DataIndex()], nil, &e.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, false, true)
		if err != 0 {
			return err
		}

		if itemName == altLoginItem {
			altLogin = e.data[schemaItem.DataIndex()].(string)
		}
	}

	e.password.Store(pass)

	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if t.uniqueVals[itemName] == nil {
			t.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		t.uniqueVals[itemName][itemVal] = true
	}

	//
	e.persistIndex = lineOn
	e.persistFile = fileOn

	// Remove data from memory if dataOnDrive is true
	if t.dataOnDrive {
		e.data = nil
	}

	// Apply altLogin
	if altLogin != "" {
		t.altLogins[altLogin] = &e
	}

	// Insert item
	t.entries[name] = &e
	return 0
}