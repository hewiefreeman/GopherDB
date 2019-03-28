package tables

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

var (
	tablesMux      sync.Mutex
	logPersistTime int = 30
	tables         map[string]*Table = make(map[string]*Table)
)

type Table struct {
	logFolder     string
	persistFolder string
	partitionMax  int
	maxEntries    int
	indexChunks   int
	schema        TableSchema

	eMux    sync.Mutex
	entries map[string]*TableEntry

	iMux  sync.Mutex
	index []*indexChunk

	uMux       sync.Mutex
	uniqueItems map[int]*UniqueTableItem

	gMux        sync.Mutex
	groupedItems map[int]*GroupedTableEntry

	pMux   sync.Mutex
	fileOn int
	lineOn int
}

type TableEntry struct {
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
	defaultIndexChunks  = 10
	defaultConfig       = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"balancers\":[],\"Tables\":[]}"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   table   ////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func New(name string, schema TableSchema, maxEntries int, indexChunks int, partitionMax int, fileOn int, lineOn int) int {
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
		ic := indexChunk{entries: []*TableEntry{}}
		indexChunkList[i] = &ic
	}

	// make table
	t := Table{	logFolder: prefixTableLogging+name,
				persistFolder: prefixTableDataFolder+name,
				partitionMax: partitionMax,
				maxEntries: maxEntries,
				indexChunks: indexChunks,
				schema: schema,
				entries: make(map[string]*TableEntry),
				index: indexChunkList,
				uniqueItems: make(map[int]*UniqueTableItem),
				groupedItems: make(map[int]*GroupedTableEntry),
				fileOn: fileOn,
				lineOn: lineOn }

	// Apply schema
	for _, v := range schema.items {
		if v.unique {
			ute := UniqueTableItem{vals: make(map[interface{}]*TableEntry)}
			t.uniqueItems[v.index] = &ute
		} else if v.grouped {
			gte := GroupedTableEntry{groups: make(map[interface{}]*TableEntryGroup)}
			for i := 0; i < len(v.groupedVals); i++ {
				teg := TableEntryGroup{entries: []*TableEntry{}}
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

func Delete(name string) int {
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

func Get(n string) *Table {
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

func (t *Table) Size() int {
	(*t).eMux.Lock()
	s := len((*t).entries)
	(*t).eMux.Unlock()
	return s
}

func (t *Table) Get(n string) *TableEntry {
	(*t).eMux.Lock()
	e := (*t).entries[n]
	(*t).eMux.Unlock()
	return e
}

func (t *Table) GetUnique(index int) *UniqueTableItem {
	t.uMux.Lock()
	u := t.uniqueItems[index]
	t.uMux.Unlock()
	return u
}

func (t *Table) GetGrouped(index int) *GroupedTableEntry {
	t.gMux.Lock()
	g := t.groupedItems[index]
	t.gMux.Unlock()
	return g
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//   TableEntry   ///////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func (t *TableEntry) Key() string {
	return t.key
}

func (t *TableEntry) PersistFile() int {
	return t.persistFile
}

func (t *TableEntry) PersistIndex() int {
	return t.persistIndex
}
