package authtable

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"sync"
	"sync/atomic"
)

////////////////// TODOs
//////////////////
//////////////////     - Password reset for AuthTable:
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
//////////////////     - Keystore & Datelist tables

var (
	tablesMux      sync.Mutex
	tables         map[string]*AuthTable = make(map[string]*AuthTable)
)

type AuthTable struct {
	fileOn    uint16 // locked by eMux - placed for memory efficiency

	// Settings and schema - read only
	memOnly       bool // Store data in memory only (overrides dataOnDrive)
	dataOnDrive   bool // when true, entry data is not stored in memory, only indexing and password
	persistName   string // table's logger/persist folder name
	schema        *schema.Schema // table's schema

	// Atomic changable settings values - 99% read
	partitionMax  atomic.Value // *uint16* maximum entries per data file
	maxEntries    atomic.Value // *uint64* maximum amount of entries in the AuthTable
	minPassword   atomic.Value // *uint8* minimum password length
	encryptCost   atomic.Value // *int* encryption cost of passwords
	passResetLen  atomic.Value // *uint8* the length of passwords created by the database
	emailItem     atomic.Value // *string* item in schema that represents a user's email address
	altLoginItem  atomic.Value // *string* item in schema that a user can log in with as if it's their user name (usually the emailItem)

	// entries
	eMux      sync.Mutex // entries/altLogins map lock
	entries   map[string]*AuthTableEntry // AuthTable uses a Map for storage since it's only look-up is with user name and password
	altLogins map[string]*AuthTableEntry

	// unique values
	uMux       sync.Mutex
	uniqueVals map[string]map[interface{}]bool
}

type AuthTableEntry struct {
	persistFile  uint16
	persistIndex uint16

	password atomic.Value

	mux  sync.Mutex
	data []interface{}
}

// File/folder prefixes
const (
	dataFolderPrefix = "AT-"
)

