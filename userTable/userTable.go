package userTable

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

////////////////// TODOs
//////////////////
//////////////////     - Logging & persisting data to storage
//////////////////         - Logging
//////////////////         - Updating/Restoring with log files
//////////////////         - Persisting data to storage
//////////////////         - Updating storage data
//////////////////
//////////////////     - Queries

var (
	tablesMux      sync.Mutex
	logPersistTime int16 = 30
	tables         map[string]*UserTable = make(map[string]*UserTable)
)

type UserTable struct {
	// settings and schema
	logFolder     string
	persistFolder string
	partitionMax  uint16
	maxEntries    uint64
	schema        UserTableSchema

	// entries
	eMux    sync.Mutex
	entries map[string]*UserTableEntry

	// persistance
	pMux   sync.Mutex
	fileOn uint16
	lineOn uint16
}

type UserTableEntry struct {
	name string

	persistFile  uint16
	persistIndex uint16

	mux   sync.Mutex
	password []byte
	data []interface{}
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
	defaultConfig       = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"balancers\":[],\"UserTables\":[]}"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   UserTable   ////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func New(name string, schema UserTableSchema, maxEntries uint64, partitionMax uint16, fileOn uint16, lineOn uint16) (*UserTable, int) {
	if len(name) == 0 {
		return nil, helpers.ErrorUserTableNameRequired
	} else if Get(name) != nil {
		return nil, helpers.ErrorUserTableExists
	} else if !validSchema(schema) {
		return nil, helpers.ErrorUserTableExists
	}

	// defaults
	if partitionMax <= 0 {
		partitionMax = defaultPartitionMax
	}
	if fileOn < 0 {
		fileOn = 0
	}
	if lineOn < 0 {
		lineOn = 0
	}

	// make table
	t := UserTable{	logFolder: prefixUserTableLogging+name,
					persistFolder: prefixUserTableDataFolder+name,
					partitionMax: partitionMax,
					maxEntries: maxEntries,
					schema: schema,
					entries: make(map[string]*UserTableEntry),
					fileOn: fileOn,
					lineOn: lineOn }

	//
	tablesMux.Lock()
	tr := &t
	tables[name] = tr
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
	delete (tables, name)
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
//   UserTable Methods   ////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func (t *UserTable) Size() int {
	(*t).eMux.Lock()
	s := len((*t).entries)
	(*t).eMux.Unlock()
	return s
}

func (t *UserTable) GetUserData(userName string) map[string]interface{} {
	m := make(map[string]interface{})

	// Get entry
	(*t).eMux.Lock()
	e := (*t).entries[userName]
	(*t).eMux.Unlock()

	// Make entry map
	e.mux.Lock()
	for k, v := range t.schema {
		m[k] = e.data[v.dataIndex]
	}
	e.mux.Unlock()

	return m
}

func (t *UserTable) GetUserItem(userName string, item string) interface{} {
	// Get entry
	(*t).eMux.Lock()
	e := (*t).entries[userName]
	(*t).eMux.Unlock()

	// Get item index
	schemaItem := t.schema[item]
	if !validSchemaItem(schemaItem) {
		return nil
	}

	// Get data item with schema index
	e.mux.Lock()
	i := e.data[schemaItem.dataIndex]
	e.mux.Unlock()

	return i
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
