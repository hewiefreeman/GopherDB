package ggdb

import (
	"sync"
	"time"
)

var (
	configFile string = "db.conf"

	statusMux sync.Mutex
	dbStatus  int
	replica   bool
	readOnly  bool

	replicasMux sync.Mutex
	replicas    []string = []string{}

	balancersMux sync.Mutex
	balancers    []string = []string{}

	tablesMux      sync.Mutex
	logPersistTime int = 30
	tables         map[string]*table = make(map[string]*table)
)

type tableSchema struct {
	items map[string]tableSchemaEntry
}

type tableSchemaEntry struct {
	index       int
	unique      bool
	grouped     bool
	groupedVals []interface{}
}

type table struct {
	logFile       string
	persistFolder string
	partitionMax  int
	maxEntries    int
	schema        tableSchema

	eMux    sync.Mutex
	entries map[string]*tableEntry

	iMux     sync.Mutex
	index    []*tableEntry

	optIMux      sync.Mutex
	lastOptIndex time.Time

	uMux       sync.Mutex
	uniqueVals map[int]map[interface{}]*tableEntry

	gMux
	groupedVals map[int]map[interface{}][]*tableEntry

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

// Database statuses
const (
	statusSettingUp = iota
	statusHealthy
	statusReplicationFailure
	statusOffline
)

// File/folder prefixes
const (
	prefixTableLogFile           = "log-"
	prefixTableDataFolder        = "data-"
	prefixTableDataPartitionFile = "part-"
)

// Defaults
const (
	defaultPartitionMax = 1500
	defaultConfig = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"balancers\":[],\"tables\":[]}"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   table   ////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func createTable(name string, maxEntries int, schema tableSchema, fileOn int, lineOn int, partitionMax int) int {
	if len(name) == 0 {
		return ErrorTableNameRequired
	} else if tableExists(name) {
		return ErrorTableExists
	}

	//default partitionMax
	if partitionMax <= 0 {
		partitionMax = defaultPartitionMax
	}

	//cant be under 0
	if fileOn < 0 {
		fileOn = 0
	}
	if lineOn < 0 {
		lineOn = 0
	}

	t := table{	logFile: prefixTableLogFile+name,
				persistFolder: prefixTableDataFolder+name,
				partitionMax: partitionMax,
				maxEntries: maxEntries,
				schema: schema,
				entries: make(map[string]*tableEntry),
				index: []*tableEntry{},
				lastOptIndex: time.Now(),
				uniqueVals: make(map[int]map[interface{}]*tableEntry),
				groupedVals: make(map[int]map[interface{}][]*tableEntry),
				fileOn: fileOn,
				lineOn: lineOn }

	tablesMux.Lock()
	tables[name] = &t
	tablesMux.Unlock()

	// !!!!!! make new folder on system for persisting data & create log file & update config file

	return 0
}

func deleteTable(name string) int {
	if len(name) == 0 {
		return ErrorTableNameRequired
	} else if !tableExists(name) {
		return ErrorTableDoesntExist
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

/////////////////////////////////////////////////////////////////////////////////////////////////
//   table index   //////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func (t *table) optimizeIndex() {
	(*t).iMux.Lock()
	(*t).eMux.Lock()
	(*t).index = make([]*tableEntry, len((*t).entries), len((*t).entries))
	i := 0
	for _, e := range (*t).entries {
		(*t).index[i] = e
		i++
	}
	(*t).optIMux.Lock()
	(*t).lastOptIndex = time.Now()
	(*t).optIMux.Unlock()
	(*t).eMux.Unlock()
	(*t).iMux.Unlock()
}

func (t *table) lastIndexOptimizationTime() time.Time {
	(*t).optIMux.Lock()
	i := (*t).lastOptIndex
	(*t).optIMux.Unlock()
	return i
}

func (t *table) addToIndex(v *tableEntry) {
	(*t).iMux.Lock()
	(*t).index = append((*t).index, v)
	(*t).iMux.Unlock()
}

func (t *table) removeFromIndex(v *tableEntry) {
	(*t).iMux.Lock()
	for i := 0; i < len((*t).index); i++ {
		if (*t).index[i] == v {
			(*t).index[i] = nil
		}
	}
	(*t).iMux.Unlock()
}

func (t *table) indexSize() int {
	(*t).iMux.Lock()
	s := len((*t).index)
	(*t).iMux.Unlock()
	return s
}

func (t *table) indexChunk(start int, amount int) ([]*tableEntry, int) {
	if start < 0 || amount <= 0 {
		return []*tableEntry{}, ErrorIndexChunkOutOfRange
	}
	max := start+amount;

	(*t).iMux.Lock()
	indexLen := len((*t).index)
	if start >= indexLen {
		start = indexLen-1
		max = indexLen
	} else if max > indexLen {
		max = indexLen
	}
	c := (*t).index[start:max]
	(*t).iMux.Unlock()

	return c, 0
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//   tableSchema   //////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func newTableSchema(items []string, unique []int, grouped map[int][]interface{}) (tableSchema, int) {
	if len(items) == 0 {
		return tableSchema{}, ErrorSchemaItemsRequired
	}
	var s tableSchema
	s.items = make(map[string]tableSchemaEntry)
	// Go through items
	for i := 0; i < len(items); i++ {
		se := tableSchemaEntry{index: i}
		// check for unique
		for j := 0; j < len(unique); j++ {
			if unique[j] == i {
				se.unique = true
			}
		}
		// check for grouped
		if !se.unique {
			if v, ok := grouped[i]; ok {
				// check if any of the values aren't hashable
				for j := 0; j < len(v); j++ {
					if !isHashable(v[j]) {
						return tableSchema{}, ErrorUnhashableGroupValue
					}
				}
				se.grouped = true
				se.groupedVals = v
			}
		}
		s.items[items[i]] = se
	}
	return s, 0
}