// Defaults
const (
	defaultMinPassword uint8   = 6
	defaultPassResetLen uint8  = 12
	defaultConfig string       = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"routerOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"routers\":[],\"AuthTables\":[],\"Leaderboards\":[]}"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   AuthTable   ////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

//	Example JSON query to make a new AuthTable:
//
//		{"NewAuthTable": [
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
//			0, 0, 0, 0, false
//		]};
//

// New creates a new AuthTable with the provided name, schema, and other parameters.
func New(name string, s *schema.Schema, fileOn uint16, dataOnDrive bool, memOnly bool) (*AuthTable, int) {
	if len(name) == 0 {
		return nil, helpers.ErrorTableNameRequired
	} else if Get(name) != nil {
		return nil, helpers.ErrorTableExists
	} else if !s.ValidSchema() {
		return nil, helpers.ErrorTableExists
	}

	// memOnly overrides dataOnDrive
	if memOnly {
		dataOnDrive = false
	}

	namePre := dataFolderPrefix + name

	// Make table folder   & update config file !!!
	mkErr := storage.MakeDir(namePre)
	if mkErr != nil {
		return nil, helpers.ErrorCreatingFolder
	}

	// make table
	t := AuthTable{
		persistName:   namePre,
		memOnly:       memOnly,
		dataOnDrive:   dataOnDrive,
		schema:        s,
		entries:       make(map[string]*AuthTableEntry),
		altLogins:     make(map[string]*AuthTableEntry),
		uniqueVals:    make(map[string]map[interface{}]bool),
		fileOn:        fileOn,
	}

	// set defaults
	t.partitionMax.Store(helpers.DefaultPartitionMax)
	t.maxEntries.Store(helpers.DefaultMaxEntries)
	t.minPassword.Store(defaultMinPassword)
	t.encryptCost.Store(helpers.DefaultEncryptCost)
	t.passResetLen.Store(defaultPassResetLen)
	t.emailItem.Store("")
	t.altLoginItem.Store("")

	// push to tables map
	tablesMux.Lock()
	tables[name] = &t
	tablesMux.Unlock()

	return &t, 0
}

// Delete deletes a AuthTable with the given name.
func Delete(name string) int {
	if len(name) == 0 {
		return helpers.ErrorTableNameRequired
	} else if Get(name) == nil {
		return helpers.ErrorTableDoesntExist
	}

	tablesMux.Lock()
	delete(tables, name)
	tablesMux.Unlock()

	// !!!!!! delete data folder from system & delete log file & update config file

	return 0
}

// Get retrieves a AuthTable by name
func Get(name string) *AuthTable {
	tablesMux.Lock()
	t := tables[name]
	tablesMux.Unlock()

	return t
}

func (t *AuthTable) Get(userName string, password string) (*AuthTableEntry, int) {
	// Name and password are required
	if len(userName) == 0 {
		return nil, helpers.ErrorNameRequired
	} else if len(password) < int(t.minPassword.Load().(uint8)) {
		return nil, helpers.ErrorPasswordLength
	}

	// Find entry
	t.eMux.Lock()
	ue := t.entries[userName]
	if ue == nil && t.altLoginItem.Load().(string) != "" {
		ue = t.altLogins[userName]
	}
	t.eMux.Unlock()

	// Check if found
	if ue == nil {
		return nil, helpers.ErrorNoEntryFound
	}

	// Check Password
	if !ue.CheckPassword(password) {
		return nil, helpers.ErrorNoEntryFound
	}

	return ue, 0
}

// CheckPassword compares the AuthTableEntry's encrypted password with the given string password.
func (e *AuthTableEntry) CheckPassword(pass string) bool {
	p := e.password.Load().([]byte)
	return helpers.StringMatchesEncryption(pass, p)
}

func (t *AuthTable) Size() int {
	t.eMux.Lock()
	s := len(t.entries)
	t.eMux.Unlock()
	return s
}

func (t *AuthTable) SetEncryptionCost(cost int) {
	if cost > helpers.EncryptCostMax {
		cost = helpers.EncryptCostMax
	} else if cost < helpers.EncryptCostMin {
		cost = helpers.EncryptCostMin
	}
	t.encryptCost.Store(cost)
}

func (t *AuthTable) SetMaxEntries(max uint64) {
	if max < 0 {
		max = 0
	}
	t.maxEntries.Store(max)
}

func (t *AuthTable) SetMinPasswordLength(min uint8) {
	if min < 1 {
		min = 1
	}
	if t.passResetLen.Load().(uint8) < min {
		t.passResetLen.Store(min)
	}
	t.minPassword.Store(min)
}

func (t *AuthTable) SetPasswordResetLength(len uint8) {
	mLen := t.minPassword.Load().(uint8)
	if len < mLen {
		len = mLen
	}
	t.passResetLen.Store(len)
}

// SetAltLoginItem sets the AuthTable's alternative login item. Item must be a string and unique.
func (t *AuthTable) SetAltLoginItem(item string) int {
	si := (*(t.schema))[item]
	if si == nil {
		return helpers.ErrorInvalidItem
	} else if si.TypeName() != schema.ItemTypeString || !si.Unique() {
		return helpers.ErrorInvalidItem
	}

	t.altLoginItem.Store(item)
	return 0
}

// SetAltLoginItem sets the AuthTable's email item. Item must be a string and unique.
func (t *AuthTable) SetEmailItem(item string) int {
	si := (*(t.schema))[item]
	if si == nil {
		return helpers.ErrorInvalidItem
	} else if si.TypeName() != schema.ItemTypeString || !si.Unique() {
		return helpers.ErrorInvalidItem
	}

	t.emailItem.Store(item)
	return 0
}

func (t *AuthTable) SetPartitionMax(max uint16) {
	if max < helpers.PartitionMin {
		max = helpers.DefaultPartitionMax
	}
	t.partitionMax.Store(max)
}