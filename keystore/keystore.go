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

// Keystore
type Keystore struct {
	fileOn    uint16 // locked by eMux - placed for memory efficiency

	// Settings and schema - read only
	memOnly       bool // Store data in memory only (overrides dataOnDrive)
	dataOnDrive   bool // when true, entry data is not stored in memory, only indexing
	name   string // table's logger/persist folder name
	schema        *schema.Schema // table's schema
	configFile *os.File // configuration file

	// Atomic changable settings values - 99% read
	partitionMax  atomic.Value // *uint16* maximum entries per data file
	maxEntries    atomic.Value // *uint64* maximum amount of entries in the AuthTable
	encryptCost   atomic.Value // *int* encryption cost of encrypted items

	// entries
	eMux       sync.Mutex // entries/configFile lock
	entries    map[string]*keystoreEntry // Keystore map

	// unique values
	uMux       sync.Mutex
	uniqueVals map[string]map[interface{}]bool
}

type keystoreEntry struct {
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

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Keystore   //////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

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

		// Create/open config file
		var err error
		configFile, err = os.OpenFile(namePre + helpers.FileTypeConfig, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, helpers.ErrorFileOpen
		}

		// Write config file
		if wErr := writeConfigFile(configFile, keystoreConfig{
			Name: name,
			Schema: s.MakeConfig(),
			FileOn: fileOn,
			DataOnDrive: dataOnDrive,
			MemOnly: memOnly,
			PartitionMax: helpers.DefaultPartitionMax,
			EncryptCost: helpers.DefaultEncryptCost,
			MaxEntries: helpers.DefaultMaxEntries,
		}; wErr != 0) {
			return nil, err
		}
	}

	// make table
	t := Keystore{
		name:          name,
		memOnly:       memOnly,
		dataOnDrive:   dataOnDrive,
		schema:        s,
		configFile:    configFile,
		entries:       make(map[string]*keystoreEntry),
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

func (k *Keystore) Close(save bool) {
	if save {
		k.eMux.Lock()
		fileOn := k.fileOn
		k.eMux.Unlock()
		writeConfigFile(k.configFile, keystoreConfig{
			Name: k.name,
			Schema: k.schema.MakeConfig(),
			FileOn: fileOn,
			DataOnDrive: k.dataOnDrive,
			MemOnly: k.memOnly,
			PartitionMax: k.partitionMax.Load().(uint16),
			EncryptCost: k.encryptCost.Load().(int),
			MaxEntries: k.maxEntries.Load().(uint64),
		})
	}

	storesMux.Lock()
	stores[k.name] = nil
	delete(stores, k.name)
	storesMux.Unlock()
}

// Delete deletes a Keystore with the given name.
func (k *Keystore) Delete() int {
	k.Close(false)

	// Delete data directory
	if err := os.RemoveAll(dataFolderPrefix + k.name); err != nil {
		return helpers.ErrorFileDelete
	}

	// Delete config file
	if err := os.Remove(dataFolderPrefix + k.name + helpers.FileTypeConfig); err != nil {
		return helpers.ErrorFileDelete
	}

	return 0
}

// Get retrieves a *keystoreEntry by it's key from the Keystore
func (k *Keystore) Get(key string) (*keystoreEntry, int) {
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

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Keystore Setters   //////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// SetEncryptionCost sets the bcrypt encrytion cost
func (k *Keystore) SetEncryptionCost(cost int) int {
	if cost > helpers.EncryptCostMax {
		cost = helpers.EncryptCostMax
	} else if cost < helpers.EncryptCostMin {
		cost = helpers.EncryptCostMin
	}

	// Write to configFile
	k.eMux.Lock()
	fileOn := k.fileOn
	k.eMux.Unlock()
	if err := writeConfigFile(k.configFile, keystoreConfig{
		Name: k.name,
		Schema: k.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: k.dataOnDrive,
		MemOnly: k.memOnly,
		PartitionMax: k.partitionMax.Load().(uint16),
		EncryptCost: cost,
		MaxEntries: k.maxEntries.Load().(uint64),
	}; err != 0) {
		return err
	}

	k.encryptCost.Store(cost)
}

// SetMaxEntries sets the maximum entries for the Keystore
func (k *Keystore) SetMaxEntries(max uint64) int {

	// Write to configFile
	k.eMux.Lock()
	fileOn := k.fileOn
	k.eMux.Unlock()
	if err := writeConfigFile(k.configFile, keystoreConfig{
		Name: k.name,
		Schema: k.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: k.dataOnDrive,
		MemOnly: k.memOnly,
		PartitionMax: k.partitionMax.Load().(uint16),
		EncryptCost: k.encryptCost.Load().(int),
		MaxEntries: max,
	}; err != 0) {
		return err
	}

	k.maxEntries.Store(max)
}

// SetPartitionMax sets the maximum entries stored in a data file
func (k *Keystore) SetPartitionMax(max uint16) int {
	if max < helpers.PartitionMin {
		max = helpers.DefaultPartitionMax
	}

	// Write to configFile
	k.eMux.Lock()
	fileOn := k.fileOn
	k.eMux.Unlock()
	if err := writeConfigFile(k.configFile, keystoreConfig{
		Name: k.name,
		Schema: k.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: k.dataOnDrive,
		MemOnly: k.memOnly,
		PartitionMax: max,
		EncryptCost: k.encryptCost.Load().(int),
		MaxEntries: k.maxEntries.Load().(uint64),
	}; err != 0) {
		return err
	}

	k.partitionMax.Store(max)
}

// Writes k to f and truncates file
func writeConfigFile(f *os.File, k keystoreConfig) int {
	jBytes, jErr := json.Marshal(k)
	if jErr != nil {
		return helpers.ErrorJsonEncoding
	}

	// Write to config file
	if _, wErr := f.WriteAt(jBytes, 0); wErr != nil {
		return helpers.ErrorFileUpdate
	}
	k.configFile.Truncate(int64(len(jBytes)))
	return 0
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Keystore Restoring   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

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
		f.Close()
		return nil, helpers.ErrorFileOpen
	}

	// Get file bytes
	bytes := make([]byte, fs.Size())
	_, rErr := f.ReadAt(bytes, 0)
	if rErr != nil && rErr != io.EOF {
		f.Close()
		return nil, helpers.ErrorFileRead
	}

	// Make confStruct from json bytes
	var confStruct keystoreConfig
	mErr := json.Unmarshal(bytes, &confStruct)
	if mErr != nil {
		f.Close()
		return nil, helpers.ErrorJsonDecoding
	}

	// Make schema with the schemaList
	s, schemaErr := schema.Restore(confStruct.Schema)
	if schemaErr != 0 {
		f.Close()
		return nil, schemaErr
	}

	ks, ksErr := New(name, f, s, confStruct.FileOn, confStruct.DataOnDrive, confStruct.MemOnly)
	if ksErr != 0 {
		f.Close()
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
		df.Close()
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
			dataFile.Close()
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
					if resErr := ks.RestoreKey(eKey, eData, uint16(fileNum), uint16(lineOn)); resErr != 0 {
						fmt.Println("Error restoring '"+eKey+"' on line", lineOn, "in file", fileStats.Name(), "with error:", resErr)
					}
				}

				lineByteStart = i+1
				lineOn++
			}
		}

		//
		dataFile.Close()
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