# Insider Message Sender

An application that sends messages retrieved from database periodically to a webhook url

## Running the application

Clone this repository

```bash
git clone https://github.com/Grustamli/insider-msg-sender.git --depth 1
```

Rename the included `.env.example` file to `.env`

```bash
mv .env.example .env
```

Run docker compose file

```bash
docker compose up -d
```

## API endpoints

API runs on `http://localhost:8000`

Swagger API docs can be accessed at `http://localhost:8000/swagger/index.html`

- `POST /start` endpoint starts the message sender daemon
- `POST /stop` endpoint stops the message sender daemon
- `GET /messages` returns list of sent messages with `message_id` received from webhook and `sent_at` timestamp

## CLI

Single `seed` command is written to seed the database with given count `-c` per `-i` interval.
It is used in the docker-compose as `seeder` service to initialize and continuously seed the db with fake messages
See the examples below

```bash

# seed 10 messages to database
/cli seed --db-url "postgres://postgres:password@localhost:5432/postgres" -c 10 

# seed 2 messages every 30 seconds. Database URL is read from environment variable $DATABASE_URL
DATABASE_URL="postgres://postgres:password@localhost:5432/postgres" /cli seed -c 2 -i 30

```

## Tech stack

- **Postgresql**: stores messsages
- **Redis**: caches sent messages

## Used libraries

- `github.com/alecthomas/kong`: lightweight library to build CLIs
- `github.com/gin-gonic/gin`: REST API framework
- `github.com/pkg/errors`: Used primarily to wrapp errors
- `github.com/redis/go-redis/v9`: Redis client for go
- `github.com/rs/zerolog`: Logger library
- `github.com/sethvargo/go-envconfig`: Automatic logging and parsing of config from environment

## Notes

- Used ChatGPT for most of the comments