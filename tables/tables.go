package tables

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

var (
	tablesMux      sync.Mutex
	logPersistTime int = 30
	tables         map[string]*table = make(map[string]*table)
)

type table struct {
	logFolder     string
	persistFolder string
	partitionMax  int
	maxEntries    int
	indexChunks   int
	schema        tableSchema

	eMux    sync.Mutex
	entries map[string]*tableEntry

	iMux  sync.Mutex
	index []*indexChunk

	uMux       sync.Mutex
	uniqueItems map[int]*uniqueTableEntry

	gMux        sync.Mutex
	groupedItems map[int]*groupedTableEntry

	pMux   sync.Mutex
	fileOn int
	lineOn int
}

type tableEntry struct {
	key string

	persistFile  int
	persistIndex int

	mux   sync.Mutex
	entry []interface{}
}

// File/folder prefixes
const (
	prefixTableLogging           = "log-"
	prefixTableDataFolder        = "data-"
	prefixTableDataPartitionFile = "part-"
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
	defaultIndexChunks  = 4
	defaultConfig       = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"balancers\":[],\"tables\":[]}"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   table   ////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func createTable(name string, schema tableSchema, maxEntries int, indexChunks int, partitionMax int, fileOn int, lineOn int) int {
	if len(name) == 0 {
		return helpers.ErrorTableNameRequired
	} else if tableExists(name) {
		return helpers.ErrorTableExists
	}

	//default partitionMax
	if partitionMax <= 0 {
		partitionMax = defaultPartitionMax
	}

	//default indexChunks
	if indexChunks <= 0 {
		indexChunks = defaultIndexChunks
	}

	//cant be under 0
	if fileOn < 0 {
		fileOn = 0
	}
	if lineOn < 0 {
		lineOn = 0
	}

	// make indexChunk list
	indexChunkList := make([]*indexChunk, indexChunks, indexChunks)
	for i := 0; i < indexChunks; i++ {
		ic := indexChunk{entries: []*tableEntry{}}
		indexChunkList[i] = &ic
	}

	// make table
	t := table{	logFolder: prefixTableLogging+name,
				persistFolder: prefixTableDataFolder+name,
				partitionMax: partitionMax,
				maxEntries: maxEntries,
				indexChunks: indexChunks,
				schema: schema,
				entries: make(map[string]*tableEntry),
				index: indexChunkList,
				uniqueItems: make(map[int]*uniqueTableEntry),
				groupedItems: make(map[int]*groupedTableEntry),
				fileOn: fileOn,
				lineOn: lineOn }

	// Apply schema
	for _, v := range schema.items {
		if v.unique {
			ute := uniqueTableEntry{vals: make(map[interface{}]*tableEntry)}
			t.uniqueItems[v.index] = &ute
		} else if v.grouped {
			gte := groupedTableEntry{groups: make(map[interface{}]*tableEntryGroup)}
			for i := 0; i < len(v.groupedVals); i++ {
				teg := tableEntryGroup{entries: []*tableEntry{}}
				gte.groups[v.groupedVals[i]] = &teg
			}
			t.groupedItems[v.index] = &gte
		}
	}

	//
	tablesMux.Lock()
	tables[name] = &t
	tablesMux.Unlock()

	// !!!!!! make new folder on system for persisting data & create log file & update config file

	return 0
}

func deleteTable(name string) int {
	if len(name) == 0 {
		return helpers.ErrorTableNameRequired
	} else if !tableExists(name) {
		return helpers.ErrorTableDoesntExist
	}

	tablesMux.Lock()
	delete (tables, name)
	tablesMux.Unlock()

	// !!!!!! delete data folder from system & delete log file & update config file

	return 0
}

func getTable(n string) *table {
	tablesMux.Lock()
	t := tables[n]
	tablesMux.Unlock()
	return t
}

func tableExists(n string) bool {
	tablesMux.Lock()
	var t bool = (tables[n] != nil)
	tablesMux.Unlock()
	return t
}

func (t *table) size() int {
	(*t).eMux.Lock()
	s := len((*t).entries)
	(*t).eMux.Unlock()
	return s
}

func (t *table) getEntry(n string) *tableEntry {
	(*t).eMux.Lock()
	e := (*t).entries[n]
	(*t).eMux.Unlock()
	return e
}

func (t *table) getUniqueEntry(index int) *uniqueTableEntry {
	t.uMux.Lock()
	u := t.uniqueItems[index]
	t.uMux.Unlock()
	return u
}

func (t *table) getGroupedEntry(index int) *groupedTableEntry {
	t.gMux.Lock()
	g := t.groupedItems[index]
	t.gMux.Unlock()
	return g
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//   tableEntry   ///////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func (t *tableEntry) getKey() string {
	return t.key
}

func (t *tableEntry) getPersistFile() int {
	return t.persistFile
}

func (t *tableEntry) getPersistIndex() int {
	return t.persistIndex
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//   Misc methods   /////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func makeEntryMap(entry []interface{}, schema tableSchema, sel []string) map[string]interface{} {
	entryMap := make(map[string]interface{})
	selLen := len(sel)
	for k, v := range schema.items {
		if selLen > 0 {
			for i := 0; i < selLen; i++ {
				if sel[i] == k {
					entryMap[k] = entry[v.index]
				}
			}
		} else {
			entryMap[k] = entry[v.index]
		}
	}
	return entryMap
}
