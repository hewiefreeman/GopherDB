package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var (
	newLineIndicator []byte = []byte{10}
)

func main() {
	json, jsonErr := json.Marshal(map[string]interface{}{"n": "someName", "pw": "somePassword", "email": "someEmailAddr@gmail.com", "vCode": "3RG71X", "verified": false, "friends": []interface{}{map[string]interface{}{"n": "some\nFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}}})
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return
	}

	// Insert JSON to line in file
	now := time.Now()
	insertErr := InsertJSONToFile("TestSmall.json", json, 5)
	if insertErr != nil {
		fmt.Println(insertErr)
		return
	}
	fmt.Println("Took:", time.Since(now).Seconds()*1000, "ms")
}

func GetFileLines(filePath string) ([][]byte, int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	return LinesFromReader(f)
}

func LinesFromReader(r io.Reader) ([][]byte, int, error) {
	var lines [][]byte
	var size int
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Bytes())
		size += len(scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	return lines, size, nil
}

func InsertJSONToFile(file string, json []byte, index int) error {
	lines, size, err := GetFileLines(file)
	if err != nil {
		return err
	}

	fileContent := make([]byte, size+len(lines)+len(json))
	i := 0
	for lineOn, line := range lines {
		if lineOn == index {
			for _, byteOn := range json {
				fileContent[i] = byteOn
				i++
			}
			fileContent[i] = byte(10)
			i++
		}
		for _, byteOn := range line {
			fileContent[i] = byteOn
			i++
		}
		if i < len(fileContent)-1 {
			fileContent[i] = byte(10)
			i++
		}

	}

	return ioutil.WriteFile(file, fileContent, 0644)
}
