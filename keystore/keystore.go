package keystore

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/storage"
	"github.com/hewiefreeman/GopherDB/schema"
	"sync/atomic"
	"sync"
)

var (
	storesMux      sync.Mutex
	stores         map[string]*Keystore = make(map[string]*Keystore)
)

type Keystore struct {
	fileOn    uint16 // locked by eMux - placed for memory efficiency

	// Settings and schema - read only
	memOnly       bool // Store data in memory only (overrides dataOnDrive)
	dataOnDrive   bool // when true, entry data is not stored in memory, only indexing
	persistName   string // table's logger/persist folder name
	schema        *schema.Schema // table's schema

	// Atomic changable settings values - 99% read
	partitionMax  atomic.Value // *uint16* maximum entries per data file
	maxEntries    atomic.Value // *uint64* maximum amount of entries in the AuthTable
	encryptCost   atomic.Value // *int* encryption cost of encrypted items

	// entries
	eMux      sync.Mutex // entries/altLogins map lock
	entries   map[string]*KeystoreEntry // AuthTable uses a Map for storage since it's only look-up is with user name and password

	// unique values
	uMux       sync.Mutex
	uniqueVals map[string]map[interface{}]bool
}

type KeystoreEntry struct {
	persistFile  uint16
	persistIndex uint16

	mux  sync.Mutex
	data []interface{}
}

// File/folder prefixes
const (
	dataFolderPrefix = "KS-"
)

// New creates a new Keystore with the provided name, schema, and other parameters.
func New(name string, s *schema.Schema, fileOn uint16, dataOnDrive bool, memOnly bool) (*Keystore, int) {
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
	t := Keystore{
		persistName:   namePre,
		memOnly:       memOnly,
		dataOnDrive:   dataOnDrive,
		schema:        s,
		entries:       make(map[string]*KeystoreEntry),
		uniqueVals:    make(map[string]map[interface{}]bool),
		fileOn:        fileOn,
	}

	// set defaults
	t.partitionMax.Store(helpers.DefaultPartitionMax)
	t.maxEntries.Store(helpers.DefaultMaxEntries)
	t.encryptCost.Store(helpers.DefaultEncryptCost)

	// push to stores map
	storesMux.Lock()
	stores[name] = &t
	storesMux.Unlock()

	return &t, 0
}

// Delete deletes a Keystore with the given name.
func Delete(name string) int {
	if len(name) == 0 {
		return helpers.ErrorTableNameRequired
	} else if Get(name) == nil {
		return helpers.ErrorTableDoesntExist
	}

	storesMux.Lock()
	delete(stores, name)
	storesMux.Unlock()

	// !!!!!! delete data folder from system & delete log file & update config file

	return 0
}

// Get retrieves a Keystore by name
func Get(name string) *Keystore {
	storesMux.Lock()
	k := stores[name]
	storesMux.Unlock()

	return k
}

func (k *Keystore) Get(key string) (*KeystoreEntry, int) {
	// key is required
	if len(key) == 0 {
		return nil, helpers.ErrorKeyRequired
	}

	// Find entry
	k.eMux.Lock()
	e := k.entries[key]
	k.eMux.Unlock()

	if e == nil {
		return nil, helpers.ErrorNoEntryFound
	}

	return e, 0
}

func (k *Keystore) Size() int {
	k.eMux.Lock()
	s := len(k.entries)
	k.eMux.Unlock()
	return s
}

func (k *Keystore) SetEncryptionCost(cost int) {
	if cost > helpers.EncryptCostMax {
		cost = helpers.EncryptCostMax
	} else if cost < helpers.EncryptCostMin {
		cost = helpers.EncryptCostMin
	}
	k.encryptCost.Store(cost)
}

func (k *Keystore) SetMaxEntries(max uint64) {
	if max < 0 {
		max = 0
	}
	k.maxEntries.Store(max)
}

func (k *Keystore) SetPartitionMax(max uint16) {
	if max < helpers.PartitionMin {
		max = helpers.DefaultPartitionMax
	}
	k.partitionMax.Store(max)
}