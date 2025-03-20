# Noteo

Noteo is a notification service that allows users to subscribe to projects and 
receive notifications through a Telegram bot. The service includes subscription 
management features such as muting, pausing, and unsubscribing from 
notifications.

## Rationale

I needed a simple way to receive notifications anywhere when I am online: on 
the mobile or on desktop. The communication app of my choice is Telegram, so I 
decided to build a bot that would allow me to do just that.

I chose Go as the programming language because it is fast, has a small 
footprint, and can be run anywhere independently of the operating system and 
libraries.

Also I wanted to use LLM-powered IDEs (Cursor and later Windsurf) for the 
entirety of the development process.

## Features

- Telegram bot for user interaction
- Project subscription management
- Notification controls (mute, unmute, pause, resume)
- API service for sending notifications

## Prerequisites

- Go 1.23 or later for building the application
- `gcc` for building with CGO
- Telegram Bot Token (obtained from [BotFather](https://t.me/botfather)) for 
  running it

## Environment Variables

The application uses the following environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NOTEO_BOT_TOKEN` | Telegram Bot API token | - | Yes |
| `NOTEO_PORT` | Port for the API server | 8080 | No |
| `NOTEO_LOG_FORMAT` | Log format (json or text) | json | No |
| `NOTEO_LOG_LEVEL` | Log level (debug, info, warn, error) | info | No |
| `NOTEO_DB_DSN` | SQLite database connection string | - | Yes |

## Developing and running locally

The application is self-contained, so you can just `go run` it, or use your
favourite IDE. You will need to provide a bot token and path to database file
via environment variables. You may use `:memory:` database for tests and debug.

## Deployment to production

### Docker

The default way to run production apps today is using containers. The
`Dockerfile` in the project root is self-explainatory. `tzdata` and
certificates aren't actually required, I've added them just for the
sake of completeness.

### Railway

Railway will automatically detect a `Dockerfile` and deploy it. You will need
to add a persistent data volume to your service and pass the mount path
via environment variable. For example, if your mount path is `/data`, you
should set `NOTEO_DB_DSN` variable to `/data/noteo.sqlite`. Also add
`NOTEO_PORT={{PORT}}` variable to correctly forward the incoming http traffic.

When you fire `railway up`, open Railway console and agree to automatically
assign a domain to access your app's API.

If you want to use Railway's Go builder, remove or rename the `Dockerfile` and
add a `nixpacks.toml` file to the project root:

```toml
[variables]
CGO_ENABLED = "1"
# NIXPACKS_GO_VERSION = "1.23.4"

[phases.setup]
nixPkgs = ['...', 'gcc']
```

Under light load, running the app will cost about $0.01 per day as of Mar 2025.

### Naked binary

You may also build a binary executable and run it directly:

```bash
CGO_ENABLED=1 go build -o noteo -ldflags="-w -s" .
NOTEO_BOT_TOKEN=... NOTEO_DB_DSN=./noteo.sqlite ./noteo
```

## License

This project is licensed under the MIT License - see the LICENSE.md file for 
details.
