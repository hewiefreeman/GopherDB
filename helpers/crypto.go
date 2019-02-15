package helpers

import (
	"hash/fnv"
)

func HashString(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return Abs(int(h.Sum32()))
}

func Abs(i int) int {
	if i < 0 {
		return i*(-1)
	}
	return i
}
