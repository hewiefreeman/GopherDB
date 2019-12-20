package helpers

type Error struct {
	ID  int
	Msg string
}

const (
	// Unexpected error
	ErrorUnexpected = iota + 1000

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
	ErrorSchemaRequired = iota + 2001
	ErrorSchemaItemsRequired
	ErrorSchemaInvalidItemName
	ErrorSchemaInvalidItemType
	ErrorSchemaInvalidItemParameters
	ErrorSchemaInvalidFormat
	ErrorSchemaInvalidTimeFormat
	ErrorSchemaInvalid
	ErrorObjectItemNotRequired // Items inside Objects must be required!

	// SQuery Schema errors
	ErrorInvalidItem // 2010
	ErrorInvalidItemValue
	ErrorInvalidMethod
	ErrorInvalidMethodParameters
	ErrorTooManyMethodParameters
	ErrorNotEnoughMethodParameters
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorArrayItemsRequired
	ErrorArrayEmpty // 2020
	ErrorMapItemsRequired
	ErrorInvalidTimeFormat
	ErrorUniqueValueDuplicate
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
	ErrorLoggerExists = iota + 9001
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