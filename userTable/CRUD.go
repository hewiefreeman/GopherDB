package userTable

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

func (t *UserTable) Insert(name string, password string, insertObj map[string]interface{}) int {
	// Name and password are required
	if len(name) == 0 {
		return helpers.ErrorInsertNameRequired
	} else if len(password) < t.minPassword {
		return helpers.ErrorInsertPasswordLength
	}

	// Create entry
	ute := UserTableEntry{
		name:         name,
		persistFile:  0,
		persistIndex: 0,
		dataObj:      make([]interface{}, len(t.schema), len(t.schema)),
	}

	// Fill entry dataObj with insertObj
	for itemName, schemaItem := range t.schema {
		insertItem := insertObj[itemName]
		if insertItem == nil {
			// Get default value
			defaultVal, defaultErr := getDefaultVal(schemaItem.iType)
			if defaultErr != 0 {
				return defaultErr
			}
			ute.dataObj[schemaItem.dataIndex] = defaultVal
		} else if checkItemType(insertItem, schemaItem.iType) {
			ute.dataObj[schemaItem.dataIndex] = insertItem
		} else {
			return ErrorInsertInvalidItemType
		}
	}

	// Encrypt password and store in entry
	ePass, ePassErr := helpers.EncryptString(password)
	if ePassErr != nil {
		return helpers.ErrorInsertPasswordEncryption
	}
	ute.password = ePass

	// Insert into table
	t.eMux.Lock()
	if entries[name] != nil {
		t.eMux.Unlock()
		return helpers.ErrorEntryNameInUse
	}
	t.entries[name] = &ute
	t.eMux.Unlock()

	return 0
}