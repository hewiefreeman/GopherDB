package authtable

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"sync"
	"sync/atomic"
	"os"
	"io"
	"encoding/json"
	"strings"
	"strconv"
	"fmt"
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

// Authtable
type AuthTable struct {
	fileOn    uint16 // locked by eMux - placed for memory efficiency

	// Settings and schema - read only
	memOnly       bool // Store data in memory only (overrides dataOnDrive)
	dataOnDrive   bool // when true, entry data is not stored in memory, only indexing and password
	name          string // table's logger/persist folder name
	schema        schema.Schema // table's schema
	configFile    *os.File // config file

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
	entries   map[string]*authTableEntry // AuthTable uses a Map for storage since it's only look-up is with user name and password
	altLogins map[string]*authTableEntry

	// unique values
	uMux       sync.Mutex
	uniqueVals map[string]map[interface{}]bool
}

type authTableEntry struct {
	persistFile  uint16
	persistIndex uint16

	password atomic.Value

	mux  sync.Mutex
	data []interface{}
}

type authtableConfig struct {
	Name string
	Schema []schema.SchemaConfigItem
	FileOn uint16
	DataOnDrive bool
	MemOnly bool
	PartitionMax uint16
	EncryptCost int
	MaxEntries uint64
	MinPass uint8
	PassResetLen uint8
	EmailItem string
	AltLogin string
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
//		{"NewTable": [
//			"authtable", /* or keystore, datelist, etc. */
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
//			0, false, false
//		]};
//

// New creates a new AuthTable with the provided name, schema, and other parameters.
func New(name string, configFile *os.File, s schema.Schema, fileOn uint16, dataOnDrive bool, memOnly bool) (*AuthTable, int) {
	if len(name) == 0 {
		return nil, helpers.ErrorTableNameRequired
	} else if Get(name) != nil {
		return nil, helpers.ErrorTableExists
	} else if !s.Validate() {
		return nil, helpers.ErrorTableExists
	}

	// memOnly overrides dataOnDrive
	if memOnly {
		dataOnDrive = false
	}

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
		if wErr := writeConfigFile(configFile, authtableConfig{
			Name: name,
			Schema: s.MakeConfig(),
			FileOn: fileOn,
			DataOnDrive: dataOnDrive,
			MemOnly: memOnly,
			PartitionMax: helpers.DefaultPartitionMax,
			EncryptCost: helpers.DefaultEncryptCost,
			MaxEntries: helpers.DefaultMaxEntries,
		}); wErr != 0 {
			return nil, wErr
		}
	}

	// Make table folder   & update config file !!!
	mkErr := storage.MakeDir(namePre)
	if mkErr != nil {
		return nil, helpers.ErrorCreatingFolder
	}

	// make table
	t := AuthTable{
		name:          name,
		memOnly:       memOnly,
		dataOnDrive:   dataOnDrive,
		schema:        s,
		configFile:    configFile,
		entries:       make(map[string]*authTableEntry),
		altLogins:     make(map[string]*authTableEntry),
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

func (t *AuthTable) Close(save bool) {
	if save {
		t.eMux.Lock()
		fileOn := t.fileOn
		t.eMux.Unlock()
		writeConfigFile(t.configFile, authtableConfig{
			Name: t.name,
			Schema: t.schema.MakeConfig(),
			FileOn: fileOn,
			DataOnDrive: t.dataOnDrive,
			MemOnly: t.memOnly,
			PartitionMax: t.partitionMax.Load().(uint16),
			EncryptCost: t.encryptCost.Load().(int),
			MaxEntries: t.maxEntries.Load().(uint64),
			MinPass: t.minPassword.Load().(uint8),
			PassResetLen: t.passResetLen.Load().(uint8),
			EmailItem: t.emailItem.Load().(string),
			AltLogin: t.altLoginItem.Load().(string),
		})
	}

	tablesMux.Lock()
	delete(tables, t.name)
	tablesMux.Unlock()

	t.configFile.Close()
}

// Delete deletes the AuthTable from memory and disk
func (t *AuthTable) Delete() int {
	t.Close(false)

	// Delete data directory
	if err := os.RemoveAll(dataFolderPrefix + t.name); err != nil {
		return helpers.ErrorFileDelete
	}

	// Delete config file
	if err := os.Remove(dataFolderPrefix + t.name + helpers.FileTypeConfig); err != nil {
		return helpers.ErrorFileDelete
	}

	return 0
}

// Get retrieves a AuthTable by name
func Get(name string) *AuthTable {
	tablesMux.Lock()
	t := tables[name]
	tablesMux.Unlock()

	return t
}

func (t *AuthTable) Get(userName string, password string) (*authTableEntry, int) {
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

// CheckPassword compares the authTableEntry's encrypted password with the given string password.
func (e *authTableEntry) CheckPassword(pass string) bool {
	p := e.password.Load().([]byte)
	return helpers.StringMatchesEncryption(pass, p)
}

func (t *AuthTable) Size() int {
	t.eMux.Lock()
	s := len(t.entries)
	t.eMux.Unlock()
	return s
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Authtable Setters   /////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *AuthTable) SetEncryptionCost(cost int) int {
	if cost > helpers.EncryptCostMax {
		cost = helpers.EncryptCostMax
	} else if cost < helpers.EncryptCostMin {
		cost = helpers.EncryptCostMin
	}

	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	if err := writeConfigFile(t.configFile, authtableConfig{
		Name: t.name,
		Schema: t.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: t.dataOnDrive,
		MemOnly: t.memOnly,
		PartitionMax: t.partitionMax.Load().(uint16),
		EncryptCost: cost,
		MaxEntries: t.maxEntries.Load().(uint64),
		MinPass: t.minPassword.Load().(uint8),
		PassResetLen: t.passResetLen.Load().(uint8),
		EmailItem: t.emailItem.Load().(string),
		AltLogin: t.altLoginItem.Load().(string),
	}); err != 0 {
		return err
	}

	t.encryptCost.Store(cost)
	return 0
}

func (t *AuthTable) SetMaxEntries(max uint64) int {
	if max < 0 {
		max = 0
	}

	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	if err := writeConfigFile(t.configFile, authtableConfig{
		Name: t.name,
		Schema: t.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: t.dataOnDrive,
		MemOnly: t.memOnly,
		PartitionMax: t.partitionMax.Load().(uint16),
		EncryptCost: t.encryptCost.Load().(int),
		MaxEntries: max,
		MinPass: t.minPassword.Load().(uint8),
		PassResetLen: t.passResetLen.Load().(uint8),
		EmailItem: t.emailItem.Load().(string),
		AltLogin: t.altLoginItem.Load().(string),
	}); err != 0 {
		return err
	}

	t.maxEntries.Store(max)
	return 0
}

func (t *AuthTable) SetMinPasswordLength(min uint8) int {
	if min < 1 {
		min = 1
	}
	if t.passResetLen.Load().(uint8) < min {
		t.passResetLen.Store(min)
	}

	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	if err := writeConfigFile(t.configFile, authtableConfig{
		Name: t.name,
		Schema: t.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: t.dataOnDrive,
		MemOnly: t.memOnly,
		PartitionMax: t.partitionMax.Load().(uint16),
		EncryptCost: t.encryptCost.Load().(int),
		MaxEntries: t.maxEntries.Load().(uint64),
		MinPass: min,
		PassResetLen: t.passResetLen.Load().(uint8),
		EmailItem: t.emailItem.Load().(string),
		AltLogin: t.altLoginItem.Load().(string),
	}); err != 0 {
		return err
	}

	t.minPassword.Store(min)
	return 0
}

func (t *AuthTable) SetPasswordResetLength(len uint8) int {
	mLen := t.minPassword.Load().(uint8)
	if len < mLen {
		len = mLen
	}

	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	if err := writeConfigFile(t.configFile, authtableConfig{
		Name: t.name,
		Schema: t.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: t.dataOnDrive,
		MemOnly: t.memOnly,
		PartitionMax: t.partitionMax.Load().(uint16),
		EncryptCost: t.encryptCost.Load().(int),
		MaxEntries: t.maxEntries.Load().(uint64),
		MinPass: t.minPassword.Load().(uint8),
		PassResetLen: len,
		EmailItem: t.emailItem.Load().(string),
		AltLogin: t.altLoginItem.Load().(string),
	}); err != 0 {
		return err
	}

	t.passResetLen.Store(len)
	return 0
}

// SetAltLoginItem sets the AuthTable's alternative login item. Item must be a string and unique.
func (t *AuthTable) SetAltLoginItem(item string) int {
	si := t.schema[item]
	if !si.QuickValidate() {
		return helpers.ErrorInvalidItem
	} else if si.TypeName() != schema.ItemTypeString || !si.Unique() || !si.Required() {
		return helpers.ErrorInvalidItem
	}

	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	if err := writeConfigFile(t.configFile, authtableConfig{
		Name: t.name,
		Schema: t.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: t.dataOnDrive,
		MemOnly: t.memOnly,
		PartitionMax: t.partitionMax.Load().(uint16),
		EncryptCost: t.encryptCost.Load().(int),
		MaxEntries: t.maxEntries.Load().(uint64),
		MinPass: t.minPassword.Load().(uint8),
		PassResetLen: t.passResetLen.Load().(uint8),
		EmailItem: t.emailItem.Load().(string),
		AltLogin: item,
	}); err != 0 {
		return err
	}

	t.altLoginItem.Store(item)
	return 0
}

// SetAltLoginItem sets the AuthTable's email item. Item must be a string and unique.
func (t *AuthTable) SetEmailItem(item string) int {
	si := t.schema[item]
	if !si.QuickValidate() {
		return helpers.ErrorInvalidItem
	} else if si.TypeName() != schema.ItemTypeString || !si.Unique() {
		return helpers.ErrorInvalidItem
	}

	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	if err := writeConfigFile(t.configFile, authtableConfig{
		Name: t.name,
		Schema: t.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: t.dataOnDrive,
		MemOnly: t.memOnly,
		PartitionMax: t.partitionMax.Load().(uint16),
		EncryptCost: t.encryptCost.Load().(int),
		MaxEntries: t.maxEntries.Load().(uint64),
		MinPass: t.minPassword.Load().(uint8),
		PassResetLen: t.passResetLen.Load().(uint8),
		EmailItem: item,
		AltLogin: t.altLoginItem.Load().(string),
	}); err != 0 {
		return err
	}

	t.emailItem.Store(item)
	return 0
}

func (t *AuthTable) SetPartitionMax(max uint16) int {
	if max < helpers.PartitionMin {
		max = helpers.DefaultPartitionMax
	}

	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	if err := writeConfigFile(t.configFile, authtableConfig{
		Name: t.name,
		Schema: t.schema.MakeConfig(),
		FileOn: fileOn,
		DataOnDrive: t.dataOnDrive,
		MemOnly: t.memOnly,
		PartitionMax: max,
		EncryptCost: t.encryptCost.Load().(int),
		MaxEntries: t.maxEntries.Load().(uint64),
		MinPass: t.minPassword.Load().(uint8),
		PassResetLen: t.passResetLen.Load().(uint8),
		EmailItem: t.emailItem.Load().(string),
		AltLogin: t.altLoginItem.Load().(string),
	}); err != 0 {
		return err
	}

	t.partitionMax.Store(max)
	return 0
}

func writeConfigFile(f *os.File, c authtableConfig) int {
	jBytes, jErr := json.Marshal(c)
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
//   AuthTable Restoring   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Restore restores an AuthTable by name; requires a valid config file and data folder
func Restore(name string) (*AuthTable, int) {
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
	var confStruct authtableConfig
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

	t, tErr := New(name, f, s, confStruct.FileOn, confStruct.DataOnDrive, confStruct.MemOnly)
	if tErr != 0 {
		f.Close()
		return nil, tErr
	}

	t.eMux.Lock()
	t.uMux.Lock()

	// Set optional settings if different from defaults
	if confStruct.EncryptCost != helpers.DefaultEncryptCost {
		t.encryptCost.Store(confStruct.EncryptCost)
	}

	if confStruct.MaxEntries != helpers.DefaultMaxEntries {
		t.maxEntries.Store(confStruct.MaxEntries)
	}

	if confStruct.PartitionMax != helpers.DefaultPartitionMax {
		t.partitionMax.Store(confStruct.PartitionMax)
	}

	if confStruct.MinPass != defaultMinPassword {
		t.minPassword.Store(confStruct.MinPass)
	}

	if confStruct.PassResetLen != defaultPassResetLen {
		t.passResetLen.Store(confStruct.PassResetLen)
	}

	if confStruct.EmailItem != "" {
		t.emailItem.Store(confStruct.EmailItem)
	}

	if confStruct.AltLogin != "" {
		t.altLoginItem.Store(confStruct.AltLogin)
	}

	// Load data/indexing into memory...

	// Open data folder
	df, err := os.Open(namePre)
	if err != nil {
		t.eMux.Unlock()
		t.uMux.Unlock()
		df.Close()
		t.Close(false)
		return nil, helpers.ErrorFileOpen
	}
	files, err := df.Readdir(-1)
	df.Close()
	if err != nil {
		t.eMux.Unlock()
		t.uMux.Unlock()
		t.Close(false)
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
				if eKey, ePass, eData := restoreDataLine(fb[lineByteStart:i]); eData != nil {
					if resErr := t.RestoreUser(eKey, ePass, eData, uint16(fileNum), uint16(lineOn), confStruct.AltLogin); resErr != 0 {
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
	t.uMux.Unlock()
	t.eMux.Unlock()

	return t, 0
}

func restoreDataLine(line []byte) (string, []byte, []interface{}) {
	var jEntry jsonEntry
	mErr := json.Unmarshal(line, &jEntry)
	if mErr != nil {
		return "", nil, nil
	}

	if jEntry.D == nil || jEntry.N == "" || len(jEntry.P) == 0 {
		return "", nil, nil
	}

	return jEntry.N, jEntry.P, jEntry.D
}