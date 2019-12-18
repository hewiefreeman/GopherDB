package storage

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"os"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	newLineIndicator    byte          = byte(10)
	defaultFileOpenTime time.Duration = 20 * time.Second
	defaultMaxOpenFiles uint16        = 50
)

// File actions
const (
	FileActionClose = iota
	FileActionRead
	FileActionInsert
	FileActionInsertMulti
	FileActionUpdate
	FileActionUpdateMulti
)

var (
	// Open Files
	openFilesMux   sync.Mutex
	openFiles      map[string]*openFile   = make(map[string]*openFile)
	openFileTimers map[string]*time.Timer = make(map[string]*time.Timer)

	// Settings
	fileOpenTime atomic.Value // time.Duration *int64* in seconds
	maxOpenFiles atomic.Value // uint16
)

type openFile struct {
	name       string
	mux        sync.Mutex
	file       *os.File
	bytes      []byte
	lineByteOn []int
}

// Init initializes the storage package. Must be called before using.
func Init() {
	fileOpenTime.Store(defaultFileOpenTime)
	maxOpenFiles.Store(defaultMaxOpenFiles)
}

//
func newOpenFile(file string) (*openFile, int) {
	// Close the first found (random) openFile if max files are open
	if len(openFiles) >= int(maxOpenFiles.Load().(uint16)) {
		for fileName, f := range openFiles {
			openFileTimers[fileName].Stop()
			delete(openFiles, fileName)
			delete(openFileTimers, fileName)
			f.mux.Lock()
			f.file.Truncate(int64(len(f.bytes)))
			f.bytes = nil
			f.lineByteOn = nil
			f.file.Close()
			f.mux.Unlock()
			break
		}
	}

	// Open the File
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, helpers.ErrorFileOpen
	}

	// Get file stats
	fs, fsErr := f.Stat()
	if fsErr != nil {
		return nil, helpers.ErrorFileOpen
	}

	// Make openFile object
	newOF := openFile{file: f, name: file}

	// Get file bytes
	newOF.bytes = make([]byte, fs.Size())
	_, rErr := f.ReadAt(newOF.bytes, 0)
	if rErr != nil && rErr != io.EOF {
		return nil, helpers.ErrorFileRead
	}

	// Make lineByteOn list
	newOF.lineByteOn = []int{0}
	for i, b := range newOF.bytes {
		if b == newLineIndicator && len(newOF.bytes[i+1:]) > 0 {
			newOF.lineByteOn = append(newOF.lineByteOn, i+1)
		}
	}

	newOFPointer := &newOF

	// Start close timer
	openFileTimers[file] = time.NewTimer(fileOpenTime.Load().(time.Duration))
	go fileCloseTimer(openFileTimers[file], newOFPointer)

	return newOFPointer, 0
}

//
func getOpenFile(file string) (*openFile, int) {
	var f *openFile
	openFilesMux.Lock()
	f = openFiles[file]
	if f == nil {
		var fileErr int
		f, fileErr = newOpenFile(file)
		if fileErr != 0 {
			openFilesMux.Unlock()
			return nil, fileErr
		}
		openFiles[file] = f
	} else {
		if !openFileTimers[file].Reset(fileOpenTime.Load().(time.Duration)) {
			t := time.NewTimer(fileOpenTime.Load().(time.Duration))
			openFileTimers[file] = t;
			go fileCloseTimer(t, f)
		}
	}
	openFilesMux.Unlock()
	return f, 0
}

func fileCloseTimer(timer *time.Timer, f *openFile) {
	<-timer.C
	openFilesMux.Lock()
	f.mux.Lock()
	if timer != openFileTimers[f.name] {
		// The openFile has already been reset - don't close file
		f.mux.Unlock()
		openFilesMux.Unlock()
		return
	}
	delete(openFiles, f.name)
	delete(openFileTimers, f.name)
	openFilesMux.Unlock()
	f.file.Truncate(int64(len(f.bytes)))
	f.bytes = nil
	f.lineByteOn = nil
	f.file.Close()
	f.mux.Unlock()
}

// MakeDir creates a directory path on the system
func MakeDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}

// DeleteDir deletes a directory from the system
func DeleteDir(dir string) error {
	return os.RemoveAll(dir)
}

// Read reads a specific line from a file.
func Read(file string, index uint16) ([]byte, int) {
	f, fErr := getOpenFile(file)
	if fErr != 0 {
		return nil, fErr
	}

	f.mux.Lock()

	// Get the start and end index of line
	bStart := f.lineByteOn[index-1]
	bEnd := 0
	for i := bStart; i < len(f.bytes); i++ {
		if f.bytes[i] == newLineIndicator {
			bEnd = i
			break
		}
	}
	if bEnd < bStart {
		bEnd = len(f.bytes)
	}

	bytes := f.bytes[bStart:bEnd]
	f.mux.Unlock()
	return bytes, 0
}

// Update updates JSON encoded []byte line at given index of given file
func Update(file string, index uint16, json []byte) int {
	f, fErr := getOpenFile(file)
	if fErr != 0 {
		return fErr
	}

	f.mux.Lock()

	// Get the start and end index of line
	iStart := f.lineByteOn[index-1]
	var iEnd int
	for i := iStart; i < len(f.bytes); i++ {
		if f.bytes[i] == newLineIndicator {
			iEnd = i
			break
		}
	}
	// Calculate byte difference for subsequent lines
	iDif := len(json)-(iEnd-iStart)
	for i := int(index); i < len(f.lineByteOn); i++ {
		f.lineByteOn[i] += iDif
	}
	// Make & push data
	rHalf := append(json, f.bytes[iEnd:]...)
	f.bytes = append(f.bytes[:iStart], rHalf...)
	if _, wErr := f.file.WriteAt(rHalf, int64(iStart)); wErr != nil {
		f.mux.Unlock()
		return helpers.ErrorFileUpdate
	}
	f.mux.Unlock()
	return 0
}

// Insert appends a JSON encoded []byte at the end of given JSON file and reports back the
// line number that was written to
func Insert(file string, json []byte) (uint16, int) {
	f, fErr := getOpenFile(file)
	if fErr != 0 {
		return 0, fErr
	}

	f.mux.Lock()

	// Insert and get line on
	lineOn := uint16(len(f.lineByteOn)+1)
	json = append(json, newLineIndicator)
	if _, wErr := f.file.WriteAt(json, int64(len(f.bytes))); wErr != nil {
		f.mux.Unlock()
		return 0, helpers.ErrorFileAppend
	}
	f.lineByteOn = append(f.lineByteOn, len(f.bytes))
	f.bytes = append(f.bytes, json...)
	f.mux.Unlock()
	return lineOn, 0
}

func SetFileOpenTime(t time.Duration) {
	if t <= 0 {
		return
	}
	fileOpenTime.Store(t)
}

func SetMaxOpenFiles(max uint16) {
	if max == 0 {
		return
	}
	maxOpenFiles.Store(max)
}

// Get the number of open files in system
func GetNumOpenFiles() int {
	var i int
	openFilesMux.Lock()
	i = len(openFiles)
	openFilesMux.Unlock()
	return i
}