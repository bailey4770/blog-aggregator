# gator

Boot.dev project focussing on learning SQL.

## Installation

### 1. Install Postgres

- [Official docs](https://www.postgresql.org/download/)
- Example for Debian: `sudo apt install postgresql postgresql-contrib`

Create a database:

`CREATE DATABASE <database_name>`

Run the SQL files found in 'sql/schema/' on that database.

### 2. Install Go

- [Official docs](https://go.dev/doc/install)

### 3. Install blog-aggregator CLI

Run the below command:
`go install github.com/bailey4770/blog-aggregator@latest`

Ensure your GOBIN (or $GOPATH/bin) is on your PATH, then you can run:
`blog-aggregator help`

## Configuration

Run `blog-aggregator config set-db <server_address>`

Server address should be in the form `postgres://postgres:@localhost:5432/gator?sslmode=disable`

This is the server address set as default in the config file.

## Example Usage

```
blog-aggregator register john
blog-aggregator addfeed HackerNews https://news.ycombinator.com/rss
blog-aggregator browse 10
```
