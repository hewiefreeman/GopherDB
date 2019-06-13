package helpers

const (
	// Table should or shouldn't exist
	ErrorUserTableExists      = iota + 1001
	ErrorUserTableDoesntExist

	// Missing or invalid set-up requirements
	ErrorSchemaRequired
	ErrorSchemaItemsRequired
	ErrorSchemaInvalidItemName
	ErrorSchemaInvalidItemType
	ErrorSchemaInvalidItemParameters
	ErrorSchemaInvalidFormat
	ErrorSchemaInvalidTimeFormat
	ErrorSchemaInvalid

	// User Table Query Errors
	ErrorUserTableNameRequired
	ErrorNameRequired
	ErrorPasswordLength
	ErrorPasswordEncryption

	// Generic Table Query Errors
	ErrorQueryInvalidFormat
	ErrorInvalidArithmeticOperator
	ErrorInvalidArithmeticParameter
	ErrorInvalidMethod
	ErrorInvalidMethodParameters
	ErrorTableFull // 1020

	// Schema errors
	ErrorInvalidItemValue
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorNumberTooLarge
	ErrorNumberTooSmall
	ErrorArrayItemsRequired
	ErrorArrayEmpty
	ErrorMapItemsRequired
	ErrorInvalidTimeFormat // 1030

	// Leaderboard errors
	ErrorLeaderboardExists
	ErrorLeaderboardDoesntExist

	// Unique value errors
	ErrorUniqueValueInUse
	ErrorEntryNameInUse
	ErrorInvalidNameOrPassword

	// Unexpected error
	ErrorUnexpected
)
