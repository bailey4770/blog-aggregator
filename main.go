package main

import (
	"fmt"

	"github.com/bailey4770/blog-aggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config file:", err)
	}

	currentUser := "bailey4770"
	err = cfg.SetUser(currentUser)
	if err != nil {
		fmt.Println("Error setting username:", err)
	}

	newCfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config file:", err)
	}

	fmt.Println(newCfg)

	err = config.Clear()
	if err != nil {
		fmt.Println("Error reseting config file:", err)
	}
}
