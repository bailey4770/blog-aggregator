package config

import (
	"fmt"
	"testing"
)

func TestRead(t *testing.T) {
	res, err := Read()
	if err != nil {
		t.Errorf("func returned err: %v", err)
	}

	fmt.Println(res)
}
