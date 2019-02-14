package main

import (
 "sync"
)

var (
	configFile string = "db.conf"

	statusMux sync.Mutex
	dbStatus  int
	replica   bool
	readOnly  bool

	replicasMux sync.Mutex
	replicas    []string = []string{}

	balancersMux sync.Mutex
	balancers    []string = []string{}
)

func main() {
	// initialize and start database server
}
