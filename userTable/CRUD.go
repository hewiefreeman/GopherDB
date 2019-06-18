package userTable

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
)

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
	} else if maxEntries > 0 && t.Size() == int(maxEntries) {
		return helpers.ErrorTableFull
	}

	// Create entry
	ute := UserTableEntry{
		persistFile:  0,
		persistIndex: 0,
		data:         make([]interface{}, len(*(t.schema)), len(*(t.schema))),
	}

	// Alternative login name
	altLogin := ""
	uniqueVals := make(map[string]interface{})

	// Fill entry data with insertObj - Loop through schema to also check for required items
	for itemName, schemaItem := range *(t.schema) {
		// Make filter
		filter := schema.NewFilter(insertObj[itemName], nil, &ute.data[schemaItem.DataIndex()], nil, schemaItem, &uniqueVals, false)

		// Add updateItem to entry data slice
		err := schema.QueryItemFilter(&filter)
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

	// Insert into table
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
	// Get entry data
	e.mux.Lock()
	data := e.data
	e.mux.Unlock()

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
			// Make filter
			var i interface{}
			filter := schema.NewFilter(data[si.DataIndex()], itemMethods, &i, nil, si, nil, true)
			err = schema.QueryItemFilter(&filter)
			if err != 0 {
				return nil, err
			}
			m[itemName] = i
		}
	} else {
		for itemName, si := range *(t.schema) {
			var i interface{}
			filter := schema.NewFilter(data[si.DataIndex()], nil, &i, nil, si, nil, true)
			err = schema.QueryItemFilter(&filter)
			if err != 0 {
				return nil, err
			}
			m[itemName] = i
		}
	}
	return m, 0
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

	// Lock entry
	e.mux.Lock()
	data := append([]interface{}{}, e.data...)
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

		if updateName == t.altLoginItem {
			altLoginBefore = e.data[schemaItem.DataIndex()].(string)
		}

		// Make filter
		filter := schema.NewFilter(updateItem, itemMethods, &data[schemaItem.DataIndex()], e.data[schemaItem.DataIndex()], schemaItem, &uniqueVals, false)

		// Add updateItem to entry data slice
		err := schema.QueryItemFilter(&filter)
		if err != 0 {
			e.mux.Unlock()
			return err
		}

		if updateName == t.altLoginItem {
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
	// Apply unique values
	for itemName, itemVal := range uniqueVals {
		if t.uniqueVals[itemName] == nil {
			t.uniqueVals[itemName] = make(map[interface{}]bool)
		}
		t.uniqueVals[itemName][itemVal] = true
	}
	if altLoginBefore != "" {
		t.eMux.Lock()
		delete(t.altLogins, altLoginBefore)
		t.altLogins[altLoginAfter] = e
		t.eMux.Unlock()
	}
	t.uMux.Unlock()
	e.data = data
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
	ue.mux.Lock()
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

	t.eMux.Lock()
	// Delete entry
	delete(t.entries, userName)
	t.eMux.Unlock()

	t.uMux.Lock()
	ue.mux.Lock()
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
		filter := schema.NewFilter(ue.data[si.DataIndex()], itemMethods, &i, nil, si, nil, true)
		err = schema.QueryItemFilter(&filter)
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

	//
	return 0
}