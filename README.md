# Insider Message Sender

An application that sends messages retrieved from database periodically to a webhook url

## Running the application

Clone this repository

```bash
git clone https://github.com/Grustamli/insider-msg-sender.git --depth 1 && cd ./insider-msg-sender
```

Rename the included `.env.example` file to `.env`

```bash
cp .env.example .env
```

Run docker compose file

```bash
docker compose -f ./docker-compose.yml -f ./docker-compose.w-seeder.yml up -d
```

**Note**: Make sure the `WEBHOOK_URL` env variable is up to date webhook.site url with proper response settings as
described in the task file.

## Running integration tests

```bash
WEBHOOK_URL=<webhook url> go test ./tests/...
```

Make sure you add the webhook.site url with proper response settings

## Configuration

- `WEBHOOK_URL`: Required. Webhook URL to send the messages
- `DB_PASSWORD`: Required. Postgres DB Password
- `WEBHOOK_AUTH_HEADER`: Optional. Used when Webhook required auth with header. Must accompany WEBHOOK_AUTH_KEY.
- `WEBHOOK_AUTH_KEYl`: Optional. Used when Webhook required auth with header. Must accompany WEBHOOK_AUTH_HEADER.
- `WEBHOOK_CHARACTER_LIMIT`: Default limit is 160 characters
- `SEND_INTERVAL_SECONDS`: Number of seconds until the next send starts
- `MESSAGE_COUNT_PER_INTERVAL`: Number of messages to send each interval

## API endpoints

API runs on `http://localhost:8000`

Swagger API docs can be accessed at `http://localhost:8000/swagger/index.html`

- `POST /start` endpoint starts the message sender daemon
- `POST /stop` endpoint stops the message sender daemon
- `GET /messages` returns list of sent messages with `message_id` received from webhook and `sent_at` timestamp

## CLI

Single `seed` command is written to seed the database with given count `-c` per `-i` interval.
It is used in the docker-compose as `seeder` service to initialize and continuously seed the db with fake messages.
See the examples below

```bash

# seed 10 messages to database
/cli seed --db-url "postgres://postgres:password@localhost:5432/postgres" -c 10 

# seed 2 messages every 30 seconds. Database URL is read from environment variable $DATABASE_URL
DATABASE_URL="postgres://postgres:password@localhost:5432/postgres" /cli seed -c 2 -i 30

```

## Tech stack

- **Postgresql**: stores messages
- **Redis**: caches sent messages

## Used libraries

- `github.com/alecthomas/kong`: lightweight library to build CLIs
- `github.com/gin-gonic/gin`: REST API framework
- `github.com/pkg/errors`: Used primarily to wrap errors
- `github.com/redis/go-redis/v9`: Redis client for go
- `github.com/rs/zerolog`: Logger library
- `github.com/sethvargo/go-envconfig`: Automatic loading and parsing of config from environment

## Notes

- Used ChatGPT for most of the comments
