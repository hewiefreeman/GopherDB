package helpers

import (
	"testing"
	"github.com/hewiefreeman/GopherDB/helpers"
)

func TestInitLogger(t *testing.T) {
	if err := helpers.InitLogger(3, 100); err != nil {
		t.Errorf("Error initializing logger: %v", err)
	}
}

func TestLogger(t *testing.T) {
	helpers.Log("This should not show up", 1)
	helpers.Log("This should not show up either", 2)
	helpers.Log("This should be the first logger entry", 3)
	helpers.Log("This should be the second logger entry", 5)
}

func TestCloseLogger(t *testing.T) {
	helpers.CloseLogger()
}