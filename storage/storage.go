package storage

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"bufio"
	"io/ioutil"
	"os"
	"errors"
)

const (
	newLineIndicator byte = byte(10)
	defaultFileActionBufferSize int = 100
)

// File actions
const (
	FileActionKill = iota
	FileActionRead
	FileActionInsert
	FileActionInsertMulti
	FileActionUpdate
	FileActionUpdateMulti
	FileActionMakeDir
	FileActionDeleteDir
	FileActionMakeFile
	FileActionDeleteFile
)

const (
	FileTypeLog     = ".gdbl"
	FileTypeStorage = ".gdbs"
)

var (
	FileActionChan chan fileAction
)

type fileAction struct {
	action int
	params []interface{}

	returnChan chan interface{}
}

func Start(bufferSize int) {
	if bufferSize < 1 {
		bufferSize = defaultFileActionBufferSize
	}
	FileActionChan = make(chan fileAction, bufferSize)
	go fileActionHandler()
}

func QueueFileAction(action int, params []interface{}, returnChan chan interface{}) int {
	fa := fileAction{action: action, params: params, returnChan: returnChan}
	select {
		case FileActionChan <- fa:
			return 0

		default:
			return helpers.ErrorDatabaseBusy
    }
}

//
func fileActionHandler() {
	for {
		action := <-FileActionChan
		switch action.action {
			case FileActionRead:
				if len(action.params) != 2 {
					action.returnChan <- []byte{}
					continue
				}
				bytes := ReadLine(action.params[0].(string), int(action.params[1].(uint16)))
				action.returnChan <- bytes

			case FileActionInsert:
				if len(action.params) != 2 {
					action.returnChan <- []interface{}{0, errors.New("Invalid Parameters")}
					continue
				}
				lineOn, err := AppendJSON(action.params[0].(string), action.params[1].([]byte))
				action.returnChan <- []interface{}{lineOn, err}

			case FileActionInsertMulti:
				if len(action.params) != 2 {
					action.returnChan <- errors.New("Invalid Parameters")
					continue
				}
				err := AppendJSONMulti(action.params[0].(string), action.params[1].([][]byte))
				action.returnChan <- err

			case FileActionUpdate:
				if len(action.params) != 3 {
					action.returnChan <- errors.New("Invalid Parameters")
					continue
				}
				err := UpdateJSON(action.params[0].(string), action.params[1].(uint16), action.params[2].([]byte))
				action.returnChan <- err

			case FileActionUpdateMulti:
				if len(action.params) != 2 {
					action.returnChan <- errors.New("Invalid Parameters")
					continue
				}
				err := UpdateJSONMulti(action.params[0].(string), action.params[1].(map[int][]byte))
				action.returnChan <- err

			case FileActionMakeDir:
				if len(action.params) != 1 {
					action.returnChan <- errors.New("Invalid Parameters")
					continue
				}
				err := MakeDir(action.params[0].(string))
				action.returnChan <- err

			case FileActionDeleteDir:
				if len(action.params) != 1 {
					action.returnChan <- errors.New("Invalid Parameters")
					continue
				}
				err := DeleteDir(action.params[0].(string))
				action.returnChan <- err

			case FileActionMakeFile:
				if len(action.params) != 1 {
					action.returnChan <- errors.New("Invalid Parameters")
					continue
				}
				err := MakeFile(action.params[0].(string))
				action.returnChan <- err

			case FileActionDeleteFile:
				if len(action.params) != 1 {
					action.returnChan <- errors.New("Invalid Parameters")
					continue
				}
				err := DeleteFile(action.params[0].(string))
				action.returnChan <- err

			case FileActionKill:
				close(FileActionChan)
				action.returnChan <- nil
				return

			default:
				action.returnChan <- errors.New("Invalid file action")
		}
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

// MakeFile creates a file on the disk/drive
func MakeFile(file string) error {
	r, err := os.Create(file)
	r.Close()
	return err
}

// DeleteFile deletes a file from the disk/drive
func DeleteFile(file string) error {
	err := os.Remove(file)
	return err
}

// ReadLine reads a specific line from a file.
func ReadLine(file string, lineNum int) []byte {
	r, err := os.Open(file)
	if err != nil {
		return nil
	}
	scanner := bufio.NewScanner(r)
	i := 1
	for scanner.Scan() {
		if i == lineNum {
			r.Close()
			return scanner.Bytes()
		}
		i++
	}
	r.Close()
	return nil
}

// UpdateJSON updates JSON encoded []byte lines at given indexes of given JSON file
func UpdateJSON(file string, index uint16, json []byte) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	var newFileData []byte
	scanner := bufio.NewScanner(r)
	var i uint16 = 1
	for scanner.Scan() {
		if i == index {
			newFileData = append(newFileData, append(json, newLineIndicator)...)
		} else {
			newFileData = append(newFileData, append(scanner.Bytes(), newLineIndicator)...)
		}
		i++
	}
	if err := scanner.Err(); err != nil {
		r.Close()
		return err
	}
	r.Close()
	return ioutil.WriteFile(file, newFileData, 0644)
}

// UpdateJSONMulti updates JSON encoded []byte lines at given indexes of given JSON file
func UpdateJSONMulti(file string, jsonLines map[int][]byte) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	var newFileData []byte
	scanner := bufio.NewScanner(r)
	i := 1
	for scanner.Scan() {
		if jsonLines[i] == nil {
			newFileData = append(newFileData, append(scanner.Bytes(), newLineIndicator)...)

		} else {
			newFileData = append(newFileData, append(jsonLines[i], newLineIndicator)...)
		}
		i++
	}
	if err := scanner.Err(); err != nil {
		r.Close()
		return err
	}
	r.Close()
	return ioutil.WriteFile(file, newFileData, 0644)
}

// AppendJSON appends a JSON encoded []byte at the end of given JSON file
func AppendJSON(file string, json []byte) (uint16, error) {
	r, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return 0, err
	}
	scanner := bufio.NewScanner(r)
	var lineOn uint16 = 1
	var dLen int64
	for scanner.Scan() {
		dLen += int64(len(scanner.Bytes()) + 1)
		lineOn++
	}
	if sErr := scanner.Err(); sErr != nil {
		r.Close()
		return 0, sErr
	}
	json = append(json, newLineIndicator)
	if _, rErr := r.WriteAt(json, dLen); rErr != nil {
		r.Close()
		return 0, rErr
	}
	r.Close()
	return lineOn, nil
}

// AppendJSONMulti appends multiple JSON encoded []bytes to the end of given JSON file
func AppendJSONMulti(file string, jsonLines [][]byte) error {
	r, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	newFileData := []byte{}
	for _, v := range jsonLines {
		newFileData = append(newFileData, append(v, newLineIndicator)...)
	}
	if _, err := r.Write(newFileData); err != nil {
		r.Close()
		return err
	}
	r.Close()
	return nil
}