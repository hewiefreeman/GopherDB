package storage

import (
	"bufio"
	"io/ioutil"
	"os"
)

const (
	newLineIndicator byte = byte(10)
)

func ReadLine(file string, lineNum int) ([]byte) {
	r, err := os.Open(file)
	if err != nil {
		return nil
	}
	scanner := bufio.NewScanner(r)
	i := 0
	for scanner.Scan() {
		if i == lineNum {
			return scanner.Bytes()
		}
		i++
	}
	return nil
}

// UpdateJSON updates JSON encoded []byte lines at given indexes of given JSON file
func UpdateJSON(file string, jsonLines map[int][]byte) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	var newFileData []byte
	scanner := bufio.NewScanner(r)
	i := 0
	for scanner.Scan() {
		if jsonLines[i] == nil {
			newFileData = append(newFileData, append(scanner.Bytes(), newLineIndicator)...)

		} else {
			newFileData = append(newFileData, append(jsonLines[i], newLineIndicator)...)
		}
		i++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	r.Close()
	return ioutil.WriteFile(file, newFileData, 0644)
}

// AppendJSON appends a JSON encoded []byte at the end of given JSON file
func AppendJSON(file string, jsonLines [][]byte) error {
	r, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	newFileData := []byte{}
	for _, v := range jsonLines {
		newFileData = append(newFileData, append(v, newLineIndicator)...)
	}
	if _, err := r.Write(newFileData); err != nil {
		return err
	}
	r.Close();
	return nil
}