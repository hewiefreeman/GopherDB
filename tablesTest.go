package main

import (
	"github.com/hewiefreeman/GopherGameDB/userTable"
	"github.com/hewiefreeman/GopherGameDB/schema"
	"encoding/json"
	"fmt"
)

func main() {
	// JSON query and unmarshalling test
	newTableJson := "{\"NewUserTable\": [\"users\",{\"email\": [\"String\", \"\", 0, true, true],\"friends\": [\"Array\", [\"Object\", {\"name\": [\"String\", \"\", 0, true, true],\"status\": [\"Uint8\", 0, 0, 2, false]}, false], 50, false],\"vCode\": [\"String\", \"\", 0, true, false],\"verified\": [\"Bool\", false], \"mmr\": [\"Uint16\", 0, 0, 0, false]}, 0, 0, 0, 0]}";
	v := make(map[string]interface{})
	err := json.Unmarshal([]byte(newTableJson), &v)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the schema object from the query
	s := v["NewUserTable"].([]interface{})[1].(map[string]interface{})

	// Make a schema with the query's schema object
	schemaObj, schemaErr := schema.New(s)
	if schemaErr != 0 {
		fmt.Println("Schema Error:", schemaErr)
		return
	}

	// Make a new UserTable with the schema
	table, tableErr := userTable.New("users", schemaObj, 6000, 0, 0, 0, 0)
	if tableErr != 0 {
		fmt.Println("Table Create Error:", tableErr)
		return
	}

	// Insert a test User
	insertErr := table.NewUser("DinoSpumoni", "isthatjazz", map[string]interface{}{"email": "dinospumoni@yahoo.com", "mmr": float64(1500.87), "vCode": "06AJ3T9", "friends": []interface{}{map[string]interface{}{"name": "Arnold", "status": float64(5)}}})
	if insertErr != 0 {
		fmt.Println("Insert Error:", insertErr)
		return
	}

	updateErr := table.UpdateUserData("DinoSpumoni", "isthatjazz", map[string]interface{}{"mmr": []interface{}{"%", float64(7)}})
	if updateErr != 0 {
		fmt.Println("Update Error:", updateErr)
		return
	}

	// Retrieve the test User
	ue, ueErr := table.GetUserData("DinoSpumoni", "isthatjazz")
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}

	// Consume ue for compiler
	fmt.Println(ue)
}
