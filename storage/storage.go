/*
storage package Copyright 2020 Dominique Debergue

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

// storage
package storage

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"encoding/json"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Defaults and indicators
const (
	newLineIndicator     byte   = byte(10)
	openBracketIndicator byte   = byte(91)
	uint32Max            uint32 = 2147483647

	defaultFileOpenTime  time.Duration = 20 * time.Second
	defaultMaxOpenFiles  uint16        = 25
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
	openFiles      map[string]*OpenFile
	inited           bool

	// Settings
	fileOpenTime atomic.Value // time.Duration *int64* in seconds
	maxOpenFiles atomic.Value // uint16

	// default vars
	defaultIndexingBytes []byte = []byte("[]")
)

// OpenFile represents a file on a disk that is open and ready for I/O
type OpenFile struct {
	name        string
	mux         sync.Mutex
	file        *os.File
	bytes       []byte
	lineByteOn  []int64
	indexStart  int64
	accessed    uint32
	expireTimer *time.Timer
	cancelChan  chan bool
}

// Init initializes the storage package. Must be called before using.
func Init() {
	openFilesMux.Lock()
	if inited {
		openFilesMux.Unlock()
		return
	}
	fileOpenTime.Store(defaultFileOpenTime)
	maxOpenFiles.Store(defaultMaxOpenFiles)
	openFiles = make(map[string]*OpenFile)
	inited = true
	openFilesMux.Unlock()
}

// ShutDown closes all OpenFiles and shuts the storage engine down.
func ShutDown() {
	openFilesMux.Lock()
	if !inited {
		openFilesMux.Unlock()
		return
	}
	// Close all OpenFiles
	for fName, f := range openFiles {
		f.mux.Lock()
		f.cancelChan <- true
		close(f.cancelChan)
		f.file.Truncate(int64(len(f.bytes)))
		f.bytes = nil
		f.lineByteOn = nil
		f.file.Close()
		f.mux.Unlock()
		delete(openFiles, fName)
	}
	inited = false
	openFilesMux.Unlock()
}

//
func newOpenFile(file string) (*OpenFile, int) {
	// Close the least accessed OpenFile
	if len(openFiles) >= int(maxOpenFiles.Load().(uint16)) {
		// Get least accessed OpenFile
		var laf *OpenFile
		for _, f := range openFiles {
			if laf == nil {
				laf = f
			} else if laf.accessed > f.accessed {
				laf = f
			}
		}
		// Close least accessed OpenFile and cancel it's close timer
		laf.mux.Lock()
		laf.cancelChan <- true
		close(laf.cancelChan)
		laf.file.Truncate(int64(len(laf.bytes)))
		laf.bytes = nil
		laf.lineByteOn = nil
		laf.file.Close()
		laf.mux.Unlock()
		delete(openFiles, laf.name)
	}
	// Open the File
	var f *os.File
	var err error
	if f, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0755); err != nil {
		return nil, helpers.ErrorFileOpen
	}
	// Get file stats
	var fs os.FileInfo
	if fs, err = f.Stat(); err != nil {
		return nil, helpers.ErrorFileOpen
	}
	// Make new OpenFile object
	newOF := OpenFile{file: f, name: file, accessed: 1}
	// Get file bytes
	newOF.bytes = make([]byte, fs.Size())
	_, rErr := f.ReadAt(newOF.bytes, 0)
	if rErr != nil && rErr != io.EOF {
		return nil, helpers.ErrorFileRead
	}
	// Get indexing
	if (len(newOF.bytes) == 0) {
		// New file, create indexing layer
		newOF.bytes = append(newOF.bytes, defaultIndexingBytes...)
		newOF.indexStart = 0
		newOF.lineByteOn = []int64{}
		if _, wErr := f.WriteAt(newOF.bytes, int64(0)); wErr != nil {
			return nil, helpers.ErrorFileOpen
		}
	} else {
		// Get indexing bytes
		var iPosBytes []byte
		for i := len(newOF.bytes) - 1; i >= 0; i-- {
			if newOF.bytes[i] == openBracketIndicator {
				iPosBytes = newOF.bytes[i:]
				newOF.indexStart = int64(i)
				break
			} else if i == 0 {
				return nil, helpers.ErrorJsonIndexingFormat
			}
		}
		// Get indexing list
		if err := json.Unmarshal(iPosBytes, &newOF.lineByteOn); err != nil {
			return nil, helpers.ErrorJsonIndexingFormat
		}
	}
	ofp := &newOF
	// Start close timer
	newOF.expireTimer = time.NewTimer(fileOpenTime.Load().(time.Duration))
	newOF.cancelChan = make(chan bool, 1)
	go fileCloseTimer(ofp)

	return ofp, 0
}

// GetOpenFile
func GetOpenFile(file string) (*OpenFile, int) {
	var f *OpenFile
	openFilesMux.Lock()
	if !inited {
		openFilesMux.Unlock()
		return nil, helpers.ErrorStorageNotInitialized
	}
	f = openFiles[file]
	if f == nil {
		var fileErr int
		if f, fileErr = newOpenFile(file); fileErr != 0 {
			openFilesMux.Unlock()
			return nil, fileErr
		}
		openFiles[file] = f
	} else {
		// If the closeTimer cannot be reset, the timer has already expired,
		// but has not been removed. Make a new Timer and replace the closeTimer's
		// Timer with the new Timer so that it cancels the fileCloseTimer() action.
		if !f.expireTimer.Reset(fileOpenTime.Load().(time.Duration)) {
			close(f.cancelChan)
			f.expireTimer = time.NewTimer(fileOpenTime.Load().(time.Duration))
			f.cancelChan = make(chan bool, 1)
			go fileCloseTimer(f)
		}
		// Increase accessed uint unless at max value
		if f.accessed < uint32Max {
			f.accessed++
		}
	}
	openFilesMux.Unlock()
	return f, 0
}

// Waits for a signal on either t.t.C for a timer expire, or t.c for a timer cancel.
func fileCloseTimer(f *OpenFile) {
	et := f.expireTimer
	select {
	case <- et.C:
		// Timer ended
		openFilesMux.Lock()
		if et != f.expireTimer {
			// The OpenFile has already been reset by GetOpenFile() - cancel action
			openFilesMux.Unlock()
			return
		}
		delete(openFiles, f.name)
		openFilesMux.Unlock()
		f.mux.Lock()
		close(f.cancelChan)
		f.file.Truncate(int64(len(f.bytes)))
		f.bytes = nil
		f.lineByteOn = nil
		f.file.Close()
		f.mux.Unlock()
	case <- f.cancelChan:
		// Timer cancelled
		et.Stop()
	}
}

// MakeDir creates a directory path on the system
func MakeDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}

// DeleteDir deletes a directory from the system
func DeleteDir(dir string) error {
	return os.RemoveAll(dir)
}

// Read opens a file by name, then returns the data from said line.
func Read(file string, line uint16) ([]byte, int) {
	f, fErr := GetOpenFile(file)
	if fErr != 0 {
		return nil, fErr
	}
	return f.Read(line)
}

// Read returns the data in an OpenFile from said line.
func (f *OpenFile) Read(line uint16) ([]byte, int) {
	f.mux.Lock()
	// Get the start and end index of line
	bStart := f.lineByteOn[line-1]
	var bEnd int64
	if len(f.lineByteOn) == int(line) {
		// Last item in file
		bEnd = f.indexStart - 1
	} else {
		// Not last item...
		bEnd = f.lineByteOn[line] - 1
	}
	if bEnd < bStart {
		return nil, helpers.ErrorInternalFormatting
	}
	bytes := f.bytes[bStart:bEnd]
	f.mux.Unlock()
	return bytes, 0
}

// Update updates JSON encoded []byte line at given index of given file
func Update(file string, line uint16, jData []byte) int {
	f, fErr := GetOpenFile(file)
	if fErr != 0 {
		return fErr
	}

	f.mux.Lock()

	// Get the start and end index of line
	iStart := f.lineByteOn[line-1]
	var iEnd int64
	if len(f.lineByteOn) == int(line) {
		// Last item in file
		iEnd = f.indexStart - 1
	} else {
		// Not last item...
		iEnd = f.lineByteOn[line] - 1
	}
	// Calculate byte difference for subsequent lines
	iDif := int64(len(jData)) - (iEnd - iStart)
	for i := int(line); i < len(f.lineByteOn); i++ {
		f.lineByteOn[i] += iDif
	}
	indexStart := f.indexStart
	f.indexStart += iDif
	// Make indexing data
	lineByteOnData, err := helpers.Fjson.Marshal(f.lineByteOn)
	if err != nil {
		return helpers.ErrorInternalFormatting
	}
	// Make & push data
	rHalf := append(jData, f.bytes[iEnd:indexStart]...)
	rHalf = append(rHalf, lineByteOnData...)
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
func Insert(file string, jData []byte) (uint16, int) {
	f, fErr := GetOpenFile(file)
	if fErr != 0 {
		return 0, fErr
	}
	f.mux.Lock()
	// Insert and get lineOn
	lineOn := uint16(len(f.lineByteOn) + 1)
	f.lineByteOn = append(f.lineByteOn, f.indexStart)
	// Make indexing data
	lineByteOnData, err := helpers.Fjson.Marshal(f.lineByteOn)
	if err != nil {
		return 0, helpers.ErrorInternalFormatting
	}
	// Append a new line to jData and get new indexStart
	jData = append(jData, newLineIndicator)
	iStart := f.indexStart
	f.indexStart += int64(len(jData))
	// Append lineByteOnData and write jData to disk
	jData = append(jData, lineByteOnData...)
	if _, wErr := f.file.WriteAt(jData, iStart); wErr != nil {
		f.mux.Unlock()
		return 0, helpers.ErrorFileAppend
	}
	f.bytes = append(f.bytes[:iStart], jData...)
	f.mux.Unlock()
	return lineOn, 0
}

// SetFileOpenTime preference allows you to keep OpenFiles open for a given duration.
func SetFileOpenTime(t time.Duration) {
	if t <= 0 {
		return
	}
	fileOpenTime.Store(t)
}

// SetMaxOpenFiles preference sets the maximum number of OpenFile to be open.
// When the maximum number of OpenFile has been reached, a random OpenFile will be closed.
func SetMaxOpenFiles(max uint16) {
	if max == 0 {
		return
	}
	maxOpenFiles.Store(max)
}

// GetNumopenFiles gets the number of OpenFiles in system
func GetNumOpenFiles() int {
	var i int
	openFilesMux.Lock()
	i = len(openFiles)
	openFilesMux.Unlock()
	return i
}

// Lines returns the number of data lines for this OpenFile
func (f *OpenFile) Lines() int {
	f.mux.Lock()
	l := len(f.lineByteOn)
	f.mux.Unlock()
	return l
}