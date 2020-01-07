/*
keystore package Copyright 2020 Dominique Debergue

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific
language governing permissions and limitations under the License.
*/

package keystore

import (

	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"github.com/schollz/progressbar"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"encoding/json"
	"fmt"
)

// File/folder prefixes
const (
	dataFolderPrefix = "KS-"
)

var (
	storesMux sync.Mutex
	stores    map[string]*Keystore = make(map[string]*Keystore)
)

// Keystore
type Keystore struct {
	fileOn uint32 // locked by eMux - placed for memory efficiency

	// Settings and schema - read only
	// Changing some of these setting requires a reformat and/or restore of the Keystore
	memOnly     bool          // Store data in memory only (overrides dataOnDrive)
	dataOnDrive bool          // when true, entry data is not stored in memory, only indexing
	name        string        // table's logger/persist folder name
	schema      schema.Schema // table's schema
	configFile  *os.File      // configuration file

	// Atomic changeable settings values - 99% read
	partitionMax atomic.Value // *uint16* maximum entries per data file
	maxEntries   atomic.Value // *uint64* maximum amount of entries in the AuthTable
	encryptCost  atomic.Value // *int* encryption cost of encrypted items

	// entries
	eMux    sync.Mutex                // entries/configFile lock
	entries map[string]*keystoreEntry // Keystore map

	// unique values
	uMux       sync.Mutex
	uniqueVals map[string]map[interface{}]bool
}

type keystoreEntry struct {
	persistFile  uint32
	persistIndex uint16

	mux  sync.Mutex
	data []interface{}
}

type keystoreConfig struct {
	Name         string
	Schema       []schema.SchemaConfigItem
	FileOn       uint32
	DataOnDrive  bool
	MemOnly      bool
	PartitionMax uint16
	EncryptCost  int
	MaxEntries   uint64
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Keystore   //////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// New creates a new Keystore with the provided name, schema, and other parameters.
func New(name string, configFile *os.File, s schema.Schema, fileOn uint32, dataOnDrive bool, memOnly bool) (*Keystore, int) {
	if len(name) == 0 {
		return nil, helpers.ErrorTableNameRequired
	} else if Get(name) != nil {
		return nil, helpers.ErrorTableExists
	} else if !s.Validate() {
		return nil, helpers.ErrorSchemaInvalid
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
		configFile, err = os.OpenFile(namePre+helpers.FileTypeConfig, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, helpers.ErrorFileOpen
		}

		// Write config file
		if wErr := writeConfigFile(configFile, keystoreConfig{
			Name:         name,
			Schema:       s.MakeConfig(),
			FileOn:       fileOn,
			DataOnDrive:  dataOnDrive,
			MemOnly:      memOnly,
			PartitionMax: helpers.DefaultPartitionMax,
			EncryptCost:  helpers.DefaultEncryptCost,
			MaxEntries:   helpers.DefaultMaxEntries,
		}); wErr != 0 {
			return nil, wErr
		}
	}

	// make table
	t := Keystore{
		name:        name,
		memOnly:     memOnly,
		dataOnDrive: dataOnDrive,
		schema:      s,
		configFile:  configFile,
		entries:     make(map[string]*keystoreEntry),
		uniqueVals:  make(map[string]map[interface{}]bool),
		fileOn:      fileOn,
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
	if len(name) == 0 {
		return nil
	}

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
			Name:         k.name,
			Schema:       k.schema.MakeConfig(),
			FileOn:       fileOn,
			DataOnDrive:  k.dataOnDrive,
			MemOnly:      k.memOnly,
			PartitionMax: k.partitionMax.Load().(uint16),
			EncryptCost:  k.encryptCost.Load().(int),
			MaxEntries:   k.maxEntries.Load().(uint64),
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

// Size returns the number of entries in the Keystore
func (k *Keystore) MemOnly() bool {
	return k.memOnly
}

// Size returns the number of entries in the Keystore
func (k *Keystore) DataOnDrive() bool {
	return k.dataOnDrive
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
		Name:         k.name,
		Schema:       k.schema.MakeConfig(),
		FileOn:       fileOn,
		DataOnDrive:  k.dataOnDrive,
		MemOnly:      k.memOnly,
		PartitionMax: k.partitionMax.Load().(uint16),
		EncryptCost:  cost,
		MaxEntries:   k.maxEntries.Load().(uint64),
	}); err != 0 {
		return err
	}

	k.encryptCost.Store(cost)

	return 0
}

// SetMaxEntries sets the maximum entries for the Keystore
func (k *Keystore) SetMaxEntries(max uint64) int {

	// Write to configFile
	k.eMux.Lock()
	fileOn := k.fileOn
	k.eMux.Unlock()
	if err := writeConfigFile(k.configFile, keystoreConfig{
		Name:         k.name,
		Schema:       k.schema.MakeConfig(),
		FileOn:       fileOn,
		DataOnDrive:  k.dataOnDrive,
		MemOnly:      k.memOnly,
		PartitionMax: k.partitionMax.Load().(uint16),
		EncryptCost:  k.encryptCost.Load().(int),
		MaxEntries:   max,
	}); err != 0 {
		return err
	}

	k.maxEntries.Store(max)

	return 0
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
		Name:         k.name,
		Schema:       k.schema.MakeConfig(),
		FileOn:       fileOn,
		DataOnDrive:  k.dataOnDrive,
		MemOnly:      k.memOnly,
		PartitionMax: max,
		EncryptCost:  k.encryptCost.Load().(int),
		MaxEntries:   k.maxEntries.Load().(uint64),
	}); err != 0 {
		return err
	}

	k.partitionMax.Store(max)

	return 0
}

