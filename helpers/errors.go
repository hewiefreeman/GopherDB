package helpers

type Error struct {
	ID  int
	Msg string
}

const (
	// Table should or shouldn't exist
	ErrorTableExists      = iota + 1001
	ErrorTableDoesntExist

	// Missing or invalid schema requirements
	ErrorSchemaRequired
	ErrorSchemaItemsRequired
	ErrorSchemaInvalidItemName
	ErrorSchemaInvalidItemType
	ErrorSchemaInvalidItemParameters
	ErrorSchemaInvalidFormat
	ErrorSchemaInvalidTimeFormat
	ErrorSchemaInvalid

	// Generic Table Query Errors
	ErrorTableNameRequired
	ErrorInvalidKeyCharacters
	ErrorTableFull
	ErrorQueryInvalidFormat
	ErrorInvalidItem
	ErrorInvalidMethod
	ErrorInvalidMethodParameters

	// Keystore Errors
	ErrorKeyRequired
	ErrorKeyInUse

	// Auth Table Query Errors
	ErrorNameRequired // 1020
	ErrorPasswordLength
	ErrorPasswordEncryption
	ErrorNoEmailItem
	ErrorNoEntryFound

	// Schema errors
	ErrorInvalidItemValue
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorArrayItemsRequired
	ErrorArrayEmpty // 1030
	ErrorMapItemsRequired
	ErrorInvalidTimeFormat

	// Leaderboard errors
	ErrorLeaderboardExists
	ErrorLeaderboardDoesntExist

	// Unique value errors
	ErrorUniqueValueInUse
	ErrorEntryNameInUse

	// Storage errors
	ErrorLoggerExists
	ErrorLoggerFileCreate
	ErrorTableFolderCreate
	ErrorCreatingFolder // 1040
	ErrorFileOpen
	ErrorFileAppend
	ErrorFileUpdate
	ErrorFileRead
	ErrorFileWrite
	ErrorFileDelete
	ErrorJsonEncoding
	ErrorJsonDecoding
	ErrorJsonDataFormat

	// Restoring Errors
	ErrorRestoreItemSchema

	// Unexpected error
	ErrorUnexpected
)