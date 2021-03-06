package helpers

type Error struct {
	ID   int    // Error ID number
	From string // From whence it came, and general error messages
}

const (
	// Unexpected error
	ErrorUnexpected = 1000 + iota // Unexpected internal error...

	// Generic Table Query Errors
	ErrorTableExists
	ErrorTableDoesntExist
	ErrorTableNameRequired
	ErrorInvalidKeyCharacters
	ErrorTableFull
	ErrorQueryInvalidFormat
	ErrorNoEntryFound
)

const (
	// Schema creation errors
	ErrorSchemaRequired = 2001 + iota
	ErrorSchemaItemsRequired
	ErrorSchemaInvalidItemName
	ErrorSchemaInvalidItemPosition
	ErrorSchemaInvalidItemType
	ErrorSchemaInvalidItemParameters
	ErrorSchemaInvalidFormat
	ErrorSchemaInvalidTimeFormat
	ErrorSchemaInvalid
	ErrorObjectItemNotRequired // 2010

	// Query Schema errors
	ErrorInvalidItem
	ErrorInvalidItemValue
	ErrorInvalidMethod
	ErrorInvalidMethodParameters
	ErrorTooManyMethodParameters
	ErrorNotEnoughMethodParameters
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorStringIsEncrypted // 2020
	ErrorEncryptingString
	ErrorArrayItemsRequired
	ErrorArrayEmpty
	ErrorArrayItemNotSortable
	ErrorIndexOutOfBounds
	ErrorMapItemsRequired
	ErrorInvalidTimeFormat
	ErrorUniqueValueDuplicate
	ErrorRestoreItemSchema
)

const (
	// Keystore Errors
	ErrorKeyRequired = 3001 + iota
	ErrorKeyInUse
)

const (
	// Auth Table Query Errors
	ErrorNameRequired = 4001 + iota
	ErrorNameInUse
	ErrorInvalidNameCharacters
	ErrorPasswordLength
	ErrorPasswordEncryption
	ErrorNoEmailItem
	ErrorIncorrectAuthType
	ErrorInvalidEmail
)

const (
	// Leaderboard errors
	ErrorLeaderboardExists = 5001 + iota
	ErrorLeaderboardDoesntExist
)

const (
	// Storage errors
	ErrorStorageNotInitialized = 9001 + iota
	ErrorTableFolderCreate
	ErrorCreatingFolder
	ErrorFileOpen
	ErrorFileAppend
	ErrorFileUpdate
	ErrorFileRead
	ErrorFileWrite
	ErrorFileDelete // 6010
	ErrorJsonEncoding
	ErrorJsonDecoding
	ErrorJsonDataFormat
	ErrorJsonIndexingFormat
	ErrorInternalFormatting
)

// NewError creates a new Error message with given ID and From message
func NewError(id int, from string) Error {
	return Error{ID: id, From: from}
}