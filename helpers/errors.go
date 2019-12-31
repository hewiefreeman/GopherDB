package helpers

type Error struct {
	ID  int
	From string
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
	ErrorArrayItemsRequired // 2020
	ErrorArrayEmpty
	ErrorArrayNotSortable
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
	ErrorEntryNameInUse
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
	ErrorLoggerExists = 9001 + iota
	ErrorLoggerFileCreate
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
)