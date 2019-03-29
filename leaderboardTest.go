package main

import (
	"github.com/hewiefreeman/GopherGameDB/leaderboards"
	"fmt"
	"time"
)

func main() {
	lb, err := leaderboards.New("test", 30, leaderboards.DuplicateTargetPushTop, true)
	if err != 0 {
		fmt.Println("Error: ", err)
		return
	}
	lb.CheckAndPush("Billy", 1432, make(map[string]interface{}))
	lb.CheckAndPush("Bob", 1000, make(map[string]interface{}))
	lb.CheckAndPush("Bush", 2312, make(map[string]interface{}))
	lb.CheckAndPush("Boris", 1652, make(map[string]interface{}))
	lb.CheckAndPush("Billy", 1431, make(map[string]interface{}))
	lb.CheckAndPush("Bastard", 1423, make(map[string]interface{}))
	lb.CheckAndPush("Bunk", 1732, make(map[string]interface{}))
	lb.CheckAndPush("Brad", 1323, make(map[string]interface{}))
	lb.CheckAndPush("Bruno", 1142, make(map[string]interface{}))
	lb.CheckAndPush("Bweeb", 1432, make(map[string]interface{}))
	lb.CheckAndPush("Bhem", 1645, make(map[string]interface{}))
	lb.CheckAndPush("Bjorn", 1112, make(map[string]interface{}))
	lb.CheckAndPush("Bush", 1212, make(map[string]interface{}))
	lb.CheckAndPush("Bjorn", 2112, make(map[string]interface{}))
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
	lb.CheckAndPush("Guest52434", 2342, make(map[string]interface{}))
	lb.CheckAndPush("Guest00132", 2341, make(map[string]interface{}))

	time.Sleep(time.Second*1)

	fmt.Println(lb.GetPage(10, 0))


	lb.Print()
	//fmt.Println(timePassed)
}
