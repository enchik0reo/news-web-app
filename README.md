# Go Newsline

It's a full-stack web application called Go Newsline that lets people get articles about Go in one place.
Backend powered by Go. Frontend powered by React.

## Scheme

![Scheme](./scheme.png)

## Features

- Application backend contains three separate services (Auth, News, Api)
- Backend services communicate with each other using gRPC
- Auth service for user authentication
- User can signup and login. Using JWT Access and Refresh tokens
- Logging users can suggest news
- News service uses rss feed and website parsing to save and view articles
- Web api server on net/http (use chi router)
- RESTful routing
- Data persistence using PostgreSQL
- Cache using Redis
- Using migrations for database
- Frontend using React
- App uses Prometheus to collect metrics and Grafana to show them

## Development

Software requirements:

- Docker

To start the application use three commands:

```sh
$ git clone https://github.com/enchik0reo/newsWebApp

$ cd newsWebApp

$ docker-compose up --build
```
- Go to http://localhost:3003/ and try app (a new article will be published every 15 minutes).
- Go to http://localhost:9090/ to see Prometheus.
- Go to http://localhost:3000/ to see Grafana.

To terminate services, the application uses `SIGTERM` signal (use Ctrl+C).