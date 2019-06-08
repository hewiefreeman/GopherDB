package main

import (
	"fmt"
	"github.com/hewiefreeman/GopherGameDB/leaderboards"
	"time"
)

func main() {
	lb, err := leaderboards.New("test", 30, leaderboards.DuplicateTargetPushBottom, false)
	if err != 0 {
		fmt.Println("Error: ", err)
		return
	}
	now := time.Now()
	lb.CheckAndPush("Billy", 1432.5, make(map[string]interface{}))
	lb.CheckAndPush("Bob", 1000, make(map[string]interface{}))
	lb.CheckAndPush("Bush", 2312, make(map[string]interface{}))
	lb.CheckAndPush("Boris", 1652.3, make(map[string]interface{}))
	lb.CheckAndPush("Billy", 1431.2, make(map[string]interface{}))
	lb.CheckAndPush("Bastard", 1423.8, make(map[string]interface{}))
	lb.CheckAndPush("Bunk", 1732, make(map[string]interface{}))
	lb.CheckAndPush("Brad", 1323, make(map[string]interface{}))
	lb.CheckAndPush("Bruno", 1142, make(map[string]interface{}))
	lb.CheckAndPush("Bweeb", 1432, make(map[string]interface{}))
	lb.CheckAndPush("Bhem", 1645, make(map[string]interface{}))
	lb.CheckAndPush("Bjorn", 1112, make(map[string]interface{}))
	lb.CheckAndPush("Bush", 1112, make(map[string]interface{}))
	lb.CheckAndPush("Soma", 1112, make(map[string]interface{}))
	lb.CheckAndPush("KillaG", 2112, make(map[string]interface{}))
	lb.CheckAndPush("OpNoob", 2112, make(map[string]interface{}))
	lb.CheckAndPush("Bhem", 900, make(map[string]interface{}))
	lb.CheckAndPush("Guest435234", 2341, make(map[string]interface{}))
	lb.CheckAndPush("Guest4t34", 1423, make(map[string]interface{}))
	lb.CheckAndPush("Guest43drtr34", 1612, make(map[string]interface{}))
	lb.CheckAndPush("Guest43434", 1515, make(map[string]interface{}))
	lb.CheckAndPush("Guest4656434", 1726, make(map[string]interface{}))
	lb.CheckAndPush("Guest432334", 1626, make(map[string]interface{}))
	lb.CheckAndPush("Guest72234", 1876, make(map[string]interface{}))
	lb.CheckAndPush("Guest1234", 1672, make(map[string]interface{}))
	lb.CheckAndPush("Guest43555234", 4321, make(map[string]interface{}))
	lb.CheckAndPush("Guest87634", 1234, make(map[string]interface{}))
	lb.CheckAndPush("Guest456235234", 2345, make(map[string]interface{}))
	lb.CheckAndPush("Guest62652", 2345, make(map[string]interface{}))
	lb.CheckAndPush("Guest0954", 2341, make(map[string]interface{}))
	lb.CheckAndPush("Guest76746", 2341, make(map[string]interface{}))
	lb.CheckAndPush("Guest52434", 1562, make(map[string]interface{}))
	lb.CheckAndPush("Guest00132", 1893, make(map[string]interface{}))
	lb.CheckAndPush("Guest66453", 2000, make(map[string]interface{}))
	lb.CheckAndPush("Guest22452", 1982, make(map[string]interface{}))
	lb.CheckAndPush("Guest161612", 1900, make(map[string]interface{}))
	lb.CheckAndPush("Guest58363", 1800, make(map[string]interface{}))
	lb.CheckAndPush("Guest0546212", 1700, make(map[string]interface{}))
	timePassed := time.Since(now)
	fmt.Println(timePassed)

	//
	lb.Print()
	fmt.Println(lb.GetPage(10, 0))

}
