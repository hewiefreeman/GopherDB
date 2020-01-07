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
	OpenFilesMux   sync.Mutex
	OpenFiles      map[string]*OpenFile   = make(map[string]*OpenFile)
	OpenFileTimers map[string]*time.Timer = make(map[string]*time.Timer)

	// Settings
	fileOpenTime atomic.Value // time.Duration *int64* in seconds
	maxOpenFiles atomic.Value // uint16

	// default vars
	defaultIndexingBytes []byte = []byte("[]")
)

type OpenFile struct {
	name       string
	mux        sync.Mutex
	file       *os.File
	bytes      []byte
	lineByteOn []int64
	indexStart int64
}

// Init initializes the storage package. Must be called before using.
func Init() {
	fileOpenTime.Store(defaultFileOpenTime)
	maxOpenFiles.Store(defaultMaxOpenFiles)
}

//
func newOpenFile(file string) (*OpenFile, int) {
	// Close the first found (random) OpenFile if max files are open
	if len(OpenFiles) >= int(maxOpenFiles.Load().(uint16)) {
		for fileName, f := range OpenFiles {
			OpenFileTimers[fileName].Stop()
			delete(OpenFiles, fileName)
			delete(OpenFileTimers, fileName)
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
	OpenFileTimers[file] = time.NewTimer(fileOpenTime.Load().(time.Duration))
	go fileCloseTimer(OpenFileTimers[file], newOFPointer)

	return newOFPointer, 0
}

//
func GetOpenFile(file string) (*OpenFile, int) {
	var f *OpenFile
	OpenFilesMux.Lock()
	f = OpenFiles[file]
	if f == nil {
		var fileErr int
		if f, fileErr = newOpenFile(file); fileErr != 0 {
			OpenFilesMux.Unlock()
			return nil, fileErr
		}
		OpenFiles[file] = f
	} else {
		if !OpenFileTimers[file].Reset(fileOpenTime.Load().(time.Duration)) {
			t := time.NewTimer(fileOpenTime.Load().(time.Duration))
			OpenFileTimers[file] = t
			go fileCloseTimer(t, f)
		}
	}
	OpenFilesMux.Unlock()
	return f, 0
}

// TO-DO: Make timer.C listen for two signals, and make the emergency file close send a "stopped" signal to indicate the timer has stopped
func fileCloseTimer(timer *time.Timer, f *OpenFile) {
	<-timer.C
	OpenFilesMux.Lock()
	f.mux.Lock()
	if timer != OpenFileTimers[f.name] {
		// The OpenFile has already been reset - don't close file
		f.mux.Unlock()
		OpenFilesMux.Unlock()
		return
	}
	delete(OpenFiles, f.name)
	delete(OpenFileTimers, f.name)
	OpenFilesMux.Unlock()
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
func Read(file string, line uint16) ([]byte, int) {
	f, fErr := GetOpenFile(file)
	if fErr != 0 {
		return nil, fErr
	}
	return f.Read(line)
}

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
	OpenFilesMux.Lock()
	i = len(OpenFiles)
	OpenFilesMux.Unlock()
	return i
}

func (f *OpenFile) Lines() int {
	f.mux.Lock()
	l := len(f.lineByteOn)
	f.mux.Unlock()
	return l
}