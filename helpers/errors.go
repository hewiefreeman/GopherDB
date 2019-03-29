package helpers

const (
	// Table should or shouldn't exist
	ErrorTableExists      = 1001
	ErrorTableDoesntExist = 1002

	// Missing set-up requirements
	ErrorSchemaRequired      = 1003
	ErrorTableNameRequired   = 1004
	ErrorSchemaItemsRequired = 1005

	// Unhashable errors
	ErrorUnhashableGroupValue  = 1006
	ErrorUnhashableUniqueValue = 1007
	ErrorUnhashableQueryKey    = 1011

	// Table's index is being optimized
	ErrorTableIndexOptimizing = 1008

	// Requested index chunk is out of range
	ErrorIndexChunkOutOfRange = 1009

	// Schema errors
	ErrorSchemaItemDoesntExist = 1010
	ErrorNilGroupValue         = 1012
	ErrorSchemaMustMatch       = 1015

	// Leaderboard errors
	ErrorLeaderboardExists      = 1013
	ErrorLeaderboardDoesntExist = 1014

	// Unique value errors
	ErrorUniqueValueInUse       = 1016
)
