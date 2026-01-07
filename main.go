package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/bailey4770/blog-aggregator/internal/config"
	"github.com/bailey4770/blog-aggregator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Error reading config file: ", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal("Error opening SQL database: ", err)
	}

	dbQueries := database.New(db)
	currentState := &state{db: dbQueries, cfg: &cfg}

	commandList, err := getCommands()
	if err != nil {
		log.Fatal("Error: ", err)
	}

	args := os.Args
	if len(args) < 2 {
		log.Fatal("Error: must provide command")
	}

	// arg[0] is program name, which can be safely ignored
	cmd := command{name: args[1], args: args[2:]}

	err = commandList.run(currentState, cmd)
	if err != nil {
		log.Fatal("Error running command: ", err)
	}

	newCfg, err := config.Read()
	if err != nil {
		log.Fatal("Error reading config file: ", err)
	}

	fmt.Println(newCfg)
}
