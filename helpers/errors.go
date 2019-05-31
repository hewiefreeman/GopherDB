package helpers

const (
	// Table should or shouldn't exist
	ErrorUserTableExists      = iota + 1001
	ErrorUserTableDoesntExist

	// Missing or invalid set-up requirements
	ErrorSchemaRequired
	ErrorUserTableNameRequired
	ErrorSchemaItemsRequired
	ErrorSchemaInvalidItemType
	ErrorSchemaInvalidItemParameters
	ErrorSchemaInvalidFormat
	ErrorSchemaInvalid

	// Insert Errors
	ErrorInsertNameRequired
	ErrorInsertPasswordLength
	ErrorInsertPasswordEncryption
	ErrorInsertInvalidFormat

	// Unhashable errors
	ErrorUnhashableQueryKey
	ErrorUnhashableQueryValue

	// Schema errors
	ErrorInvalidItemType
	ErrorMissingRequiredItem

	// Leaderboard errors
	ErrorLeaderboardExists
	ErrorLeaderboardDoesntExist

	// Unique value errors
	ErrorUniqueValueInUse
	ErrorEntryNameInUse

	// Unexpected error
	ErrorUnexpected
)
