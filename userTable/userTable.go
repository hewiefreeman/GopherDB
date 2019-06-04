package userTable

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"github.com/hewiefreeman/GopherGameDB/schema"
	"sync"
)

////////////////// TODOs
//////////////////
//////////////////     - Query methods
//////////////////         - Array type
//////////////////         - Object type
//////////////////         - String type (?)
//////////////////
//////////////////     - Unique value checks
//////////////////         - local
//////////////////         - distributed
//////////////////
//////////////////     - Logging & persisting
//////////////////         - Logging
//////////////////         - Updating/Restoring with log files
//////////////////         - Persisting data to storage
//////////////////         - Updating storage data

var (
	tablesMux      sync.Mutex
	logPersistTime int16                 = 30
	tables         map[string]*UserTable = make(map[string]*UserTable)
)

type UserTable struct {
	// settings and schema
	logFolder     string
	persistFolder string
	partitionMax  uint16
	maxEntries    uint64
	minPassword   uint8
	encryptCost   int
	schema        *schema.Schema

	// entries
	eMux    sync.Mutex
	entries map[string]*UserTableEntry

	// persistance
	pMux   sync.Mutex
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
	defaultPartitionMax = 1500
	defaultMinPassword = 6
	defaultEncryptCost = 8
	defaultConfig       = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"balancers\":[],\"UserTables\":[]}"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   UserTable   ////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

//
//
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
		fileOn:        fileOn,
		lineOn:        lineOn}

	//
	tablesMux.Lock()
	tables[name] = &t
	tr := tables[name]
	tablesMux.Unlock()

	// !!!!!! make new folder on system for persisting data & create log file & update config file

	return tr, 0
}

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

func Get(n string) *UserTable {
	tablesMux.Lock()
	t := tables[n]
	tablesMux.Unlock()
	return t
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//   UserTable Misc. Methods   //////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func (t *UserTable) Size() int {
	t.eMux.Lock()
	s := len(t.entries)
	t.eMux.Unlock()
	return s
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//   UserTableEntry Methods   ///////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func (t *UserTableEntry) Name() string {
	return t.name
}

func (t *UserTableEntry) CheckPassword(pass string) bool {
	t.mux.Lock()
	p := t.password
	t.mux.Unlock()
	return helpers.StringMatchesEncryption(pass, p)
}

func (t *UserTableEntry) PersistFile() uint16 {
	return t.persistFile
}

func (t *UserTableEntry) PersistIndex() uint16 {
	return t.persistIndex
}

func (t *UserTableEntry) Data() []interface{} {
	t.mux.Lock()
	d := t.data
	t.mux.Unlock()
	return d
}
