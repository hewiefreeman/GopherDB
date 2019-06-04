package userTable

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"github.com/hewiefreeman/GopherGameDB/schema"
	//"fmt"
)

// Example JSON for new user query:
//
//     {"NewUser": {"table": "tableName", "query": ["userName", "password", { *items that match schema* }]}}
//

// NewUser creates a new UserTableEntry in the UserTable
func (t *UserTable) NewUser(name string, password string, insertObj map[string]interface{}) int {
	// Name and password are required
	if len(name) == 0 {
		return helpers.ErrorNameRequired
	} else if len(password) < int(t.minPassword) {
		return helpers.ErrorPasswordLength
	}

	// Create entry
	ute := UserTableEntry{
		name:         name,
		persistFile:  0,
		persistIndex: 0,
		data:         make([]interface{}, len(*(t.schema)), len(*(t.schema))),
	}

	// Fill entry data with insertObj
	for itemName, schemaItem := range *(t.schema) {
		insertItem := insertObj[itemName]
		var filterErr int
		ute.data[schemaItem.DataIndex()], filterErr = schema.SchemaFilter(insertItem, schemaItem.ItemType())
		if filterErr != 0 {
			return filterErr
		}
	}

	// Encrypt password and store in entry
	ePass, ePassErr := helpers.EncryptString(password, t.encryptCost)
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
	// Name and password are required
	if len(userName) == 0 {
		return nil, helpers.ErrorNameRequired
	} else if len(password) < int(t.minPassword) {
		return nil, helpers.ErrorPasswordLength
	}

	m := make(map[string]interface{})

	// Get entry
	t.eMux.Lock()
	e := t.entries[userName]
	t.eMux.Unlock()

	if e == nil {
		return nil, helpers.ErrorInvalidUserName
	}
	if !e.CheckPassword(password) {
		return nil, helpers.ErrorInvalidPassword
	}

	// Make entry map
	e.mux.Lock()
	for k, v := range *(t.schema) {
		m[k] = e.data[v.DataIndex()]
	}
	e.mux.Unlock()

	return m, 0
}

// Example JSON for update query:
//
//  Changing a string:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"email": "differentemail@yahoo.com"}]}}
//
//  Increasing a float (or integer) type:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"mmr": ["+", 0.5]}]}} // can also be "-", "*", "/", "%"
//
//  Adding an item to an Array:
//     {"UpdateUserData": {"table": "tableName", "query": ["userName", "password", {"friends": ["+", 0.5]}]}} // can also be "-", "*", "/", "%"
//

// UpdateUserData
func (t *UserTable) UpdateUserData(userName string, password string, updateObj map[string]interface{}) int {
	// Name, password, and updateObj are required
	if len(userName) == 0 {
		return helpers.ErrorNameRequired
	} else if len(password) < int(t.minPassword) {
		return helpers.ErrorPasswordLength
	} else if len(updateObj) == 0 {
		return helpers.ErrorQueryInvalidFormat
	}

	// Get entry and it's data from entry map
	t.eMux.Lock()
	e := t.entries[userName]
	data := e.data
	t.eMux.Unlock()

	if e == nil {
		return helpers.ErrorInvalidUserName
	}
	if !e.CheckPassword(password) {
		return helpers.ErrorInvalidPassword
	}

	for updateName, updateItem := range updateObj {
		schemaItem := (*(*t).schema)[updateName]
		if schemaItem == nil {
			return helpers.ErrorSchemaInvalid
		}
		var err int
		// Apply number arithmetic and Array/Object methods to updateItem
		updateItem, err = schema.MethodFilter(updateItem, data[schemaItem.DataIndex()], schemaItem.ItemType())
		if err != 0 {
			return err
		}

		//  Add updateObj values to new entry data
		data[schemaItem.DataIndex()], err = schema.SchemaFilter(updateItem, schemaItem.ItemType())
		if err != 0 {
			return err
		}
	}

	// Update entry data with new data
	e.mux.Lock()
	e.data = data
	e.mux.Unlock()

	return 0
}

func (t *UserTable) DeleteUser(userName string, password string) int {
	// Name and password are required
	if len(userName) == 0 {
		return helpers.ErrorNameRequired
	} else if len(password) < int(t.minPassword) {
		return helpers.ErrorPasswordLength
	}

	t.eMux.Lock()
	ue := t.entries[userName]
	t.eMux.Unlock()

	//
	if ue == nil {
		return helpers.ErrorInvalidUserName
	}
	if !ue.CheckPassword(password) {
		return helpers.ErrorInvalidPassword
	}

	// Delete entry
	t.eMux.Lock()
	delete(t.entries, userName)
	t.eMux.Unlock()

	//
	return 0
}