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
	ErrorSchemaInvalid

	// User Table Query Errors
	ErrorUserTableNameRequired
	ErrorNameRequired
	ErrorPasswordLength
	ErrorPasswordEncryption

	// Generic Table Query Errors
	ErrorQueryInvalidFormat
	ErrorInvalidArithmeticParameters
	ErrorTableFull

	// Schema errors
	ErrorInvalidItemValue
	ErrorMissingRequiredItem
	ErrorStringTooLarge
	ErrorStringRequired
	ErrorNumberTooLarge
	ErrorNumberTooSmall

	// Leaderboard errors
	ErrorLeaderboardExists
	ErrorLeaderboardDoesntExist

	// Unique value errors
	ErrorUniqueValueInUse
	ErrorEntryNameInUse
	ErrorInvalidUserName
	ErrorInvalidPassword

	// Unexpected error
	ErrorUnexpected
)
