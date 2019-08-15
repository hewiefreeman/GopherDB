package helpers

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
	ErrorTableFull
	ErrorQueryInvalidFormat
	ErrorInvalidItem
	ErrorInvalidMethod
	ErrorInvalidMethodParameters

	// Keystore Errors
	ErrorKeyRequired
	ErrorKeyInUse

	// Auth Table Query Errors
	ErrorNameRequired
	ErrorPasswordLength // 1020
	ErrorPasswordEncryption
	ErrorNoEmailItem
	ErrorNoEntryFound

	// Schema errors
	ErrorInvalidItemValue
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorArrayItemsRequired
	ErrorArrayEmpty
	ErrorMapItemsRequired // 1030
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
	ErrorCreatingFolder
	ErrorFileOpen // 1040
	ErrorFileAppend
	ErrorFileUpdate
	ErrorFileRead
	ErrorJsonEncoding
	ErrorJsonDecoding
	ErrorJsonDataFormat

	// Unexpected error
	ErrorUnexpected
)
