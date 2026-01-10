# blog-aggregator

Boot.dev guided project focussing on learning SQL.

## Installation

### 1. Install Postgres

- [Official docs](https://www.postgresql.org/download/)
- Example for Linux (Debian): `sudo apt install postgresql postgresql-contrib`

### 2. Create a database

- Enter the psql shell with `sudo -u postgres psql` (for Linux)
- Run this command: `CREATE DATABASE <database_name>`
- Run the SQL files found in 'sql/schema/' on that database.

### 3. Install Go

- [Official docs](https://go.dev/doc/install)

### 4. Install blog-aggregator CLI

- Run this command:
`go install github.com/bailey4770/blog-aggregator@latest`

- Ensure your GOBIN (or $GOPATH/bin) is on your PATH, then you can run:
`blog-aggregator help`

## Configuration

- Run `blog-aggregator config set-db <server_address>`

- Server address should be in the form `postgres://postgres:@localhost:5432/gator?sslmode=disable`

- This is the server address set as default in the config file.

## Example Usage

- Run the below commands:

```
blog-aggregator register john
blog-aggregator addfeed HackerNews https://news.ycombinator.com/rss
```

- In one terminal, run the below command:

```
blog-aggregator agg 5s
```

which should show below output:

```
collecting feeds every 5s
Scraping posts from HackerNews https://news.ycombinator.com/rss
Scraping posts from HackerNews https://news.ycombinator.com/rss
Scraping posts from HackerNews https://news.ycombinator.com/rss
Scraping posts from HackerNews https://news.ycombinator.com/rss
```

- You can leave this running, or close the program with `Ctrl+C` after all the feeds have been scraped
- Run the below command to see the 10 most-recent posts from feeds the current user is following:

```
blog-aggregator browse 10
```

which should show similar to output below:

```
- 2026-01-10 'UpCodes (YC S17) is hiring PMs, SWEs to automate construction compliance' from HackerNews
    https://up.codes/careers?utm_source=HN
- 2026-01-10 'AI is a business model stress test' from HackerNews
    https://dri.es/ai-is-a-business-model-stress-test
- 2026-01-10 'Drones that recharge directly on transmission lines' from HackerNews
    https://www.ycombinator.com/companies/voltair
- 2026-01-10 'Microsoft May Have Created the Slowest Windows in 25 Years with Windows 11' from HackerNews
    https://www.eteknix.com/microsoft-may-have-created-the-slowest-windows-in-25-years-with-windows-11/
- 2026-01-10 'Open Chaos: A self-evolving open-source project' from HackerNews
    https://www.openchaos.dev/
- 2026-01-10 'I replaced Windows with Linux and everything's going great' from HackerNews
    https://www.theverge.com/tech/858910/linux-diary-gaming-desktop
- 2026-01-10 'UK government exempting itself from cyber law inspires little confidence' from HackerNews
    https://www.theregister.com/2026/01/10/csr_bill_analysis/
- 2026-01-10 'Show HN: Yuanzai World – LLM RPGs with branching world-lines' from HackerNews
    https://www.yuanzai.world/
- 2026-01-10 'A Eulogy for Dark Sky, a Data Visualization Masterpiece (2023)' from HackerNews
    https://nightingaledvs.com/dark-sky-weather-data-viz/
- 2026-01-10 'New information extracted from Snowden PDFs through metadata version analysis' from HackerNews
    https://libroot.org/posts/going-through-snowden-documents-part-4/
```

The number can be omitted. The default is 2.

- If your terminal supports, click on the link to view the post.
- Alternatively, copy and paste the link into your web browser.
