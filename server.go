package main

import (
	"sync"
	"net/http"
	"html"
	"fmt"
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

	defaultConfigFile string = "{\"masterPass\":"",\"dbs\":[],\"replica\":false,\"readOnly\":false,\"replicas\":[],\"routers\":[],\"AuthTables\":[],\"Leaderboards\":[]}"
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
	http.HandleFunc("/", queryHandler)
	fmt.Println("starting server...")
	if err := http.ListenAndServe("localhost:8082", nil); err != nil {
		fmt.Println(err)
	}
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}