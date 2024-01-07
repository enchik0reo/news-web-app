# Go Newsline

It's a full-stack web application called Go Newsline that lets people get articles about Go in one place.
Backend powered by Go. Frontend powered by React.

## Scheme

![Scheme](./scheme.png)

## Features

- Application backend contains three separate services (Auth, News, Api)
- Backend communicate with each other using gRPC
- Auth service for user authentication
- User can signup and login. Using JWT Access and Refresh tokens
- Logging users can suggest news
- News service uses rss feed and website parsing to save and view articles
- Web api server on net/http (use chi router)
- RESTful routing
- Data persistence using PostgreSQL
- Cache using Redis
- Using migrations
- Frontend using React

## Development

Software requirements:

- Go
- Docker
- React
- task

To start the application:

```sh
$ git clone https://github.com/enchik0reo/newsWebApp
$ cd newsWebApp

# Run docker-compose and migrator with action up
# Or use 'task down' if you need to rollback the migration
$ task up

# Run auth service on port 44044
$ task run_auth

# Run news service on port 55055
$ task run_news

# Run api service on port 8008
$ task run_api

# Run frontend react app on port 3003
$ task run_front
```
Go to http://localhost:3003/ and try it.

To terminate services, the application uses `SIGTERM` signal (you can use Ctrl+C).