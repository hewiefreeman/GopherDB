package main

import (
	"github.com/hewiefreeman/GopherGameDB/userTable"
	"encoding/json"
	"fmt"
)

func main() {
	newTableJson := "{\"NewUserTable\": [\"users\",{\"email\": [\"String\", \"\", 0, true, true],\"friends\": [\"Array\", [\"Object\", {\"name\": [\"String\", \"\", 0, true, true],\"status\": [\"Number\", 0, 0, false]}], 50],\"vCode\": [\"String\", \"\", 0, true, false],\"verified\": [\"Bool\", false]}, 0, 0, 0, 0]}";
	v := make(map[string]interface{})
	err := json.Unmarshal([]byte(newTableJson), &v)
	if err != nil {
		fmt.Println(err)
		return
	}
	s := v["NewUserTable"].([]interface{})[1].(map[string]interface{})

	schema, schemaErr := userTable.NewSchema(s)
	if schemaErr != 0 {
		fmt.Println("Schema Error:", schemaErr)
		return
	}

	table, tableErr := userTable.New("users", schema, 6000, 0, 0, 0)
	if tableErr != 0 {
		fmt.Println("Table Create Error:", tableErr)
		return
	}

	table

	fmt.Println(schema)
}
