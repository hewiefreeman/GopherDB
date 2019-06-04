package helpers

const (
	// Table should or shouldn't exist
	ErrorUserTableExists      = iota + 1001
	ErrorUserTableDoesntExist

	// Missing or invalid set-up requirements
	ErrorSchemaRequired
	ErrorUserTableNameRequired
	ErrorSchemaItemsRequired
	ErrorSchemaInvalidItemName
	ErrorSchemaInvalidItemType
	ErrorSchemaInvalidItemParameters
	ErrorSchemaInvalidFormat
	ErrorSchemaInvalid

	// User Table Query Errors
	ErrorNameRequired
	ErrorPasswordLength
	ErrorPasswordEncryption
	ErrorQueryInvalidFormat
	ErrorInvalidArithmeticParameters

	// Unhashable errors
	ErrorUnhashableQueryKey
	ErrorUnhashableQueryValue

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
