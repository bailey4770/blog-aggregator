package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
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

	err := commandList.new("version", handlerVersion)
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("help", handlerHelp)
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("login", handlerLogin)
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

	err = commandList.new("addfeed", middlewareLoggedIn(handlerAddFeed))
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("feeds", middlewareLoggedIn(handlerFeeds))
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("follow", middlewareLoggedIn(handlerFollow))
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("following", handlerFollowing)
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("unfollow", middlewareLoggedIn(handlerUnfollow))
	if err != nil {
		return commands{}, err
	}

	err = commandList.new("browse", middlewareLoggedIn(handlerBrowse))
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

func handlerVersion(_ *state, _ command) error {
	fmt.Println("Version:", version)
	return nil
}

func handlerHelp(_ *state, _ command) error {
	commandList, err := getCommands()
	if err != nil {
		return fmt.Errorf("could not get command list: %v", err)
	}

	fmt.Println("Available commands:")
	for name := range commandList.handlers {
		fmt.Println("-", name)
	}

	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("usage: %v <name>", cmd.name)
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

	err = config.Reset()
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

func scrapeFeeds(s *state) error {
	feedDB, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("could not get oldest updated feed from db: %v", err)
	}

	fmt.Printf("Scraping posts from %s %s\n", feedDB.Name, feedDB.Url)
	feed, err := s.client.FetchFeed(context.Background(), feedDB.Url)
	if err != nil {
		return fmt.Errorf("could not fetch feed: %v", err)
	}

	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		UpdatedAt:     time.Now(),
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:            feedDB.ID,
	})
	if err != nil {
		return fmt.Errorf("could not mark feed as fetched: %v", err)
	}

	feed.RemoveHTMLUnescape()
	for _, item := range feed.Channel.Items {
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return fmt.Errorf("could not parse pub date %v", err)
		}

		err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       sql.NullString{String: item.Title, Valid: true},
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: pubDate,
			FeedID:      feedDB.ID,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				// Duplicate key error. Safe to ignore.
				continue
			} else {
				return fmt.Errorf("could not create post in DB %v", err)
			}
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %v <request_frequency>", cmd.name)
	}

	requestFrequency, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("could not parse time duration: %v", err)
	}

	fmt.Printf("collecting feeds every %v\n", requestFrequency)

	ticker := time.NewTicker(requestFrequency)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		currentUser, err := s.db.GetUser(context.Background(), s.cfg.CurrentUsername)
		if err != nil {
			return err
		}
		return handler(s, cmd, currentUser)
	}
}

func handlerAddFeed(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: %v <name> <url>", cmd.name)
	}

	name := cmd.args[0]
	url := cmd.args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    currentUser.ID,
	})
	if err != nil {
		return fmt.Errorf("could not create feed: %v", err)
	}

	fmt.Println("Successfully created feed:", feed.Name, feed.Url)

	followCmd := command{name: "follow", args: []string{feed.Url}}
	err = handlerFollow(s, followCmd, currentUser)
	if err != nil {
		return fmt.Errorf("could not follow feed: %v", err)
	}

	return nil
}

func handlerFeeds(s *state, cmd command, currentUser database.User) error {
	feeds, err := s.db.GetFeedList(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Println(feed)
	}

	return nil
}

func handlerFollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.name)
	}
	url := cmd.args[0]

	feedRecord, err := s.db.GetFeedRecord(context.Background(), url)
	if err != nil {
		return fmt.Errorf("could not find feed in feeds table: %v", err)
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feedRecord.ID,
	})
	if err != nil {
		return fmt.Errorf("could not create feed_follow record: %v", err)
	}

	fmt.Printf("%s successfully followed %s\n", currentUser.Name, feedRecord.Name)
	return nil
}

func handlerFollowing(s *state, cmd command) error {
	currentUser := s.cfg.CurrentUsername

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser)
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Printf("%s is not following any feeds\n", currentUser)
	}

	for _, feed := range feeds {
		fmt.Printf("- %s\n", feed.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.name)
	}
	feedURL := cmd.args[0]

	feed, err := s.db.GetFeedRecord(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("could not find feed in feeds table: %v", err)
	}

	_, err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: currentUser.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("could not unfollow feed: %v", err)
	}

	fmt.Printf("%s successfully unfollowed feed %s\n", currentUser.Name, feedURL)
	return nil
}

func handlerBrowse(s *state, cmd command, currentUser database.User) error {
	limit := 2
	if len(cmd.args) == 1 {
		var err error
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("could not parse limit arg: %v", err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: currentUser.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("could not get posts for current user: %v", err)
	}

	for _, posts := range posts {
		title := "null"
		if posts.Title.Valid {
			title = posts.Title.String
		}

		publishDate := posts.PublishedAt.Format("2006-01-02")

		fmt.Printf("- %v '%v' from %s\n", publishDate, title, posts.FeedName)
		fmt.Printf("    %v\n", posts.Url)
	}

	return nil
}
