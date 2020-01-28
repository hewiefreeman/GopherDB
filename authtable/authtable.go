/*
Copyright 2020 Dominique Debergue

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

package authtable

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/storage"
	"github.com/schollz/progressbar"
	"sync"
	"sync/atomic"
	"os"
	"io"
	"encoding/json"
	"net/smtp"
	"strings"
	"strconv"
	"fmt"
)

// File/folder prefixes
const (
	dataFolderPrefix = "Auth-"
)

// Defaults
const (
	defaultMinPassword uint8   = 6
	defaultPassResetLen uint8  = 12
	defaultConfig string       = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"routerOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"routers\":[],\"AuthTables\":[],\"Leaderboards\":[]}"
)

// Tables
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

	// Atomic changeable settings values - 99% read
	partitionMax  atomic.Value // *uint16* maximum entries per data file
	maxEntries    atomic.Value // *uint64* maximum amount of entries in the AuthTable
	minPassword   atomic.Value // *uint8* minimum password length
	encryptCost   atomic.Value // *int* encryption cost of passwords
	passResetLen  atomic.Value // *uint8* the length of passwords created by the database
	emailItem     atomic.Value // *string* item in schema that represents a user's email address
	verifyItem    atomic.Value // *string* when set, the database will send a verified boolean for the User along with insert/update/get queries. The verified boolean is true if the User has successfully verified their account through email. Requires emailItem to be set
	emailSettings atomic.Value // *EmailSettings* Settings for email server authentication, and verification emails
	altLoginItem  atomic.Value // *string* item in schema that a user can log in with as if it's their user name (usually the emailItem)

	// entries
	eMux      sync.Mutex // entries/altLogins map lock
	entries   map[string]*authTableEntry // AuthTable uses a Map for storage since it's only look-up is with user name and password
	altLogins map[string]*authTableEntry // Alternative login item references
	vCodes    map[string]string //

	// unique values
	uMux       sync.Mutex
	uniqueVals map[string]map[interface{}]bool
}

type EmailSettings struct {
	auth smtp.Auth
	AuthType string // "CRAMMD5" or "Plain"
	AuthName string // username
	AuthPass string // password
	AuthID   string // identity (for Plain)
	AuthHost string // host (for Plain)

	VerifyFrom string // Verification email sender
	VerifySubj string // Verification email subject
	VerifyBody string // Verification email body
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
	VerifyItem string
	EmailSettings EmailSettings
	AltLogin string
}

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
func New(name string, configFile *os.File, s schema.Schema, fileOn uint16, dataOnDrive bool, memOnly bool) (*AuthTable, helpers.Error) {
	if len(name) == 0 {
		return nil, helpers.NewError(helpers.ErrorTableNameRequired, name)
	} else if Get(name) != nil {
		return nil, helpers.NewError(helpers.ErrorTableExists, name)
	} else if !s.Validate() {
		return nil, helpers.NewError(helpers.ErrorSchemaInvalid, name)
	}
	// memOnly overrides dataOnDrive
	if memOnly {
		dataOnDrive = false
	}
	namePre := dataFolderPrefix + name
	// Restoring if configFile is not nil
	if configFile == nil {
		var err error
		// Make table storage folder
		err = storage.MakeDir(namePre)
		if err != nil {
			return nil, helpers.NewError(helpers.ErrorCreatingFolder, namePre + helpers.FileTypeConfig + ": " + err.Error())
		}
		// Create/open config file
		configFile, err = os.OpenFile(namePre + helpers.FileTypeConfig, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, helpers.NewError(helpers.ErrorFileOpen, namePre + helpers.FileTypeConfig + ": " + err.Error())
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
			MinPass: defaultMinPassword,
			PassResetLen: defaultPassResetLen,
			EmailItem: "",
			VerifyItem: "",
			EmailSettings: EmailSettings{},
			AltLogin: "",
		}); wErr != 0 {
			return nil, helpers.NewError(wErr, namePre + helpers.FileTypeConfig)
		}
	}
	// Make table
	t := AuthTable{
		name:          name,
		memOnly:       memOnly,
		dataOnDrive:   dataOnDrive,
		schema:        s,
		configFile:    configFile,
		entries:       make(map[string]*authTableEntry),
		altLogins:     make(map[string]*authTableEntry),
		vCodes:        make(map[string]string),
		uniqueVals:    make(map[string]map[interface{}]bool),
		fileOn:        fileOn,
	}
	// Set defaults
	t.partitionMax.Store(helpers.DefaultPartitionMax)
	t.maxEntries.Store(helpers.DefaultMaxEntries)
	t.minPassword.Store(defaultMinPassword)
	t.encryptCost.Store(helpers.DefaultEncryptCost)
	t.passResetLen.Store(defaultPassResetLen)
	t.emailItem.Store("")
	t.verifyItem.Store("")
	t.emailSettings.Store(EmailSettings{})
	t.altLoginItem.Store("")
	// Push to tables map
	tablesMux.Lock()
	tables[name] = &t
	tablesMux.Unlock()
	return &t, helpers.Error{}
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
			VerifyItem: t.verifyItem.Load().(string),
			EmailSettings: t.emailSettings.Load().(EmailSettings),
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

func (t *AuthTable) EncryptCost() int {
	return t.encryptCost.Load().(int)
}

func (t *AuthTable) AltLoginItem() string {
	return t.altLoginItem.Load().(string)
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
	conf := t.makeDefaultConfig(fileOn)
	conf.EncryptCost = cost
	if err := writeConfigFile(t.configFile, conf); err != 0 {
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
	conf := t.makeDefaultConfig(fileOn)
	conf.MaxEntries = max
	if err := writeConfigFile(t.configFile, conf); err != 0 {
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
	conf := t.makeDefaultConfig(fileOn)
	conf.MinPass = min
	if err := writeConfigFile(t.configFile, conf); err != 0 {
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
	conf := t.makeDefaultConfig(fileOn)
	conf.PassResetLen = len
	if err := writeConfigFile(t.configFile, conf); err != 0 {
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
	conf := t.makeDefaultConfig(fileOn)
	conf.AltLogin = item
	if err := writeConfigFile(t.configFile, conf); err != 0 {
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
	conf := t.makeDefaultConfig(fileOn)
	conf.EmailItem = item
	if err := writeConfigFile(t.configFile, conf); err != 0 {
		return err
	}
	t.emailItem.Store(item)
	return 0
}

// SetVerifyItem sets the AuthTable's email verification item. Item must be a string.
func (t *AuthTable) SetVerifyItem(item string) int {
	si := t.schema[item]
	if !si.QuickValidate() {
		return helpers.ErrorInvalidItem
	} else if si.TypeName() != schema.ItemTypeString {
		return helpers.ErrorInvalidItem
	}
	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	conf := t.makeDefaultConfig(fileOn)
	conf.VerifyItem = item
	if err := writeConfigFile(t.configFile, conf); err != 0 {
		return err
	}
	t.verifyItem.Store(item)
	return 0
}

// SetVerifyItem sets the AuthTable's email verification item. Item must be a string.
func (t *AuthTable) SetEmailSettings(settings EmailSettings) int {
	//Build Auth for EmailSettings
	switch settings.AuthType {
	case "Plain":
		settings.auth = smtp.PlainAuth(settings.AuthID, settings.AuthName, settings.AuthPass, settings.AuthHost)
	case "CRAMMD5":
		settings.auth = smtp.CRAMMD5Auth(settings.AuthName, settings.AuthPass)
	default:
		return helpers.ErrorIncorrectAuthType
	}
	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	conf := t.makeDefaultConfig(fileOn)
	conf.EmailSettings = settings
	if err := writeConfigFile(t.configFile, conf); err != 0 {
		return err
	}
	t.emailSettings.Store(settings)
	return 0
}

func (t *AuthTable) SetPartitionMax(max uint16) int {
	if max < helpers.PartitionMin {
		max = helpers.DefaultPartitionMax
	}
	t.eMux.Lock()
	fileOn := t.fileOn
	t.eMux.Unlock()
	conf := t.makeDefaultConfig(fileOn)
	conf.PartitionMax = max
	if err := writeConfigFile(t.configFile, conf); err != 0 {
		return err
	}
	t.partitionMax.Store(max)
	return 0
}

func (t *AuthTable) makeDefaultConfig(fileOn uint16) authtableConfig {
	return authtableConfig{
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
		VerifyItem: t.verifyItem.Load().(string),
		EmailSettings: t.emailSettings.Load().(EmailSettings),
		AltLogin: t.altLoginItem.Load().(string),
	}
}

func writeConfigFile(f *os.File, c authtableConfig) int {
	jBytes, jErr := helpers.Fjson.MarshalIndent(c, "", "   ")
	if jErr != nil {
		return helpers.ErrorJsonEncoding
	}

	// Write to config file
	if _, wErr := f.WriteAt(jBytes, 0); wErr != nil {
		return helpers.ErrorFileUpdate
	}
	f.Truncate(int64(len(jBytes)))
	//
	return 0
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   AuthTable Restoring   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Restore restores an AuthTable by name; requires a valid config file and data folder
func Restore(name string) (*AuthTable, helpers.Error) {
	fmt.Printf("Restoring Auth '%v'...\n", name)
	namePre := dataFolderPrefix + name
	// Open the File
	f, err := os.OpenFile(namePre + helpers.FileTypeConfig, os.O_RDWR, 0755)
	if err != nil {
		return nil, helpers.NewError(helpers.ErrorFileOpen, "Config file missing for Auth '" + name + "'")
	}
	// Get file stats
	fs, fsErr := f.Stat()
	if fsErr != nil {
		f.Close()
		return nil, helpers.NewError(helpers.ErrorFileOpen, "Error reading config file for Auth '"+name+"'")
	}
	// Get file bytes
	bytes := make([]byte, fs.Size())
	_, rErr := f.ReadAt(bytes, 0)
	if rErr != nil && rErr != io.EOF {
		f.Close()
		return nil, helpers.NewError(helpers.ErrorFileRead, "Config data is corrupt for Auth '" + name + "'")
	}
	// Make confStruct from json bytes
	var confStruct authtableConfig
	mErr := json.Unmarshal(bytes, &confStruct)
	if mErr != nil {
		f.Close()
		return nil, helpers.NewError(helpers.ErrorJsonDecoding,
			"Config contains JSON syntax errors for Auth '" + name + "': " + mErr.Error())
	}
	// Make schema with the schemaList
	s, schemaErr := schema.Restore(confStruct.Schema)
	if schemaErr.ID != 0 {
		f.Close()
		schemaErr.From = "(Auth '" + name + "') " + schemaErr.From
		return nil, schemaErr
	}
	at, tErr := New(name, f, s, confStruct.FileOn, confStruct.DataOnDrive, confStruct.MemOnly)
	if tErr.ID != 0 {
		f.Close()
		return nil, tErr
	}
	at.eMux.Lock()
	at.uMux.Lock()
	// Set optional settings if different from defaults
	if confStruct.EncryptCost != helpers.DefaultEncryptCost {
		at.encryptCost.Store(confStruct.EncryptCost)
	}
	if confStruct.MaxEntries != helpers.DefaultMaxEntries {
		at.maxEntries.Store(confStruct.MaxEntries)
	}
	if confStruct.PartitionMax != helpers.DefaultPartitionMax {
		at.partitionMax.Store(confStruct.PartitionMax)
	}
	if confStruct.MinPass != defaultMinPassword {
		at.minPassword.Store(confStruct.MinPass)
	}
	if confStruct.PassResetLen != defaultPassResetLen {
		at.passResetLen.Store(confStruct.PassResetLen)
	}
	if confStruct.EmailItem != "" {
		at.emailItem.Store(confStruct.EmailItem)
	}
	if confStruct.EmailSettings.AuthType != "" {
		at.emailSettings.Store(confStruct.EmailSettings)
	}
	if confStruct.AltLogin != "" {
		at.altLoginItem.Store(confStruct.AltLogin)
	}
	// Open data folder
	df, err := os.Open(namePre)
	if err != nil {
		at.eMux.Unlock()
		at.uMux.Unlock()
		df.Close()
		at.Close(false)
		return nil, helpers.NewError(helpers.ErrorFileOpen, "Missing data folder for Auth '" + name + "'")
	}
	files, err := df.Readdir(-1)
	df.Close()
	if err != nil {
		at.eMux.Unlock()
		at.uMux.Unlock()
		at.Close(false)
		return nil, helpers.NewError(helpers.ErrorFileRead, "Error reading files in data folder for Auth '" + name + "'")
	}
	fmt.Printf("Loading Auth data for '%v'...\n", name)
	// Make progress bar
	pBar := progressbar.New(len(files))
	// Go through files
	for _, fileStats := range files {
		// Get file number
		fileNameSplit := strings.Split(fileStats.Name(), ".")
		fileNum, fnErr := strconv.Atoi(fileNameSplit[0])
		if fnErr != nil || len(fileNameSplit) < 2 || "."+fileNameSplit[1] != helpers.FileTypeStorage {
			// Not a storage file...
			pBar.Add(1)
			continue
		}
		var of *storage.OpenFile
		var err int
		if of, err = storage.GetOpenFile(namePre + "/" + fileStats.Name()); err != 0 {
			fmt.Printf("Error: Auth '%v':: Data file '%v' is corrupt!\n", name, fileStats.Name())
			pBar.Add(1)
			continue
		}
		for i := 0; i < of.Lines(); i++ {
			// Get line bytes
			var lb []byte
			if lb, err = of.Read(uint16(i+1)); err != 0 {
				fmt.Printf("Error: Auth '%v':: Could not read line %v of '%v'!\n", name, i + 1, fileStats.Name())
				continue
			}
			eKey, ePass, eData := restoreDataLine(lb)
			if eData == nil {
				fmt.Printf("Error: Auth '%v':: Incorrect JSON format on line %v of '%v'!\n", name, i + 1, fileStats.Name())
				continue
			}
			if err = at.restoreUser(eKey, []byte(ePass), eData, uint16(fileNum), uint16(i+1)); err != 0 {
				fmt.Printf("Error: Auth '%v':: Line %v of '%v' error code %v\n", name, i + 1, fileStats.Name(), err)
				continue
			}
		}
		pBar.Add(1)
	}
	at.uMux.Unlock()
	at.eMux.Unlock()
	//
	return at, helpers.Error{}
}

func restoreDataLine(line []byte) (string, string, []interface{}) {
	var jEntry jsonEntry
	mErr := json.Unmarshal(line, &jEntry)
	if mErr != nil {
		return "", "", nil
	}
	if jEntry.D == nil || jEntry.N == "" || len(jEntry.P) == 0 {
		return "", "", nil
	}
	return jEntry.N, jEntry.P, jEntry.D
}