package helpers

const (
	// Shared table defaults
	DefaultPartitionMax uint16 = 250
	PartitionMin uint16        = 1
	DefaultMaxEntries uint64   = 0
	DefaultEncryptCost int     = 4
	EncryptCostMax int         = 31
	EncryptCostMin int         = 4
)

// File types
const (
	FileTypeConfig = ".gdbconf"
	FileTypeLog     = ".gdbl"
	FileTypeStorage = ".gdbs"
)