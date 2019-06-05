package main

import (
	"encoding/json"
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"fmt"
	"time"
)

var (
	newLineIndicator []byte = []byte{10}
)

func main() {
	json, jsonErr := json.Marshal(map[string]interface{}{"n":"someName","pw":"somePassword","email":"someEmailAddr@gmail.com","vCode":"3RG71X","verified":false,"friends":[]interface{}{map[string]interface{}{"n":"some\nFriend","s":2},map[string]interface{}{"n":"someFriend","s":2},map[string]interface{}{"n":"someFriend","s":2},map[string]interface{}{"n":"someFriend","s":2},map[string]interface{}{"n":"someFriend","s":2},map[string]interface{}{"n":"someFriend","s":2},map[string]interface{}{"n":"someFriend","s":2},map[string]interface{}{"n":"someFriend","s":2},map[string]interface{}{"n":"someFriend","s":2},map[string]interface{}{"n":"someFriend","s":2}}})
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

func GetFileLines(filePath string) ([][]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LinesFromReader(f)
}

func LinesFromReader(r io.Reader) ([][]byte, error) {
	var lines [][]byte
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func InsertJSONToFile(file string, json []byte, index int) error {
	lines, err := GetFileLines(file)
	if err != nil {
		return err
	}

	fileContent := []byte{}
	for i, line := range lines {
		if i == index {
			fileContent = append(fileContent, json...)
			fileContent = append(fileContent, 10)
		}
		fileContent = append(fileContent, line...)
		fileContent = append(fileContent, 10)
	}

	return ioutil.WriteFile(file, fileContent, 0644)
}