package main

import (
	"github.com/hewiefreeman/GopherDB/storage"
)

func main() {
	storage.Insert("UT-users/0.gdbs", []byte{52, 53, 54, 55, 56, 57})
	storage.Insert("UT-users/0.gdbs", []byte{52, 53, 54, 55, 56, 57})
	storage.Insert("UT-users/0.gdbs", []byte{52, 53, 54, 55, 56, 57})
	storage.Insert("UT-users/0.gdbs", []byte{52, 53, 54, 55, 56, 57})
	storage.Insert("UT-users/0.gdbs", []byte{52, 53, 54, 55, 56, 57})
	storage.Update("UT-users/0.gdbs", 1, []byte{65})
	storage.Update("UT-users/0.gdbs", 5, []byte{65, 66, 67, 68, 69, 70})
}