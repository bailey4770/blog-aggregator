package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/bailey4770/blog-aggregator/internal/config"
)

type state struct {
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

	err := s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Println("username successfully set")
	return nil
}

func main() {
	err := config.Reset()
	if err != nil {
		log.Fatal("Error resetting config file:", err)
	}

	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	currentState := &state{cfg: &cfg}
	currentCommands := commands{make(map[string]func(*state, command) error)}
	err = currentCommands.register("login", handlerLogin)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	args := os.Args
	if len(args) < 2 {
		log.Fatal("Error: must provide arguments")
	}

	// arg[0] is program name, which can be safely ignored
	currentCmd := command{name: args[1], args: args[2:]}

	err = currentCommands.run(currentState, currentCmd)
	if err != nil {
		log.Fatal("Error running command: ", err)
	}

	newCfg, err := config.Read()
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	fmt.Println(newCfg)
}
