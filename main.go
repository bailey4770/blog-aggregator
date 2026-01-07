package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/bailey4770/blog-aggregator/internal/config"
	"github.com/bailey4770/blog-aggregator/internal/database"
	"github.com/bailey4770/blog-aggregator/internal/rssclient"
)

type state struct {
	client rssclient.Client
	db     *database.Queries
	cfg    *config.Config
}

func main() {
	client := rssclient.NewClient(5 * time.Second)

	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Error reading config file: ", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal("Error opening SQL database: ", err)
	}

	dbQueries := database.New(db)
	currentState := &state{client: client, db: dbQueries, cfg: cfg}

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
}
