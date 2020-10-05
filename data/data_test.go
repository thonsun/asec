package data

import (
	"fmt"
	"testing"
)

func TestSHA256Hash(t *testing.T) {
	fmt.Println(SHA256Hash("J@nusec123"+"afa8bae009c9dbf4135f62e165847227"))
	fmt.Println("1f7d7e9decee9561f457bbc64dd76173ea3e1c6f13f0f55dc1bc4e99e5b8b494")
}
