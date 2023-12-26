# Go Newsline

It's a full-stack Go web application called Newsline that lets people get articles about Go in one place.

## Features

- Application contains three separate services (Auth, News, Web)
- Communicate with each other using gRPC
- Auth service for user authentication
- User can signup and login. Using JWT Access and Refresh tokens
- News service uses rss feed and website parsing
- Save and view articles
- RESTful routing
- Web server on net/http (use chi router)
- Data persistence using PostgreSQL database
- Using migrations
- Cache using Redis database
- Dynamic HTML using Go templates

## Development

Software requirements:

- This project supports Go modules.
- Docker
- task

To start the application:

```sh
$ git clone https://github.com/enchik0reo/newsWebApp
$ cd newsWebApp

# Run docker-compose
$ task dup

# Run migrator action up, or (task mdown) if you need
$ task mup

# Run docker-compose
$ task dup

# Run wauth service on port 44044
$ task run_auth

# Run news service on port 55055
$ task run_news

# Run web server on port 8000
$ task run_web
```

To finish the services use `sigterm` signal.