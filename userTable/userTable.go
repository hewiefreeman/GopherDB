package userTable

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"sync"
)

////////////////// TODOs
//////////////////
//////////////////     - Logging & persisting
//////////////////         - Logging
//////////////////         - Persisting data to storage
//////////////////         - Updating storage data
//////////////////         - Updating/Restoring with log/persist files
//////////////////
//////////////////     - Password reset for UserTable
//////////////////         - Setting server name/address, subject & body message for password reset emails
//////////////////         - Send emails for password resets
//////////////////
//////////////////     - Database server
//////////////////         - Connection authentication
//////////////////         - Connection privillages
//////////////////
//////////////////     - Rate limiting
//////////////////
//////////////////     - Query router
//////////////////         - Connection authentication
//////////////////         - Connection privillages
//////////////////         - Sharding & Replication
//////////////////         - Distributed queries
//////////////////
//////////////////     - Distributed unique value checks
//////////////////
//////////////////     - Key-value & List tables

var (
	tablesMux      sync.Mutex
	logPersistTime int16                 = 60
	tables         map[string]*UserTable = make(map[string]*UserTable)
)

type UserTable struct {
	// settings and schema
	persistName   string // table's logger/persist folder name
	schema        *schema.Schema // table's schema
	dataOnDrive   bool // when true, entry data and password are not stored in memory (new entries only temporarily)
	partitionMax  uint16

	sMux          sync.Mutex // locks all table settings below
	maxEntries    uint64 // maximum amount of entries in the UserTable
	minPassword   uint8 // minimum password length
	encryptCost   int // encryption cost of passwords
	passResetLen  uint8 // the length of passwords created by the database
	emailItem     string // item in schema that represents a user's email address
	altLoginItem  string // item in schema that a user can log in with as if it's their user name (usually the emailItem)

	// entries
	eMux      sync.Mutex // entries/altLogins map lock
	entries   map[string]*UserTableEntry // UserTable uses a Map for storage since it's only look-up is with user name and password
	altLogins map[string]*UserTableEntry

	// unique values
	uMux       sync.Mutex
	uniqueVals map[string]map[interface{}]bool

	// persist numbers
	pMux      sync.Mutex
	fileOn    uint16
	lineOn    uint16
}

type UserTableEntry struct {
	persistFile  uint16
	persistIndex uint16

	mux      sync.Mutex
	password []byte
	data     []interface{}
}

// File/folder prefixes
const (
	prefixUserTable = "UT-"
)

// Defaults
const (
	defaultPartitionMax uint16 = 2500
	defaultMinPassword uint8   = 6
	defaultPassResetLen uint8  = 12
	defaultEncryptCost int     = 8
	encryptCostMax int         = 31
	encryptCostMin int         = 4
	defaultConfig string       = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"routerOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"routers\":[],\"UserTables\":[],\"Leaderboards\":[]}"
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
func New(name string, s *schema.Schema, maxEntries uint64, minPassword uint8, partitionMax uint16, fileOn uint16, lineOn uint16, dataOnDrive bool) (*UserTable, int) {
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

	namePre := prefixUserTable + name

	// Make table folder   & update config file !!!
	retChan := make(chan interface{})
	qErr := storage.QueueFileAction(storage.FileActionMakeDir, []interface{}{namePre}, retChan)
	if qErr != 0 {
		close(retChan)
		return nil, qErr
	}
	qRes := <-retChan
	close(retChan)
	if qRes != nil {
		return nil, helpers.ErrorCreatingFolder
	}

	// make table
	t := UserTable{persistName: namePre,
		dataOnDrive:   dataOnDrive,
		partitionMax:  partitionMax,
		maxEntries:    maxEntries,
		minPassword:   minPassword,
		passResetLen:  defaultPassResetLen,
		encryptCost:   defaultEncryptCost,
		schema:        s,
		entries:       make(map[string]*UserTableEntry),
		altLogins:     make(map[string]*UserTableEntry),
		uniqueVals:    make(map[string]map[interface{}]bool),
		fileOn:        fileOn,
		lineOn:        lineOn,
	}

	//
	tablesMux.Lock()
	tables[name] = &t
	tr := tables[name]
	tablesMux.Unlock()

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

func (t *UserTable) Get(userName string, password string) (*UserTableEntry, int) {
	t.sMux.Lock()
	minPass := t.minPassword
	t.sMux.Unlock()

	// Name and password are required
	if len(userName) == 0 {
		return nil, helpers.ErrorNameRequired
	} else if len(password) < int(minPass) {
		return nil, helpers.ErrorPasswordLength
	}

	// Find entry
	t.eMux.Lock()
	ue := t.entries[userName]
	if ue == nil && t.altLoginItem != "" {
		ue = t.altLogins[userName]
	}
	t.eMux.Unlock()

	// Check if found
	if ue == nil {
		return nil, helpers.ErrorInvalidNameOrPassword
	}

	// Check Password
	if !ue.CheckPassword(password) {
		return nil, helpers.ErrorInvalidNameOrPassword
	}

	return ue, 0
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
	if t.passResetLen < min {
		t.passResetLen = min
	}
	t.minPassword = min
	t.sMux.Unlock()
}

func (t *UserTable) SetPasswordResetLength(len uint8) {
	t.sMux.Lock()
	if len < t.minPassword {
		len = t.minPassword
	}
	t.passResetLen = len
	t.sMux.Unlock()
}

func (t *UserTable) SetAltLoginItem(item string) int {
	si := (*(t.schema))[item]
	if si == nil {
		return helpers.ErrorInvalidItem
	} else if si.TypeName() != schema.ItemTypeString || !si.Unique() {
		return helpers.ErrorInvalidItem
	}

	t.sMux.Lock()
	t.altLoginItem = item
	t.sMux.Unlock()
	return 0
}

func (t *UserTable) SetEmailItem(item string) int {
	si := (*(t.schema))[item]
	if si == nil {
		return helpers.ErrorInvalidItem
	} else if si.TypeName() != schema.ItemTypeString || !si.Unique() {
		return helpers.ErrorInvalidItem
	}

	t.sMux.Lock()
	t.emailItem = item
	t.sMux.Unlock()
	return 0
}