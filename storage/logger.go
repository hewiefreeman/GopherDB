package storage

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"sync"
)

// NOTES: An idea for storing recent inserts/updates for a certain amount of time, then storing info to actual location.
//        Theoretically should make storage faster, since this will be able to update multiple entries in a file at once,
//        where without, the same file could be opened/closed a bunch of times in a row for different queries.

const (
	loggerFileExtension = ".log"
	persistFolderSuffix = "_data"
)

var (
	lMux    sync.Mutex
	loggers map[string]*Logger = map[string]*Logger{}
)

type Logger struct {
	dataOnDrive  bool
	partitionMax uint16 // maximum persist file entries
	name         string

	// persistance
	pMux   sync.Mutex // fileOn/lineOn lock
	fileOn uint16
	lineOn uint16
}

func NewLogger(name string, fileOn uint16, lineOn uint16, partitionMax uint16, dataOnDrive bool) (*Logger, int) {
	l := Logger{
		dataOnDrive:  dataOnDrive,
		partitionMax: partitionMax,
		name:         name,
		fileOn:       fileOn,
		lineOn:       lineOn,
	}

	// Make table and persist folder
	//respChan := make(chan interface{})

	/*if err := MakeDir(name + "/" + name + persistFolderSuffix); err != nil {
		return nil, helpers.ErrorTableFolderCreate
	}*/

	// Make logger file
	/*if err := MakeFile(name + "/" + name + loggerFileExtension); err != nil {
		return nil, helpers.ErrorLoggerFileCreate
	}*/

	lMux.Lock()
	if loggers[name] != nil {
		lMux.Unlock()
		return nil, helpers.ErrorLoggerExists
	}
	loggers[name] = &l
	lMux.Unlock()
	return &l, 0
}

// Insert is used for table inserts
func (l *Logger) Insert(dataID interface{}, data interface{}) (uint16, uint16) {
	l.pMux.Lock()
	l.lineOn++
	if l.lineOn > l.partitionMax {
		l.lineOn = 1
		l.fileOn++
	}
	lineOn := l.lineOn
	fileOn := l.fileOn

	// Use QueueFileAction() to append data to log file

	l.pMux.Unlock()
	return fileOn, lineOn
}

// Update is used for table updates
func (l *Logger) Update(dataID interface{}, data interface{}, fileIn uint16, lineOn uint16) {
	l.pMux.Lock()
	// Use QueueFileAction() to append data to log file
	l.pMux.Unlock()
}