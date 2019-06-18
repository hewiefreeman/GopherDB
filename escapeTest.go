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
	line := storage.ReadLine("Test.json", 2503)
	if line == nil {
		fmt.Println("No item found in file!")
		return
	}
	fmt.Println("Took:", time.Since(now).Seconds()*1000, "ms")
	fmt.Println("Found:", string(line))
}