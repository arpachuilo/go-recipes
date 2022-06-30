# go-recipes

Website for scraping/editing/creating/reading recipes.

Created this for personal use on my local network.

## Requirements

- go 1.18+
- python 3.0+
- pip

## Setup

Steps to setting up the program.

### Configuration File

Configuration is done via [https://github.com/spf13/viper](https://github.com/spf13/viper). Can use any filetype viper supports. I'm liking toml.

Example Config

```toml
[server]
address = ":80"
read_timeout = "10s"
write_timeout = "15s"

  [server.rate_limiter]
  limit = 30
  burst = 50
  timeout = "3m"

[database]
path = "./recipes.db"
```

### Database

Currently only supports sqlite3. Does make use of [sqlboiler](https://github.com/volatiletech/sqlboiler#sqlboiler) so could be possible enough to easily extend to supporting other databases. But why bother for a personal project.

The quickest way to setup the database would be to checkout the scraping program in `cmds/scrape`. If you would rather not, just create your own using the same commands used in `scrape.py`.
