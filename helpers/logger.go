package helpers

import (
	"time"
	"os"
	"sync"
	"strings"
	"strconv"
	"fmt"
)

const (
	logsFolder string = "logs"
	endOfLog   string = "   --- END OF LOG ---\n"
)

var (
	// Logger
	logFile *os.File
	logMux  sync.Mutex
	byteOn  int64
	entryOn int
	logNum  int
	logInit bool

	// Settings
	maxLogFileSize int // Maximum log messages in a single log file
	logPrio        int // Minimum priority level to display. Priority level can be from 1 to 5
)

// InitLogger initializes the logging system and returns an error if anything goes wrong. The logger can only be used if
// it is successfully initialized.
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
		now := time.Now()
		var today string = strconv.Itoa(now.Day()) + "-" + strconv.Itoa(int(now.Month())) + "-" + strconv.Itoa(now.Year())
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
					ln, lnErr := strconv.Atoi(fileNameSplit[0][len(today) + 2:len(fileNameSplit[0]) - 1])
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
		if logFile, err = os.OpenFile(logsFolder + "/" + today + " (" + strconv.Itoa(logNum) + ")" + FileTypeLog, os.O_RDWR | os.O_CREATE, 0755); err != nil {
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

// Log appends the given string to the current log file. If the priority level is less than the minimum priority level
// threshold, the log will not be written.
func Log(in string, priority int) bool {
	if !logInit || priority < logPrio {
		return false
	}
	var now time.Time = time.Now()
	logMux.Lock()
	if logFile == nil {
		return false
	} else if maxLogFileSize > 0 && entryOn >= maxLogFileSize {
		//
		logFile.WriteAt([]byte(endOfLog), byteOn)
		// Get today's date
		var today string = strconv.Itoa(now.Day()) + "-" + strconv.Itoa(int(now.Month())) + "-" + strconv.Itoa(now.Year())
		// Close current log file and increase logNum
		logFile.Close()
		logNum++
		// Create new log file
		var err error
		if logFile, err = os.OpenFile(logsFolder + "/" + today + " (" + strconv.Itoa(logNum) + ")" + FileTypeLog, os.O_RDWR | os.O_CREATE, 0755); err != nil {
			logMux.Unlock()
			return false
		}
		// set byteOn and entryOn to 0
		byteOn = 0
		entryOn = 0
	}
	log := []byte("[" + now.Format("2006-01-02T15:04:05Z07:00") + "]: " + in + "\n")
	// Append log to logFile
	if _, wErr := logFile.WriteAt(log, byteOn); wErr != nil {
		logMux.Unlock()
		return false
	}
	byteOn += int64(len(log))
	entryOn++
	logMux.Unlock()
	return true
}

func LogAndPrint(in string, priority int) bool {
	if !Log(in, priority) {
		return false
	}
	fmt.Println("[GopherDB Logger @ " + time.Now().Format("2006-01-02T15:04:05Z07:00") + "]: " + in)
	return true
}

// CloseLogger closes the logger. Any subsequent logs will only be output to the console.
func CloseLogger() {
	logMux.Lock()
	logFile.WriteAt([]byte(endOfLog), byteOn)
	logFile.Close()
	logFile = nil
	logMux.Unlock()
}