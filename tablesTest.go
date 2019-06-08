package main

import (
	"github.com/hewiefreeman/GopherGameDB/schema"
	"github.com/hewiefreeman/GopherGameDB/userTable"
	//"github.com/hewiefreeman/GopherGameDB/helpers"
	"encoding/json"
	"fmt"
	"time"
)

type ii struct {
	name string
	pass string
}

var (
	insertItems = []ii{
		ii{name: "lilly", pass: "4e5yhrthg"},
		ii{name: "harry", pass: "rtghsr5thh"},
		ii{name: "potter", pass: "sr5thshrth"},
		ii{name: "andthe", pass: "4r5yhshsfdgh"},
		ii{name: "sorcerers", pass: "5tj6trhrtyh"},
		ii{name: "stoned", pass: "uyh4rhrthg"},
		ii{name: "baked", pass: "rgse5g5r4"},
		ii{name: "seared", pass: "sret5yhtrs"},
		ii{name: "crispened", pass: "r6ujh6thys"},
		ii{name: "dead", pass: "rtgdfg34"},
		ii{name: "alive", pass: "2w3rfwefaw"},
		ii{name: "killed", pass: "gvfbhghrt6hf"},
		ii{name: "rezed", pass: "sergdfv54re4"},
		ii{name: "groper", pass: "sdfgsegsdf"},
		ii{name: "amazingJoe", pass: "sdfbserg55h"},
		ii{name: "badass", pass: "hj65rngn4edx"},
		ii{name: "stupid", pass: "imwith"},
		ii{name: "lopl", pass: "loling"},
		ii{name: "wtlf", pass: "whatthe"},
		ii{name: "smirnoffChugger", pass: "holyballs"},
		ii{name: "67jdsrthndt", pass: "sdfgsegsdf"},
		ii{name: "grasgaeoper", pass: "sdfgsegsdf"},
		ii{name: "hth5groper", pass: "sdfgsegsdf"},
		ii{name: "grophgfser", pass: "sdfgsegsdf"},
		ii{name: "grxcvbtroper", pass: "sdfgsegsdf"},
		ii{name: "groj86per", pass: "sdfgsegsdf"},
		ii{name: "234tgroper", pass: "sdfgsegsdf"},
		ii{name: "grop653er", pass: "sdfgsegsdf"},
		ii{name: "grdhj65roper", pass: "sdfgsegsdf"},
		ii{name: "grocxju6per", pass: "sdfgsegsdf"},
		ii{name: "gromr65per", pass: "sdfgsegsdf"},
		ii{name: "grope2verr", pass: "sdfgsegsdf"},
		ii{name: "grop8986er", pass: "sdfgsegsdf"},
		ii{name: "grop12346er", pass: "sdfgsegsdf"},
		ii{name: "gropasdgh5er", pass: "sdfgsegsdf"},
		ii{name: "grop56yherser", pass: "sdfgsegsdf"},
		ii{name: "grop234ter", pass: "sdfgsegsdf"},
		ii{name: "gropyuh64er", pass: "sdfgsegsdf"},
		ii{name: "grop2q3rer", pass: "sdfgsegsdf"},
		ii{name: "groyh654rper", pass: "sdfgsegsdf"},
		ii{name: "67j1dsrthndt", pass: "sdfgsegsdf"},
		ii{name: "gras2gaeoper", pass: "sdfgsegsdf"},
		ii{name: "hth5g3roper", pass: "sdfgsegsdf"},
		ii{name: "grophg4fser", pass: "sdfgsegsdf"},
		ii{name: "grxcvbt5roper", pass: "sdfgsegsdf"},
		ii{name: "groj86p6er", pass: "sdfgsegsdf"},
		ii{name: "234tgro7per", pass: "sdfgsegsdf"},
		ii{name: "grop6538er", pass: "sdfgsegsdf"},
		ii{name: "grdhj659roper", pass: "sdfgsegsdf"},
		ii{name: "grocxju06per", pass: "sdfgsegsdf"},
		ii{name: "gromr65-per", pass: "sdfgsegsdf"},
		ii{name: "grop1e2verr", pass: "sdfgsegsdf"},
		ii{name: "grop28986er", pass: "sdfgsegsdf"},
		ii{name: "grop312346er", pass: "sdfgsegsdf"},
		ii{name: "grop4asdgh5er", pass: "sdfgsegsdf"},
		ii{name: "grop556yherser", pass: "sdfgsegsdf"},
		ii{name: "grop6234ter", pass: "sdfgsegsdf"},
		ii{name: "grop7yuh64er", pass: "sdfgsegsdf"},
		ii{name: "grop82q3rer", pass: "sdfgsegsdf"},
		ii{name: "groy9h654rper", pass: "sdfgsegsdf"},
		ii{name: "g1rop2q3rer", pass: "sdfgsegsdf"},
		ii{name: "g2royh654rper", pass: "sdfgsegsdf"},
		ii{name: "637j1dsrthndt", pass: "sdfgsegsdf"},
		ii{name: "g4ras2gaeoper", pass: "sdfgsegsdf"},
		ii{name: "h5th5g3roper", pass: "sdfgsegsdf"},
		ii{name: "g6rophg4fser", pass: "sdfgsegsdf"},
		ii{name: "g7rxcvbt5roper", pass: "sdfgsegsdf"},
		ii{name: "g8roj86p6er", pass: "sdfgsegsdf"},
		ii{name: "2934tgro7per", pass: "sdfgsegsdf"},
		ii{name: "g0rop6538er", pass: "sdfgsegsdf"},
		ii{name: "g1rdhj659roper", pass: "sdfgsegsdf"},
		ii{name: "gr2ocxju06per", pass: "sdfgsegsdf"},
		ii{name: "g3romr65-per", pass: "sdfgsegsdf"},
		ii{name: "g4rop1e2verr", pass: "sdfgsegsdf"},
		ii{name: "g5rop28986er", pass: "sdfgsegsdf"},
		ii{name: "g6rop312346er", pass: "sdfgsegsdf"},
		ii{name: "g7rop4asdgh5er", pass: "sdfgsegsdf"},
		ii{name: "g8rop556yherser", pass: "sdfgsegsdf"},
		ii{name: "g9rop6234ter", pass: "sdfgsegsdf"},
		ii{name: "g0rop7yuh64er", pass: "sdfgsegsdf"},
		ii{name: "g1rop82q3rer", pass: "sdfgsegsdf"},
		ii{name: "g2roy9h654rper", pass: "sdfgsegsdf"},
	}
)

