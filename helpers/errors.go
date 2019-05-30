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

	// Unhashable errors
	ErrorUnhashableQueryKey
	ErrorUnhashableQueryValue

	// Schema errors
	ErrorSchemaItemDoesntExist
	ErrorSchemaMissingRequiredItem

	// Leaderboard errors
	ErrorLeaderboardExists
	ErrorLeaderboardDoesntExist

	// Unique value errors
	ErrorUniqueValueInUse
	ErrorEntryNameInUse

	// Unexpected error
	ErrorUnexpected
)
