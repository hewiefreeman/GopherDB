package helpers

import (
	"time"
	"os"
	"sync"
	"sync/atomic"
	"strings"
	"strconv"
)

const (
	logsFolder string = "/logs"
)

var (
	// Logger
	logFile *os.File
	logMux  sync.Mutex
	byteOn  int64
	entryOn int64
	logNum  int64
	logInit bool

	// Settings
	maxLogFileSize int = 500 // Maximum log messages in a single log file
	logPrio        int = 5   // Minimum priority level to display. Priority level can be from 1 to 5
)

func InitLogger(prio int, maxSize int) error {
	if logInit {
		return nil
	}
	if prio < 1 {
		prio = 1
	} else if prio > 5 {
		prio = 5
	}
	if maxSize < 0 {
		maxSize = 0
	}
	logPrio = prio
	maxLogFileSize = maxSize
	logMux.Lock()
	logNum = 1
	if logFile == nil {
		// Check if log folder exists
		if _, err := os.Stat(logsFolder); os.IsNotExist(err) {
			// Create logs folder
			os.MkdirAll(logsFolder, os.ModePerm)
		} else {
			// Open log folder
			df, err := os.Open(logsFolder)
			if err != nil {
				logMux.Unlock()
				return err
			}
			// Get file list
			files, flErr := df.Readdir(-1)
			df.Close()
			if flErr != nil {
				logMux.Unlock()
				return flErr
			}
			// Check for other log files from today's date and get logNum
			for _, fileStats := range files {
				fileNameSplit := strings.Split(fileStats.Name(), ".")
				if fileNameSplit[0][:len(today)] == today {
					// Found log file from today.
					// Get logNum
					ln, lnErr := strconv.Atoi(fileNameSplit[0][len(today) + 3:])
					if lnErr != nil || len(fileNameSplit) < 2 || "." + fileNameSplit[1] != FileTypeLog {
						continue
					} else if logNum <= ln {
						logNum = ln + 1
					}
				}
			}
		}
		// Create new log file
		var err error
		if logFile, err = os.OpenFile(today + " : " + strconv.Itoa(logNum), os.O_CREATE, 0755); err != nil {
			logMux.Unlock()
			return err
		}
		// set byteOn and entryOn to 0
		byteOn = 0
		entryOn = 0
	}
	logMux.Unlock()
	logInit = true
	return nil
}

func Log(in string, priority int) {
	if !logInit || priority < logPrio {
		return
	}
	var now time.Time = time.Now()
	var today string = strconv.Itoa(now.Day()) + "/" + strconv.Itoa(now.Month()) + "/" + strconv.Itoa(now.Year())
	logMux.Lock()
	if logFile == nil {
		// Logger was not properly initialized - output error to console instead
		fmt.Println("[GopherDB Logger]: " + in)
		return
	} else if maxLogFileSize > 0 && entryOn >= maxLogFileSize {
		// Close current log file and increase logNum
		logFile.Close()
		logNum++
		// Create new log file
		var err error
		if logFile, err = os.OpenFile(today + " : " + strconv.Itoa(logNum), os.O_CREATE, 0755); err != nil {
			logMux.Unlock()
			return
		}
		// set byteOn and entryOn to 0
		byteOn = 0
		entryOn = 0
	}
	log := []byte(now.Format("2006-01-02T15:04:05Z07:00") + ": " + in + "\n")
	// Append log to logFile
	if _, wErr := logFile.WriteAt(log, byteOn); wErr != nil {
		logMux.Unlock()
		return
	}
	byteOn += len(log)
	entryOn++
	logMux.Unlock()
}

func CloseLogger() {
	logMux.Lock()
	logFile.Close()
	logMux.Unlock()
}