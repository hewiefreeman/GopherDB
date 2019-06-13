package userTable

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"github.com/hewiefreeman/GopherGameDB/schema"
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
		name:         name,
		persistFile:  0,
		persistIndex: 0,
		data:         make([]interface{}, len(*(t.schema)), len(*(t.schema))),
	}

	// Fill entry data with insertObj - Loop through schema to also check for required items
	for itemName, schemaItem := range *(t.schema) {
		// Make filter
		filter := schema.NewFilter(insertObj[itemName], nil, &ute.data[schemaItem.DataIndex()], nil, schemaItem, &t.uMux, &t.uniqueVals)

		// Add updateItem to entry data slice
		err := schema.QueryItemFilter(&filter)
		if err != 0 {
			return err
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
	t.entries[name] = &ute
	t.eMux.Unlock()

	return 0
}

// Example JSON for get query:
//
//     {"GetUserData": {"table": "tableName", "query": ["userName", "password"]}}
//

// GetUserData
func (t *UserTable) GetUserData(userName string, password string) (map[string]interface{}, int) {
	t.sMux.Lock()
	minPass := t.minPassword
	t.sMux.Unlock()

	// Name and password are required
	if len(userName) == 0 {
		return nil, helpers.ErrorNameRequired
	} else if len(password) < int(minPass) {
		return nil, helpers.ErrorPasswordLength
	}

	// Get entry
	t.eMux.Lock()
	e := t.entries[userName]
	t.eMux.Unlock()

	if e == nil {
		return nil, helpers.ErrorInvalidNameOrPassword
	}
	if !e.CheckPassword(password) {
		return nil, helpers.ErrorInvalidNameOrPassword
	}

	// Make entry map
	m := make(map[string]interface{})
	e.mux.Lock()
	for k, v := range *(t.schema) {
		m[k] = schema.Format(e.data[v.DataIndex()], v)
	}
	e.mux.Unlock()

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
	t.sMux.Lock()
	minPass := t.minPassword
	t.sMux.Unlock()

	// Name, password, and updateObj are required
	if len(userName) == 0 {
		return helpers.ErrorNameRequired
	} else if len(password) < int(minPass) {
		return helpers.ErrorPasswordLength
	} else if len(updateObj) == 0 {
		return helpers.ErrorQueryInvalidFormat
	}

	// Get entry
	t.eMux.Lock()
	e := t.entries[userName]
	t.eMux.Unlock()

	// Check for valid entry and password
	if e == nil {
		return helpers.ErrorInvalidNameOrPassword
	}
	if !e.CheckPassword(password) {
		return helpers.ErrorInvalidNameOrPassword
	}

	// Lock entry
	e.mux.Lock()
	data := append([]interface{}{}, e.data...)

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

		// Make filter
		filter := schema.NewFilter(updateItem, itemMethods, &data[schemaItem.DataIndex()], data[schemaItem.DataIndex()], schemaItem, &t.uMux, &t.uniqueVals)

		// Add updateItem to entry data slice
		var err int
		err = schema.QueryItemFilter(&filter)
		if err != 0 {
			e.mux.Unlock()
			return err
		}
	}

	// Unlock entry
	e.data = data
	e.mux.Unlock()

	return 0
}

func (t *UserTable) DeleteUser(userName string, password string) int {
	t.sMux.Lock()
	minPass := t.minPassword
	t.sMux.Unlock()

	// Name and password are required
	if len(userName) == 0 {
		return helpers.ErrorNameRequired
	} else if len(password) < int(minPass) {
		return helpers.ErrorPasswordLength
	}

	t.eMux.Lock()
	ue := t.entries[userName]

	//
	if ue == nil {
		t.eMux.Unlock()
		return helpers.ErrorInvalidNameOrPassword
	} else if !ue.CheckPassword(password) {
		t.eMux.Unlock()
		return helpers.ErrorInvalidNameOrPassword
	}

	// Delete entry
	delete(t.entries, userName)
	t.eMux.Unlock()

	//
	return 0
}