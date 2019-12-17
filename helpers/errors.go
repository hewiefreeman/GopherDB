package helpers

type Error struct {
	ID  int
	Msg string
}

const (
	// Unexpected error
	ErrorUnexpected = iota + 1001

	// Generic Table Query Errors
	ErrorTableExists
	ErrorTableDoesntExist
	ErrorTableNameRequired
	ErrorInvalidKeyCharacters
	ErrorTableFull
	ErrorQueryInvalidFormat
	ErrorNoEntryFound
	ErrorInvalidItem
	ErrorInvalidMethod // 1010
	ErrorInvalidMethodParameters
)

const (
	// Schema creation errors
	ErrorSchemaRequired = iota + 2001
	ErrorSchemaItemsRequired
	ErrorSchemaInvalidItemName
	ErrorSchemaInvalidItemType
	ErrorSchemaInvalidItemParameters
	ErrorSchemaInvalidFormat
	ErrorSchemaInvalidTimeFormat
	ErrorSchemaInvalid
	ErrorObjectItemNotRequired // Items inside Objects must be required!

	// Schema query errors
	ErrorInvalidItemValue // 2010
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorArrayItemsRequired
	ErrorArrayEmpty
	ErrorMapItemsRequired
	ErrorInvalidTimeFormat
	ErrorUniqueValueDuplicate // Value already in database
	ErrorUniqueValueDuplicates // Values in query
	ErrorRestoreItemSchema
)

const (
	// Keystore Errors
	ErrorKeyRequired = iota + 3001
	ErrorKeyInUse
)

const (
	// Auth Table Query Errors
	ErrorNameRequired = iota + 4001
	ErrorEntryNameInUse
	ErrorPasswordLength
	ErrorPasswordEncryption
	ErrorNoEmailItem
	ErrorIncorrectAuthType
	ErrorInvalidEmail
)

const (
	// Leaderboard errors
	ErrorLeaderboardExists = iota + 5001
	ErrorLeaderboardDoesntExist
)

const (
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
	ErrorFileDelete // 6010
	ErrorJsonEncoding
	ErrorJsonDecoding
	ErrorJsonDataFormat
)