package storage

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"os"
	"io"
	"sync"
	"time"
)

const (
	newLineIndicator byte = byte(10)
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

const (
	FileTypeLog     = ".gdbl"
	FileTypeStorage = ".gdbs"
)

var (
	openFilesMux   sync.Mutex
	openFiles map[string]*openFile = make(map[string]*openFile)
	fileOpenTime time.Duration = 10 // in seconds
)

type openFile struct {
	mux sync.Mutex
	file *os.File
	timer *time.Timer
	bytes []byte
	//lineOn uint16
	lineByteOn []int
}

//
func newOpenFile(file string) (*openFile, int) {
	// Open the File
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		openFilesMux.Unlock()
		return nil, helpers.ErrorFileOpen
	}

	// Get file stats
	fs, fsErr := f.Stat()
	if fsErr != nil {
		openFilesMux.Unlock()
		return nil, helpers.ErrorFileOpen
	}

	// Make openFile object
	newOF := openFile{file: f}

	// Get file bytes
	newOF.bytes = make([]byte, fs.Size())
	_, rErr := f.ReadAt(newOF.bytes, 0)
	if rErr != nil && rErr != io.EOF {
		return nil, helpers.ErrorFileRead
	}

	// Make lineByteOn list
	newOF.lineByteOn = []int{0}
	for i, b := range newOF.bytes {
		if b == newLineIndicator {
			newOF.lineByteOn = append(newOF.lineByteOn, i+1)
		}
	}

	// Start close timer
	newOF.timer = time.NewTimer(fileOpenTime * time.Second)
	go fileCloseTimer(newOF.timer, file)

	return &newOF, 0
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
	}
	openFilesMux.Unlock()
	return f, 0
}

func fileCloseTimer(timer *time.Timer, file string) {
	<-timer.C
	openFilesMux.Lock()
	f := openFiles[file]
	openFilesMux.Unlock()
	f.mux.Lock()
	if timer != f.timer {
		// The openFile has already been reset - don't close file
		f.mux.Unlock()
		return
	}
	f.timer = nil
	f.file.Truncate(int64(len(f.bytes)))
	f.bytes = nil
	f.lineByteOn = nil
	f.file.Close()
	f.mux.Unlock()
	openFilesMux.Lock()
	delete(openFiles, file)
	openFilesMux.Unlock()
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
	if f.timer == nil {
		// Timer has already closed and cleared item
		f.mux.Unlock()
		// Reload item
		f, fErr = getOpenFile(file)
		if fErr != 0 {
			return nil, fErr
		}
		f.mux.Lock()
	} else if !f.timer.Reset(fileOpenTime * time.Second) {
		// Timer has ended, but has not cleared item. Remake timer & closer
		f.timer = time.NewTimer(fileOpenTime * time.Second)
		go fileCloseTimer(f.timer, file)
	}


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
	if f.timer == nil {
		// Timer has already closed and cleared item
		f.mux.Unlock()
		// Reload item
		f, fErr = getOpenFile(file)
		if fErr != 0 {
			return fErr
		}
		f.mux.Lock()
	} else if !f.timer.Reset(fileOpenTime * time.Second) {
		// Timer has ended, but has not cleared item. Remake timer & closer
		f.timer = time.NewTimer(fileOpenTime * time.Second)
		go fileCloseTimer(f.timer, file)
	}
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

// Insert appends a JSON encoded []byte at the end of given JSON file
func Insert(file string, json []byte) (uint16, int) {
	f, fErr := getOpenFile(file)
	if fErr != 0 {
		return 0, fErr
	}

	f.mux.Lock()
	if f.timer == nil {
		// Timer has already closed and cleared item
		f.mux.Unlock()
		// Reload item
		f, fErr = getOpenFile(file)
		if fErr != 0 {
			return 0, fErr
		}
		f.mux.Lock()
	} else if !f.timer.Reset(fileOpenTime * time.Second) {
		// Timer has ended, but has not cleared item. Remake timer & closer
		f.timer = time.NewTimer(fileOpenTime * time.Second)
		go fileCloseTimer(f.timer, file)
	}

	//

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
	openFilesMux.Lock()
	fileOpenTime = t
	openFilesMux.Unlock()
}