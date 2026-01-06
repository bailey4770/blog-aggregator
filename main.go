package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bailey4770/blog-aggregator/internal/config"
	"github.com/bailey4770/blog-aggregator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	if f, ok := c.handlers[cmd.name]; ok {
		err := f(s, cmd)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("command '%s' does not exist", cmd.name)
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) error {
	if _, ok := c.handlers[name]; ok {
		return errors.New("registering func that already exists")
	}

	c.handlers[name] = f
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("username not provided")
	} else if len(cmd.args) > 1 {
		return errors.New("too many args provided. need just one arg for username")
	}

	name := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if errors.Is(err, sql.ErrNoRows) {
		log.Fatal("Error: user not found")
	} else if err != nil {
		log.Fatal("Error: ", err)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}

	fmt.Println("username successfully set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %v <name>", cmd.name)
	}

	name := cmd.args[0]
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Constraint == "users_name_key" {
			return fmt.Errorf("user already exists")
		}
		return fmt.Errorf("couldn't create user: %w", err)
	}

	fmt.Println("User created successfully: ")
	printUser(user)

	err = handlerLogin(s, command{name: "login", args: []string{name}})
	if err != nil {
		return err
	}

	return nil
}

func handlerReset(s *state, _ command) error {
	err := s.db.DropAllUsers(context.Background())
	if err != nil {
		return err
	}

	err = config.ResetConfig()
	if err != nil {
		return err
	}

	return nil
}

func printUser(user database.User) {
	fmt.Printf("- ID:		%v\n", user.ID)
	fmt.Printf("- Name:		%v\n", user.Name)
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

	commandList := commands{make(map[string]func(*state, command) error)}
	err = commandList.register("login", handlerLogin)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	err = commandList.register("register", handlerRegister)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	err = commandList.register("reset", handlerReset)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	args := os.Args
	if len(args) < 2 {
		log.Fatal("Error: must provide arguments")
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
