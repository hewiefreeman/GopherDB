package main

import (
	"github.com/hewiefreeman/GopherGameDB/userTable"
	"encoding/json"
	"fmt"
)

func main() {
	newTableJson := "{\"NewUserTable\": [\"users\",{\"email\": [\"String\", \"\", 0, true, true],\"friends\": [\"Array\", [\"Object\", {\"name\": [\"String\", \"\", 0, true, true],\"status\": [\"Number\", 0, 0, false]}], 50],\"vCode\": [\"String\", \"\", 0, true, false],\"verified\": [\"Bool\", false]}]}";
	v := make(map[string]interface{})
	err := json.Unmarshal([]byte(newTableJson), &v)
	if err != nil {
		fmt.Println(err)
		return
	}
	s := v["NewUserTable"].([]interface{})[1].(map[string]interface{})

	schema, schemaErr := userTable.NewSchema(s)
	if schemaErr != 0 {
		fmt.Println("Error #:", schemaErr)
		return
	}

	fmt.Println(schema)
}
