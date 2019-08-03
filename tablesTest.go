package main

import (
	"github.com/hewiefreeman/GopherDB/schema"
	"github.com/hewiefreeman/GopherDB/authtable"
	"encoding/json"
	"strconv"
	"fmt"
	"time"
)

func main() {
	// JSON query and unmarshalling test
	newTableJson := "{\"NewAuthTable\": [\"users\",{\"email\": [\"String\", \"\", 0, true, true],\"friends\": [\"Array\", [\"Object\", {\"name\": [\"String\", \"\", 0, true, true],\"status\": [\"Uint8\", 0, 0, 2, false, false]}, false], 50, false],\"vCode\": [\"String\", \"\", 0, true, false],\"verified\": [\"Bool\", false], \"mmr\": [\"Uint16\", 1500, 1100, 2250, false, false], \"testMap\": [\"Map\", [\"Map\", [\"Int16\", 100, 0, 0, true, true, true], 0, false], 0, false], \"timeStamp\": [\"Time\", \"Kitchen\", false]}, 0, 0, 0, 0]}"
	v := make(map[string]interface{})
	err := json.Unmarshal([]byte(newTableJson), &v)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the schema object from the query
	s := v["NewAuthTable"].([]interface{})[1].(map[string]interface{})

	// Make a schema with the query's schema object
	schemaObj, schemaErr := schema.New(s)
	if schemaErr != 0 {
		fmt.Println("Schema Error:", schemaErr)
		return
	}

	// Make a new AuthTable with the schema
	table, tableErr := authtable.New("users", schemaObj, 0, 0, 0, 0, false)
	if tableErr != 0 {
		fmt.Println("Table Create Error:", tableErr)
		return
	}

	// Table settings
	alErr := table.SetAltLoginItem("email")
	if alErr != 0 {
		fmt.Println("Set Login item failure:", alErr)
		return
	}

	testSize := 100

	// insert
	var averageTime float64
	now := time.Now()
	done := make(chan int)
	for v := 0; v < testSize; v++ {
		// Insert a test User
		go func(a int){
			done <- table.NewUser("guest"+strconv.Itoa(a), "myPass", map[string]interface{}{"email": "dinospumoni"+strconv.Itoa(a)+"@yahoo.com", "mmr": 1674+v, "vCode": "06AJ3T9"})
		}(v)
	}
	for v := 0; v < testSize; v++ {
		res := <-done
		if res != 0 {
			fmt.Println("Insert Error: ", res)
			close(done)
			return
		}
	}
	close(done)
	averageTime = time.Since(now).Seconds()*1000
	fmt.Println("Bulk insert time (ms):", averageTime)

	averageTime = 0
	now = time.Now()
	done = make(chan int, 100)
	for v := 0; v < testSize; v++ {
		go func(v int){
			done <- table.UpdateUserData("guest"+strconv.Itoa(v), "myPass", map[string]interface{}{"mmr.*add": []interface{}{2}})
		}(v)
	}
	for v := 0; v < testSize; v++ {
		res := <-done
		if res != 0 {
			fmt.Println("Update Error: ", res)
			close(done)
			return
		}
	}
	close(done)
	averageTime = time.Since(now).Seconds()*1000
	fmt.Println("Bulk update time (ms):", averageTime)

	averageTime = 0
	now = time.Now()
	done = make(chan int, 100)
	for v := 0; v < testSize; v++ {
		go func(v int){
			done <- table.UpdateUserData("guest"+strconv.Itoa(v), "myPass", map[string]interface{}{"mmr.*add": []interface{}{2}})
		}(v)
	}
	for v := 0; v < testSize; v++ {
		res := <-done
		if res != 0 {
			fmt.Println("Update 2 Error: ", res)
			close(done)
			return
		}
	}
	close(done)
	averageTime = time.Since(now).Seconds()*1000
	fmt.Println("Bulk update 2 time (ms):", averageTime)

	averageTime = 0
	now = time.Now()
	done = make(chan int)
	for v := 0; v < testSize; v++ {
		go func(v int){
			_, ueErr := table.GetUserData("guest"+strconv.Itoa(v), "myPass", []string{"verified"})
			done <- ueErr
		}(v)
	}
	for v := 0; v < testSize; v++ {
		res := <-done
		if res != 0 {
			fmt.Println("Get Error: ", res)
			close(done)
			return
		}
	}
	close(done)
	averageTime = time.Since(now).Seconds()*1000
	fmt.Println("Bulk get time (ms):", averageTime)

	averageTime = 0
	now = time.Now()
	done = make(chan int)
	for v := 0; v < testSize; v++ {
		go func(v int){
			_, ueErr := table.GetUserData("guest"+strconv.Itoa(v), "myPass", []string{"verified"})
			done <- ueErr
		}(v)
	}
	for v := 0; v < testSize; v++ {
		res := <-done
		if res != 0 {
			fmt.Println("Get 2 Error: ", res)
			close(done)
			return
		}
	}
	close(done)
	averageTime = time.Since(now).Seconds()*1000
	fmt.Println("Bulk get 2 time (ms):", averageTime)

	ud, ueErr := table.GetUserData("dinospumoni99@yahoo.com", "myPass", nil)
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}
	fmt.Println("Before:", ud)

	// Multiply by 1.5, divide by 2, add 4, then subtract 1 from entry's mmr (using methods)
	updateErr := table.UpdateUserData("guest99", "myPass", map[string]interface{}{"mmr.*mul.*div.*add.*sub": []interface{}{1.5, 2, 4, 1}})
	if updateErr != 0 {
		fmt.Println("Update Error 1:", updateErr)
		return
	}

	// Append a friend to friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.*append": []interface{}{map[string]interface{}{"name": "Mag"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 2:", updateErr)
		return
	}

	// Prepend a friend to friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.*prepend": []interface{}{map[string]interface{}{"name": "Jason"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 3:", updateErr)
		return
	}

	// Append 2 friends to index 1 of friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.*append[1]": []interface{}{map[string]interface{}{"name": "Harry"}, map[string]interface{}{"name": "Potter"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 4:", updateErr)
		return
	}

	// Delete 2 friends from friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.*delete": []interface{}{2, 1}})
	if updateErr != 0 {
		fmt.Println("Update Error 5:", updateErr)
		return
	}

	// Chage name of friend at index 1 of friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.1.name.*append": []interface{}{"icarp"}})
	if updateErr != 0 {
		fmt.Println("Update Error 6:", updateErr)
		return
	}

	// Chage status of friend at index 0 of friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.0.status": 3})
	if updateErr != 0 {
		fmt.Println("Update Error 7:", updateErr)
		return
	}

	// Add something to testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.*append": map[string]interface{}{"items": map[string]interface{}{}}})
	if updateErr != 0 {
		fmt.Println("Update Error 8:", updateErr)
		return
	}

	// Add something to items in testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.arrows": 12})
	if updateErr != 0 {
		fmt.Println("Update Error 8:", updateErr)
		return
	}

	// Add something to items in testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.beanz": 87})
	if updateErr != 0 {
		fmt.Println("Update Error 9:", updateErr)
		return
	}

	// Apply arithmetic to beanz in items in testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.beanz.*add.*mul": []interface{}{3, 2}})
	if updateErr != 0 {
		fmt.Println("Update Error 10:", updateErr)
		return
	}

	// Delete arrows in items in testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.*delete": []interface{}{"arrows"}})
	if updateErr != 0 {
		fmt.Println("Update Error 11:", updateErr)
		return
	}

	// Append rupees and silk to items
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.*append": map[string]interface{}{"rupees": 99, "silk": 1}})
	if updateErr != 0 {
		fmt.Println("Update Error 12:", updateErr)
		return
	}

	// Set timeStamp manually
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"timeStamp": "*now"})
	if updateErr != 0 {
		fmt.Println("Update Error 13:", updateErr)
		return
	}

	// Add friend
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.*append": []interface{}{map[string]interface{}{"name": "as"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 14:", updateErr)
		return
	}

	// Change email item (also altLoginItem & unique)
	now = time.Now()
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"email": "someemail@yahoo.com"})
	if updateErr != 0 {
		fmt.Println("Update Error 15:", updateErr)
		return
	}
	fmt.Println("Email update took", (time.Since(now).Seconds() * 1000), "ms")

	// Delete a AuthTable entry
	deleteErr := table.DeleteUser("guest98", "myPass")
	if deleteErr != 0 {
		fmt.Println("Update Error 14:", deleteErr)
		return
	}
	fmt.Println("Delete success!")

	// Try to get deleted account
	ud, ueErr = table.GetUserData("dinospumoni98@yahoo.com", "myPass", []string{"email"})
	if ueErr != 0 {
		fmt.Println("Error getting deleted account:", ueErr)
	}

	now = time.Now()
	ud, ueErr = table.GetUserData("someemail@yahoo.com", "myPass", []string{"timeStamp.*since.*mil"})
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}
	fmt.Println("Get took", (time.Since(now).Seconds() * 1000), "ms")
	fmt.Println("After:", ud)

	ud, ueErr = table.GetUserData("someemail@yahoo.com", "myPass", []string{"timeStamp.*since.*mil"})
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}
	time.Sleep(time.Second * 1)
	ud, ueErr = table.GetUserData("someemail@yahoo.com", "myPass", []string{"timeStamp.*since.*mil"})
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}
}
