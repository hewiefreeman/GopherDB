package keystore

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/storage"
	"github.com/hewiefreeman/GopherDB/schema"
	"strings"
	"strconv"
	"sync/atomic"
	"sync"
	"os"
	"io"
	"encoding/json"
	"fmt"
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
	name   string // table's logger/persist folder name
	schema        *schema.Schema // table's schema

	// Atomic changable settings values - 99% read
	partitionMax  atomic.Value // *uint16* maximum entries per data file
	maxEntries    atomic.Value // *uint64* maximum amount of entries in the AuthTable
	encryptCost   atomic.Value // *int* encryption cost of encrypted items

	// entries
	eMux       sync.Mutex // entries/configFile lock
	configFile *os.File // configuration file
	entries    map[string]*KeystoreEntry // AuthTable uses a Map for storage since it's only look-up is with user name and password

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

type keystoreConfig struct {
	Name string
	Schema []schema.SchemaConfigItem
	FileOn uint16
	DataOnDrive bool
	MemOnly bool
	PartitionMax uint16
	EncryptCost int
	MaxEntries uint64
}

// File/folder prefixes
const (
	dataFolderPrefix = "KS-"
)

// New creates a new Keystore with the provided name, schema, and other parameters.
func New(name string, configFile *os.File, s *schema.Schema, fileOn uint16, dataOnDrive bool, memOnly bool) (*Keystore, int) {
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

	// table name with prefix
	namePre := dataFolderPrefix + name

	// Restoring if configFile is not nil
	if configFile == nil {
		// Make table storage folder
		mkErr := storage.MakeDir(namePre)
		if mkErr != nil {
			return nil, helpers.ErrorCreatingFolder
		}

		// Make json bytes for config file
		jBytes, jErr := json.Marshal(keystoreConfig{
			Name: name,
			Schema: s.MakeConfig(),
			FileOn: fileOn,
			DataOnDrive: dataOnDrive,
			MemOnly: memOnly,
			PartitionMax: helpers.DefaultPartitionMax,
			EncryptCost: helpers.DefaultEncryptCost,
			MaxEntries: helpers.DefaultMaxEntries,
		})
		if jErr != nil {
			return nil, helpers.ErrorJsonEncoding
		}

		// Save config file
		var err error
		configFile, err = os.OpenFile(namePre + helpers.FileTypeConfig, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, helpers.ErrorFileOpen
		}

		if _, err = configFile.Write(jBytes); err != nil {
			return nil, helpers.ErrorFileWrite
		}

		configFile.Truncate(int64(len(jBytes)))
	}

	// make table
	t := Keystore{
		name:          name,
		memOnly:       memOnly,
		dataOnDrive:   dataOnDrive,
		schema:        s,
		configFile:    configFile,
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

// Get retrieves a Keystore by name
func Get(name string) *Keystore {
	storesMux.Lock()
	k := stores[name]
	storesMux.Unlock()

	return k
}

func (k *Keystore) Close() {
	storesMux.Lock()
	stores[k.name] = nil
	delete(stores, k.name)
	storesMux.Unlock()

	k.eMux.Lock()
	k.uMux.Lock()
	k.entries = nil
	k.uniqueVals = nil
	k.configFile.Close()
	k.uMux.Unlock()
	k.eMux.Unlock()
}

// Delete deletes a Keystore with the given name.
func (k *Keystore) Delete() int {
	k.Close()

	// !!!!!! delete data folder & config file

	return 0
}

// Get retrieves a *KeystoreEntry by it's key from the Keystore
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

// Size returns the number of entries in the Keystore
func (k *Keystore) Size() int {
	k.eMux.Lock()
	s := len(k.entries)
	k.eMux.Unlock()
	return s
}

// SetEncryptionCost sets the bcrypt encrytion cost
func (k *Keystore) SetEncryptionCost(cost int) {
	if cost > helpers.EncryptCostMax {
		cost = helpers.EncryptCostMax
	} else if cost < helpers.EncryptCostMin {
		cost = helpers.EncryptCostMin
	}

	// Save in config file !!!

	k.encryptCost.Store(cost)
}

// SetMaxEntries sets the maximum entries for the Keystore
func (k *Keystore) SetMaxEntries(max uint64) {

	// Save in config file !!!

	k.maxEntries.Store(max)
}

// SetPartitionMax sets the maximum entries stored in a data file
func (k *Keystore) SetPartitionMax(max uint16) {
	if max < helpers.PartitionMin {
		max = helpers.DefaultPartitionMax
	}

	// Save in config file !!!

	k.partitionMax.Store(max)
}

// Restore restores a Keystore by name; requires a valid config file and data folder
func Restore(name string) (*Keystore, int) {
	namePre := dataFolderPrefix + name

	// Open the File
	f, err := os.OpenFile(namePre + helpers.FileTypeConfig, os.O_RDWR, 0755)
	if err != nil {
		return nil, helpers.ErrorFileOpen
	}

	// Get file stats
	fs, fsErr := f.Stat()
	if fsErr != nil {
		return nil, helpers.ErrorFileOpen
	}

	// Get file bytes
	bytes := make([]byte, fs.Size())
	_, rErr := f.ReadAt(bytes, 0)
	if rErr != nil && rErr != io.EOF {
		return nil, helpers.ErrorFileRead
	}

	// Make confStruct from json bytes
	var confStruct keystoreConfig
	mErr := json.Unmarshal(bytes, &confStruct)
	if mErr != nil {
		return nil, helpers.ErrorJsonDecoding
	}

	// Make schema with the schemaList
	s, schemaErr := schema.Restore(confStruct.Schema)
	if schemaErr != 0 {
		return nil, schemaErr
	}

	ks, ksErr := New(name, f, s, confStruct.FileOn, confStruct.DataOnDrive, confStruct.MemOnly)
	if ksErr != 0 {
		return nil, ksErr
	}

	ks.eMux.Lock()
	ks.uMux.Lock()

	// Set optional settings if different from defaults
	if confStruct.EncryptCost != helpers.DefaultEncryptCost {
		ks.encryptCost.Store(confStruct.EncryptCost)
	}

	if confStruct.MaxEntries != helpers.DefaultMaxEntries {
		ks.maxEntries.Store(confStruct.MaxEntries)
	}

	if confStruct.PartitionMax != helpers.DefaultPartitionMax {
		ks.partitionMax.Store(confStruct.PartitionMax)
	}

	// Load data/indexing into memory...

	// Open data folder
	df, err := os.Open(namePre)
	if err != nil {
		ks.eMux.Unlock()
		ks.uMux.Unlock()
		ks.Close()
		return nil, helpers.ErrorFileOpen
	}
	files, err := df.Readdir(-1)
	df.Close()
	if err != nil {
		ks.eMux.Unlock()
		ks.uMux.Unlock()
		ks.Close()
		return nil, helpers.ErrorFileRead
	}

	// Go through files
	for _, fileStats := range files {
		// Get file number
		fileNameSplit := strings.Split(fileStats.Name(), ".")
		fileNum, fnErr := strconv.Atoi(fileNameSplit[0])
		if fnErr != nil || len(fileNameSplit) < 2 || "."+fileNameSplit[1] != helpers.FileTypeStorage {
			fmt.Println("'"+fileStats.Name()+"' is not a valid storage file.")
			continue
		}

		// Open data file
		dataFile, err := os.OpenFile(namePre + "/" + fileStats.Name(), os.O_RDWR, 0755)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Get file bytes
		fb := make([]byte, fileStats.Size())
		_, rErr := dataFile.ReadAt(fb, 0)
		if rErr != nil && rErr != io.EOF {
			fmt.Println(rErr)
			continue
		}

		// Go through file bytes and restore entries
		lineOn := 1
		lineByteStart := 0
		for i, b := range fb {
			if b == 10 {
				// Restore line
				if eKey, eData := restoreDataLine(fb[lineByteStart:i]); eData != nil {
					if resErr := ks.Restore(eKey, eData, uint16(fileNum), uint16(lineOn)); resErr != 0 {
						fmt.Println("Error restoring '"+eKey+"' on line", lineOn, "in file", fileStats.Name(), "with error:", resErr)
					}
				}

				lineByteStart = i+1
				lineOn++
			}
		}
	}
	ks.uMux.Unlock()
	ks.eMux.Unlock()

	return ks, 0
}

func restoreDataLine(line []byte) (string, []interface{}) {
	var jEntry jsonEntry
	mErr := json.Unmarshal(line, &jEntry)
	if mErr != nil {
		return "", nil
	}

	if jEntry.D == nil || jEntry.K == "" {
		return "", nil
	}

	return jEntry.K, jEntry.D
}