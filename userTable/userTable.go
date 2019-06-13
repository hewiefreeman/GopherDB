package userTable

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"github.com/hewiefreeman/GopherGameDB/schema"
	"sync"
)

////////////////// TODOs
//////////////////
//////////////////     - Unique value checks
//////////////////         - local
//////////////////         - distributed
//////////////////
//////////////////     - Password reset for UserTable
//////////////////
//////////////////     - Alternative login name item for UserTable
//////////////////
//////////////////     - Rate limit
//////////////////
//////////////////     - Logging & persisting
//////////////////         - Logging
//////////////////         - Updating/Restoring with log files
//////////////////         - Persisting data to storage
//////////////////         - Updating storage data
//////////////////
//////////////////     - Admin connections
//////////////////

var (
	tablesMux      sync.Mutex
	logPersistTime int16                 = 60
	tables         map[string]*UserTable = make(map[string]*UserTable)
)

type UserTable struct {
	// settings and schema
	logFolder     string
	persistFolder string
	schema        *schema.Schema
	partitionMax  uint16
	sMux          sync.Mutex // locks all table settings below
	maxEntries    uint64
	minPassword   uint8
	encryptCost   int

	// entries
	eMux    sync.Mutex // entries map lock
	entries map[string]*UserTableEntry // UserTable uses a Map for storage since it's only look-up is with user name and password

	// unique values
	uMux       sync.Mutex
	uniqueVals map[string]map[interface{}]bool

	// persistance
	pMux   sync.Mutex // fileOn/lineOn lock
	fileOn uint16
	lineOn uint16
}

type UserTableEntry struct {
	persistFile  uint16 // 0 - Not persisted
	persistIndex uint16 // 0 - Not persisted

	mux      sync.Mutex
	name     string
	password []byte
	data     []interface{}
}

// File/folder prefixes
const (
	prefixUserTableLogging           = "log-"
	prefixUserTableDataFolder        = "data-"
	prefixUserTableDataPartitionFile = "part-"
)

// File types
const (
	fileTypeConfig = ".gcf"
	fileTypeLog    = ".glf"
	fileTypeData   = ".gdf"
)

// Defaults
const (
	defaultPartitionMax = 2500
	defaultMinPassword  = 6
	defaultEncryptCost  = 8
	encryptCostMax      = 31
	encryptCostMin      = 4
	defaultConfig       = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"balancers\":[],\"UserTables\":[]}"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   UserTable   ////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

//	Example JSON query to make a new UserTable:
//
//		{"NewUserTable": [
//			"users",
//			{
//				"email": ["String", "", 0, true, true],
//				"friends": ["Array", ["Object", {
//									"name": ["String", "", 0, true, true],
//									"status": ["Uint8", 0, 0, 2, false]
//				}, false], 50, false],
//				"vCode": ["String", "", 0, true, false],
//				"verified": ["Bool", false]
//			},
//			0, 0, 0, 0
//		]};
//

// New creates a new UserTable with the provided name, schema, and other parameters.
func New(name string, s *schema.Schema, maxEntries uint64, minPassword uint8, partitionMax uint16, fileOn uint16, lineOn uint16) (*UserTable, int) {
	if len(name) == 0 {
		return nil, helpers.ErrorUserTableNameRequired
	} else if Get(name) != nil {
		return nil, helpers.ErrorUserTableExists
	} else if !s.ValidSchema() {
		return nil, helpers.ErrorUserTableExists
	}

	// defaults
	if partitionMax == 0 {
		partitionMax = defaultPartitionMax
	}
	if minPassword == 0 {
		minPassword = defaultMinPassword
	}

	// make table
	t := UserTable{logFolder: prefixUserTableLogging + name,
		persistFolder: prefixUserTableDataFolder + name,
		partitionMax:  partitionMax,
		maxEntries:    maxEntries,
		minPassword:   minPassword,
		encryptCost:   defaultEncryptCost,
		schema:        s,
		entries:       make(map[string]*UserTableEntry),
		uniqueVals:    make(map[string]map[interface{}]bool),
		fileOn:        fileOn,
		lineOn:        lineOn,
	}

	//
	tablesMux.Lock()
	tables[name] = &t
	tr := tables[name]
	tablesMux.Unlock()

	// !!!!!! make new folder on system for persisting data & create log file & update config file

	return tr, 0
}

// Delete deletes a UserTable with the given name.
func Delete(name string) int {
	if len(name) == 0 {
		return helpers.ErrorUserTableNameRequired
	} else if Get(name) == nil {
		return helpers.ErrorUserTableDoesntExist
	}

	tablesMux.Lock()
	delete(tables, name)
	tablesMux.Unlock()

	// !!!!!! delete data folder from system & delete log file & update config file

	return 0
}

// Get retrieves a UserTable by name
func Get(name string) *UserTable {
	tablesMux.Lock()
	t := tables[name]
	tablesMux.Unlock()

	return t
}

// CheckPassword compares the UserTableEntry's encrypted password with the given string password.
func (t *UserTableEntry) CheckPassword(pass string) bool {
	t.mux.Lock()
	p := t.password
	t.mux.Unlock()
	return helpers.StringMatchesEncryption(pass, p)
}

func (t *UserTable) Size() int {
	t.eMux.Lock()
	s := len(t.entries)
	t.eMux.Unlock()
	return s
}

func (t *UserTable) SetEncryptionCost(cost int) {
	if cost > encryptCostMax {
		cost = encryptCostMax
	} else if cost < encryptCostMin {
		cost = encryptCostMin
	}
	t.sMux.Lock()
	t.encryptCost = cost
	t.sMux.Unlock()
}

func (t *UserTable) SetMaxEntries(max uint64) {
	if max < 0 {
		max = 0
	}
	t.sMux.Lock()
	t.maxEntries = max
	t.sMux.Unlock()
}

func (t *UserTable) SetMinPasswordLength(min uint8) {
	if min < 1 {
		min = 1
	}
	t.sMux.Lock()
	t.minPassword = min
	t.sMux.Unlock()
}