package helpers

type Error struct {
	ID  int
	Msg string
}

const (
	// Generic Table Query Errors
	ErrorTableExists      = iota + 1001
	ErrorTableDoesntExist
	ErrorTableNameRequired
	ErrorInvalidKeyCharacters
	ErrorTableFull
	ErrorQueryInvalidFormat
	ErrorNoEntryFound
	ErrorInvalidItem
	ErrorInvalidMethod
	ErrorInvalidMethodParameters

	// Missing or invalid schema requirements for creation
	ErrorSchemaRequired = iota + 2001
	ErrorSchemaItemsRequired
	ErrorSchemaInvalidItemName
	ErrorSchemaInvalidItemType
	ErrorSchemaInvalidItemParameters
	ErrorSchemaInvalidFormat
	ErrorSchemaInvalidTimeFormat
	ErrorSchemaInvalid

	// Schema errors
	ErrorInvalidItemValue
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorArrayItemsRequired
	ErrorArrayEmpty
	ErrorMapItemsRequired
	ErrorInvalidTimeFormat
	ErrorUniqueValueInUse
	ErrorRestoreItemSchema

	// Keystore Errors
	ErrorKeyRequired = iota + 3001
	ErrorKeyInUse

	// Auth Table Query Errors
	ErrorNameRequired = iota + 4001
	ErrorEntryNameInUse
	ErrorPasswordLength
	ErrorPasswordEncryption
	ErrorNoEmailItem
	ErrorIncorrectAuthType
	ErrorInvalidEmail

	// Leaderboard errors
	ErrorLeaderboardExists = iota + 5001
	ErrorLeaderboardDoesntExist

	// Storage errors
	ErrorLoggerExists = iota + 6001
	ErrorLoggerFileCreate
	ErrorTableFolderCreate
	ErrorCreatingFolder
	ErrorFileOpen
	ErrorFileAppend
	ErrorFileUpdate
	ErrorFileRead
	ErrorFileWrite
	ErrorFileDelete
	ErrorJsonEncoding
	ErrorJsonDecoding
	ErrorJsonDataFormat

	// Unexpected error
	ErrorUnexpected = 1000
)