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
	newLineIndicator     byte = byte(10)
	openBracketIndicator byte = byte(91)

	defaultFileOpenTime  time.Duration = 20 * time.Second
	defaultMaxOpenFiles  uint16        = 50
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
	openFiles      map[string]*OpenFile   = make(map[string]*OpenFile)
	openFileTimers map[string]*closeTimer = make(map[string]*closeTimer)

	// Settings
	fileOpenTime atomic.Value // time.Duration *int64* in seconds
	maxOpenFiles atomic.Value // uint16

	// default vars
	defaultIndexingBytes []byte = []byte("[]")
)

// OpenFile represents a file on a disk that is open and ready for I/O
type OpenFile struct {
	name       string
	mux        sync.Mutex
	file       *os.File
	bytes      []byte
	lineByteOn []int64
	indexStart int64
}

type closeTimer struct {
	t *time.Timer // Close timer
	c chan bool  // Cancel chan
}

// Init initializes the storage package. Must be called before using.
func Init() {
	fileOpenTime.Store(defaultFileOpenTime)
	maxOpenFiles.Store(defaultMaxOpenFiles)
}

//
func newOpenFile(file string) (*OpenFile, int) {
	// Close the first found (random) OpenFile if maximum files are open
	if len(openFiles) >= int(maxOpenFiles.Load().(uint16)) {
		for fileName, f := range openFiles {
			ct := openFileTimers[fileName]
			ct.c <- true
			close(ct.c)
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
	newOF := OpenFile{file: f, name: file}

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

	newOFPointer := &newOF

	// Start close timer
	newCT := closeTimer{
		t: time.NewTimer(fileOpenTime.Load().(time.Duration)),
		c: make(chan bool),
	}
	openFileTimers[file] = &newCT
	go fileCloseTimer(openFileTimers[file], newOFPointer)

	return newOFPointer, 0
}

// GetOpenFile
func GetOpenFile(file string) (*OpenFile, int) {
	var f *OpenFile
	openFilesMux.Lock()
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
		// Timer with the new Timer so that it cancells the fileCloseTimer() action.
		if !openFileTimers[file].t.Reset(fileOpenTime.Load().(time.Duration)) {
			t := time.NewTimer(fileOpenTime.Load().(time.Duration))
			openFileTimers[file].t = t
			go fileCloseTimer(t, f)
		}
	}
	openFilesMux.Unlock()
	return f, 0
}

// Waits for a signal on either t.t.C for a timer expire, or t.c for a timer cancel.
func fileCloseTimer(t *closeTimer, f *OpenFile) {
	select {
	case <- t.t.C:
		// Timer ended
		openFilesMux.Lock()
		if timer != openFileTimers[f.name].t {
			// The OpenFile has already been reset by GetOpenFile() - cancel action
			openFilesMux.Unlock()
			return
		}
		delete(openFiles, f.name)
		delete(openFileTimers, f.name)
		openFilesMux.Unlock()
		f.mux.Lock()
		f.file.Truncate(int64(len(f.bytes)))
		f.bytes = nil
		f.lineByteOn = nil
		f.file.Close()
		f.mux.Unlock()
	case <- t.c:
		// Timer cancelled
		t.t.Stop()
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
func GetNumopenFiles() int {
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