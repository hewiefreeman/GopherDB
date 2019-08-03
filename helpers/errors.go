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

	// Auth Table Query Errors
	ErrorNameRequired
	ErrorPasswordLength
	ErrorPasswordEncryption
	ErrorNoEmailItem // 1020
	ErrorInvalidNameOrPassword

	// Schema errors
	ErrorInvalidItemValue
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorArrayItemsRequired
	ErrorArrayEmpty
	ErrorMapItemsRequired
	ErrorInvalidTimeFormat

	// Leaderboard errors
	ErrorLeaderboardExists // 1030
	ErrorLeaderboardDoesntExist

	// Unique value errors
	ErrorUniqueValueInUse
	ErrorEntryNameInUse

	// Storage errors
	ErrorLoggerExists
	ErrorLoggerFileCreate
	ErrorTableFolderCreate
	ErrorCreatingFolder
	ErrorFileOpen
	ErrorFileAppend
	ErrorFileUpdate // 1040
	ErrorFileRead
	ErrorJsonEncoding
	ErrorJsonDecoding
	ErrorJsonDataFormat

	// Unexpected error
	ErrorUnexpected
)
