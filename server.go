package main

////////////////// TO-DOs
//////////////////
//////////////////     - Password reset for AuthTable:
//////////////////         - Setting server name/address, subject & body message for password reset emails
//////////////////         - Send emails for password resets
//////////////////
//////////////////     - Database server
//////////////////         - Connection authentication
//////////////////         - Connection privillages
//////////////////
//////////////////     - Rate & connection limiting
//////////////////
//////////////////     - Clustering
//////////////////         - Connect to cluster nodes & agree upon master node
//////////////////         - Master assigns nodes key numbers and creates a keyspace unless valid ones have been created already
//////////////////         - Master ensures all nodes contain the same table schemas
//////////////////         -
//////////////////         - Global unique values
//////////////////
//////////////////     - Ordered tables

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"fmt"
	//"html"
	"net/http"
	"sync"
)

var (
	statusMux  sync.Mutex
	masterPass []byte
	dbStatus   int
	replica    bool
	readOnly   bool

	replicasMux sync.Mutex
	replicas    []string = []string{}

	balancersMux sync.Mutex
	balancers    []string = []string{}
)

const (
	configFile string = "db.conf"

	defaultConfigFile string = "{\"masterPass\":\"\",\"dbs\":[],\"replica\":false,\"readOnly\":false,\"replicas\":[],\"routers\":[],\"AuthTables\":[],\"Leaderboards\":[]}"
)

// Database statuses
const (
	statusSettingUp = iota
	statusHealthy
	statusRecovering
	statusReplicationFailure
	statusOffline
)

func main() {
	// initialize and start database server
	/*http.HandleFunc("/", queryHandler)
	fmt.Println("starting server...")
	if err := http.ListenAndServe("localhost:8082", nil); err != nil {
		fmt.Println(err)
	}*/
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	//
}