// Writes k to f and truncates file
func writeConfigFile(f *os.File, k keystoreConfig) int {
	jBytes, jErr := json.MarshalIndent(k, "", "\t")
	if jErr != nil {
		return helpers.ErrorJsonEncoding
	}

	// Write to config file
	if _, wErr := f.WriteAt(jBytes, 0); wErr != nil {
		return helpers.ErrorFileUpdate
	}
	f.Truncate(int64(len(jBytes)))
	return 0
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Keystore Restoring   ////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Restore restores a Keystore by name; requires a valid config file and data folder
func Restore(name string) (*Keystore, int) {
	fmt.Printf("Restoring Keystore table '%v'...\n", name)
	namePre := dataFolderPrefix + name
	// Open the File
	f, err := os.OpenFile(namePre+helpers.FileTypeConfig, os.O_RDWR, 0755)
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
	// Make Keystore table
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
	// Open data folder
	df, err := os.Open(namePre)
	if err != nil {
		ks.eMux.Unlock()
		ks.uMux.Unlock()
		df.Close()
		ks.Close(false)
		return nil, helpers.ErrorFileOpen
	}
	// Get file names
	files, err := df.Readdir(-1)
	df.Close()
	if err != nil {
		ks.eMux.Unlock()
		ks.uMux.Unlock()
		ks.Close(false)
		return nil, helpers.ErrorFileRead
	}
	// Make progress bar
	pBar := progressbar.New(len(files))
	// Go through files & restor entries
	for _, fileStats := range files {
		// Get file number
		fileNameSplit := strings.Split(fileStats.Name(), ".")
		fileNum, fnErr := strconv.Atoi(fileNameSplit[0])
		if fnErr != nil || len(fileNameSplit) < 2 || "."+fileNameSplit[1] != helpers.FileTypeStorage {
			// Not a valid storage file
			pBar.Add(1)
			continue
		}
		var of *storage.OpenFile
		var err int
		if of, err = storage.GetOpenFile(namePre + "/" + fileStats.Name()); err != 0 {
			pBar.Finish()
			return nil, err
		}
		for i := 0; i < of.Lines(); i++ {
			// Get line bytes
			var lb []byte
			if lb, err = of.Read(uint16(i+1)); err != 0 {
				pBar.Finish()
				return nil, err
			}
			eKey, eData := restoreDataLine(lb)
			if eData == nil {
				pBar.Finish()
				return nil, helpers.ErrorJsonDataFormat
			}
			if err = ks.restoreKey(eKey, eData, uint32(fileNum), uint16(i+1)); err != 0 {
				pBar.Finish()
				return nil, err
			}
		}
		pBar.Add(1)
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
