package utils

import (
	"fmt"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	b := InitBF(1000)
	b.SetItem("hello")
	b.SetItem("yes")
	fmt.Println(b.GetItem("hello"))
	fmt.Println(b.GetItem("yes"))
	fmt.Println(b.GetItem("daw"))
}
