package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bailey4770/blog-aggregator/internal/config"
	"github.com/bailey4770/blog-aggregator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

// add all new commands to below list
// create new handlerFunc() for command
func getCommands() (commands, error) {
	commandList := commands{make(map[string]func(*state, command) error)}

	err := commandList.new("login", handlerLogin)
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("register", handlerRegister)
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("reset", handlerReset)
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("users", handlerUsers)
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("agg", handlerAgg)
	if err != nil {
		return commands{}, err
	}

	return commandList, nil
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

func (c *commands) new(name string, f func(*state, command) error) error {
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
	_, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
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

	fmt.Println("User created successfully: ", name)

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

func printUsers(users []database.User, currentUser string) {
	for _, user := range users {
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUserList(context.Background())
	if err != nil {
		return err
	}

	currentUser := s.cfg.CurrentUsername
	printUsers(users, currentUser)

	return nil
}

func handlerAgg(s *state, cmd command) error {
	var feedURL string
	if len(cmd.args) != 1 {
		feedURL = "https://www.wagslane.dev/index.xml"
	} else {
		feedURL = cmd.args[0]
	}

	RSSFeed, err := s.client.FetchFeed(context.Background(), feedURL)
	if err != nil {
		return err
	}

	RSSFeed.RemoveHTMLUnescape()

	fmt.Println("Title: ", RSSFeed.Channel.Title)
	fmt.Println("Link: ", RSSFeed.Channel.Link)
	fmt.Println("Description: ", RSSFeed.Channel.Description)

	for i, item := range RSSFeed.Channel.Items {
		fmt.Printf("Item %d: %s\n", i, item)
	}

	return nil
}
