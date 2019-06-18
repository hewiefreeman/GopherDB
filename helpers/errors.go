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
	ErrorNoEmailItem

	// Generic Table Query Errors
	ErrorQueryInvalidFormat
	ErrorInvalidItem
	ErrorInvalidArithmeticOperator
	ErrorInvalidArithmeticParameter
	ErrorInvalidMethod // 1020
	ErrorInvalidMethodParameters
	ErrorTableFull

	// Schema errors
	ErrorInvalidItemValue
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorNumberTooLarge
	ErrorNumberTooSmall
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
	ErrorInvalidNameOrPassword

	// Unexpected error
	ErrorUnexpected
)