func main() {
	// JSON query and unmarshalling test
	newTableJson := "{\"NewUserTable\": [\"users\",{\"email\": [\"String\", \"\", 0, true, true],\"friends\": [\"Array\", [\"Object\", {\"name\": [\"String\", \"\", 0, true, true],\"status\": [\"Uint8\", 0, 0, 2, false]}, false], 50, false],\"vCode\": [\"String\", \"\", 0, true, false],\"verified\": [\"Bool\", false], \"mmr\": [\"Uint16\", 1500, 1100, 2250, false], \"testMap\": [\"Map\", [\"Map\", [\"Uint16\", 100, 0, 0, false], 0, false], 0, false]}, 0, 0, 0, 0]}"
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

	// More inserts
	var averageTime float64
	for v := range insertItems {
		now := time.Now()
		// Insert a test User
		insertErr := table.NewUser(insertItems[v].name, insertItems[v].pass, map[string]interface{}{"email": "dinospumoni@yahoo.com", "mmr": 1674, "vCode": "06AJ3T9"})
		if insertErr != 0 {
			fmt.Println("Insert Error:", insertErr)
			return
		}
		if averageTime == 0 {
			averageTime = time.Since(now).Seconds()
		} else {
			averageTime = (averageTime + time.Since(now).Seconds()) / 2
		}
	}
	fmt.Println("Average insert time (ms):", averageTime*1000)

	averageTime = 0
	for v := range insertItems {
		now := time.Now()
		// add 1 to entry's mmr
		updateErr := table.UpdateUserData(insertItems[v].name, insertItems[v].pass, map[string]interface{}{"mmr.*add": []interface{}{2}})
		if updateErr != 0 {
			fmt.Println("Update Error:", updateErr)
			return
		}
		if averageTime == 0 {
			averageTime = time.Since(now).Seconds()
		} else {
			averageTime = (averageTime + time.Since(now).Seconds()) / 2
		}
	}
	fmt.Println("Average update time (ms):", averageTime*1000)

	averageTime = 0
	for v := range insertItems {
		now := time.Now()
		_, ueErr := table.GetUserData(insertItems[v].name, insertItems[v].pass)
		if ueErr != 0 {
			fmt.Println("User Data Error:", ueErr)
			return
		}
		if averageTime == 0 {
			averageTime = time.Since(now).Seconds()
		} else {
			averageTime = (averageTime + time.Since(now).Seconds()) / 2
		}
	}
	fmt.Println("Average select time (ms):", averageTime*1000)

	ud, ueErr := table.GetUserData("wtlf", "whatthe")
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}
	fmt.Println("Before:", ud)

	// Multiply by 1.5, divide by 2, add 4, then subtract 1 from entry's mmr (using methods)
	updateErr := table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"mmr.*mul.*div.*add.*sub": []interface{}{1.5, 2, 4, 1}})
	if updateErr != 0 {
		fmt.Println("Update Error 1:", updateErr)
		return
	}

	// Append a friend to friends
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"friends.*append": []interface{}{map[string]interface{}{"name": "Mag"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 2:", updateErr)
		return
	}

	// Prepend a friend to friends
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"friends.*prepend": []interface{}{map[string]interface{}{"name": "Jason"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 3:", updateErr)
		return
	}

	// Append 2 friends to index 1 of friends
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"friends.*append[1]": []interface{}{map[string]interface{}{"name": "Harry"}, map[string]interface{}{"name": "Potter"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 4:", updateErr)
		return
	}

	// Delete 2 friends from friends
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"friends.*delete": []interface{}{2, 1}})
	if updateErr != 0 {
		fmt.Println("Update Error 5:", updateErr)
		return
	}

	// Chage name of friend at index 1 of friends
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"friends.1.name.*append": []interface{}{"icarp"}})
	if updateErr != 0 {
		fmt.Println("Update Error 6:", updateErr)
		return
	}

	// Chage status of friend at index 0 of friends
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"friends.0.status": 3})
	if updateErr != 0 {
		fmt.Println("Update Error 7:", updateErr)
		return
	}

	// Add something to testMap
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"testMap.*append": map[string]interface{}{"items": map[string]interface{}{}}})
	if updateErr != 0 {
		fmt.Println("Update Error 8:", updateErr)
		return
	}

	// Add something to items in testMap
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"testMap.items.arrows": 12})
	if updateErr != 0 {
		fmt.Println("Update Error 8:", updateErr)
		return
	}

	// Add something to items in testMap
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"testMap.items.beanz": 87})
	if updateErr != 0 {
		fmt.Println("Update Error 9:", updateErr)
		return
	}

	// Apply arithmetic to beanz in items in testMap
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"testMap.items.beanz.*add.*mul": []interface{}{3, 2}})
	if updateErr != 0 {
		fmt.Println("Update Error 10:", updateErr)
		return
	}

	// Delete arrows in items in testMap
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"testMap.items.*delete": []interface{}{"arrows"}})
	if updateErr != 0 {
		fmt.Println("Update Error 11:", updateErr)
		return
	}

	// Append rupees and silk to items
	updateErr = table.UpdateUserData("wtlf", "whatthe", map[string]interface{}{"testMap.items.*append": map[string]interface{}{"rupees": 99, "silk": 1}})
	if updateErr != 0 {
		fmt.Println("Update Error 12:", updateErr)
		return
	}

	ud, ueErr = table.GetUserData("wtlf", "whatthe")
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}
	fmt.Println("After:", ud)
}
