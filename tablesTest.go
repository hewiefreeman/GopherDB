package main

import (
	"github.com/hewiefreeman/GopherGameDB/userTable"
	"github.com/hewiefreeman/GopherGameDB/schema"
	"encoding/json"
	"fmt"
)

func main() {
	newTableJson := "{\"NewUserTable\": [\"users\",{\"email\": [\"String\", \"\", 0, true, true],\"friends\": [\"Array\", [\"Object\", {\"name\": [\"String\", \"\", 0, true, true],\"status\": [\"Number\", 0, 0, false, 0, 2]}, false], 50, false],\"vCode\": [\"String\", \"\", 0, true, false],\"verified\": [\"Bool\", false]}, 0, 0, 0, 0]}";
	v := make(map[string]interface{})
	err := json.Unmarshal([]byte(newTableJson), &v)
	if err != nil {
		fmt.Println(err)
		return
	}
	s := v["NewUserTable"].([]interface{})[1].(map[string]interface{})

	schemaObj, schemaErr := schema.New(s)
	if schemaErr != 0 {
		fmt.Println("Schema Error:", schemaErr)
		return
	}

	table, tableErr := userTable.New("users", schemaObj, 6000, 0, 0, 0, 0)
	if tableErr != 0 {
		fmt.Println("Table Create Error:", tableErr)
		return
	}

	insertErr := table.NewUser("hobo", "myPass", map[string]interface{}{"email": "hoboJenkins@yahoo.com", "vCode": "06AJ3T9", "friends": []interface{}{map[string]interface{}{"name": "freddy", "status": float64(-6)}}})
	if insertErr != 0 {
		fmt.Println("Insert Error:", insertErr)
		return
	}

	ue, ueErr := table.GetUserData("hobo", "myPass")
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}

	fmt.Println(ue)
}
