package userTable

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"github.com/hewiefreeman/GopherGameDB/schema"
)

// NewUser creates a new UserTableEntry in the UserTable
func (t *UserTable) NewUser(name string, password string, insertObj map[string]interface{}) int {
	// Name and password are required
	if len(name) == 0 {
		return helpers.ErrorInsertNameRequired
	} else if len(password) < int(t.minPassword) {
		return helpers.ErrorInsertPasswordLength
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
		if insertItem == nil {
			// Get default value
			defaultVal, defaultErr := schema.GetDefaultVal(schemaItem.ItemType())
			if defaultErr != 0 {
				return defaultErr
			}
			ute.data[schemaItem.DataIndex()] = defaultVal
		} else {
			var iTypeErr int
			insertItem, iTypeErr = schema.CheckQueryItemType(insertItem, schemaItem.ItemType())
			if iTypeErr != 0 {
				return iTypeErr
			}
			ute.data[schemaItem.DataIndex()] = insertItem
		}
	}

	// Encrypt password and store in entry
	ePass, ePassErr := helpers.EncryptString(password, t.encryptCost)
	if ePassErr != nil {
		return helpers.ErrorInsertPasswordEncryption
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

func (t *UserTable) GetUserData(userName string, password string) (map[string]interface{}, int) {
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

func (t *UserTable) UpdateUserData(userName string, password string, params map[string]interface{}) int {

	return 0
}

func (t *UserTable) DeleteUser(userName string, password string) int {
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