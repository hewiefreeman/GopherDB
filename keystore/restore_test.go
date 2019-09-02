package keystore

import (
	"testing"
	"github.com/hewiefreeman/GopherDB/keystore"
)

var (
	k *keystore.Keystore
)

func TestRestore(t *testing.T) {
	var err int
	k, err = keystore.Restore("test")
	if err != 0 {
		t.Errorf("Error #%v", err)
	}
}

func TestGetAfterRestore(t *testing.T) {
	if k == nil {
		t.Errorf("Table wasn't made.")
		return
	}
	data, iErr := k.GetKeyData("test", []string{"subbed.*since.*sec"})
	if iErr != 0 {
		t.Errorf("Get error: %v", iErr)
		return
	}
	t.Logf("Got: %v", data)
	k.Close(true)
}