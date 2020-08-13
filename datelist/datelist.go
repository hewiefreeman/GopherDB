package datelist

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
	"time"
)

// Month names
const (
	January = "January"
	February = "February"
	March = "March"
	April = "April"
	May = "May"
	June = "June"
	July = "July"
	August = "August"
	September = "September"
	October = "October"
	November = "November"
	December = "December"

	// Alias names (May doesn't need one ?)
	JanuaryAlias = "Jan"
	FebruaryAlias = "Feb"
	MarchAlias = "Mar"
	AprilAlias = "Apr"
	JuneAlias = "Jun"
	JulyAlias = "Jul"
	AugustAlias = "Aug"
	SeptemberAlias = "Sep"
	OctoberAlias = "Oct"
	NovemberAlias = "Nov"
	DecemberAlias = "Dec"
)

// Month IDs
const (
	JanuaryID = iota
	FebruaryID
	MarchID
	AprilID
	MayID
	JuneID
	JulyID
	AugustID
	SeptemberID
	OctoberID
	NovemberID
	DecemberID
)

const (
	monthsInYear = 12
	daysInMonth  = 32
	hoursInDay   = 24

	dataFolderPrefix = "DL-"
)

var (
	listsMux     sync.Mutex
	lists        map[string]*DateList = make(map[string]*DateList)
	timeLocation *time.Location = time.UTC
)

type DateList struct {
	fileOn    uint16 // locked by eMux - placed for memory efficiency

	// Settings and schema - read only
	memOnly       bool // Store data in memory only (overrides dataOnDrive)
	dataOnDrive   bool // when true, entry data is not stored in memory, only indexing
	persistName   string // table's logger/persist folder name
	schema        *schema.Schema // table's schema

	// Atomic changeable settings values - 99% read
	partitionMax  atomic.Value // *uint16* maximum entries per data file
	maxEntries    atomic.Value // *uint64* maximum amount of entries in the Datelist
	encryptCost   atomic.Value // *int* encryption cost of encrypted items
	updateTime    atomic.Value // *bool* when true, inserts as well as updates will set an entry's time stamp and database position
	unixNano      atomic.Value // *bool* when true, time stamps will be stored in Unix Nano instead of the default Unix

	// date list & entry counter
	eMux       sync.Mutex // entries lock
	datelist   []Year
	entryCount uint64

	// unique values
	uMux       sync.Mutex
	uniqueVals map[string]map[interface{}]bool
}

type Year struct {

	months []Month
}

type Month struct {
	entryCount uin64
	days []Day
}

type Day struct {
	entryCount uint64
	entries    []*DateListEntry
}

type DateListEntry struct {
	persistFile  uint16
	persistIndex uint16

	mux   sync.Mutex
	iTime time.Time
	data  []interface{}
}

// New creates a new Datelist
func New(name string, s *schema.Schema, fileOn uint16, dataOnDrive bool, memOnly bool) (*DateList, int) {
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

	// Make table folder   & update config file !!!
	mkErr := storage.MakeDir(namePre)
	if mkErr != nil {
		return nil, helpers.ErrorCreatingFolder
	}

	// make table
	d := Datelist{
		persistName:   namePre,
		memOnly:       memOnly,
		dataOnDrive:   dataOnDrive,
		schema:        s,
		datelist:      make([]Year),
		entryCount:    0,
		uniqueVals:    make(map[string]map[interface{}]bool),
		fileOn:        fileOn,
	}

	// set defaults
	d.partitionMax.Store(helpers.DefaultPartitionMax)
	d.maxEntries.Store(helpers.DefaultMaxEntries)
	d.encryptCost.Store(helpers.DefaultEncryptCost)
	d.updateTime.Store(false)
	d.unixNano.Store(false)

	// push to stores map
	listsMux.Lock()
	lists[name] = &d
	listsMux.Unlock()

	return &d, 0
}

// Delete deletes a Datelist with the given name.
func Delete(name string) int {
	if len(name) == 0 {
		return helpers.ErrorTableNameRequired
	} else if Get(name) == nil {
		return helpers.ErrorTableDoesntExist
	}

	listsMux.Lock()
	delete(lists, name)
	listsMux.Unlock()

	// !!!!!! delete data folder from system & delete log file & update config file

	return 0
}

// Get retrieves a Datelist by name
func Get(name string) *DateList {
	listsMux.Lock()
	d := lists[name]
	listsMux.Unlock()

	return d
}