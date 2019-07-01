package main

import (
	"github.com/hewiefreeman/GopherDB/storage"
	//"encoding/json"
	"time"
	"fmt"
)

var (
	newLineIndicator byte = byte(10)
)

func main() {
	/*json, jsonErr := json.Marshal(map[string]interface{}{"n": "someName", "pw": "somePassword", "email": "thisdude@gmail.com", "vCode": "3RG71X", "verified": false, "friends": []interface{}{map[string]interface{}{"n": "some\nFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}, map[string]interface{}{"n": "someFriend", "s": 2}}})
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return
	}*/

	// Test storage functions...
	now := time.Now()
	/*err := */storage.ReadLine("Test.json", 2512)
	/*if err != nil {
		fmt.Println(err)
		return
	}*/

	// Test make file in directory that doesn't exist
	/*now := time.Now()
	err := storage.DeleteDir("test123")
	if err != nil {
		fmt.Println(err)
		return
	}*/

	fmt.Println("Took:", time.Since(now).Seconds()*1000, "ms")
}