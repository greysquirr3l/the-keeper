# Keeper Bot

Keeper Bot is a modular Discord bot built in Go, designed to provide various features like managing player IDs, redeeming gift codes, scraping gift code websites, and more. It uses SQLite for the backend database, GORM for ORM, Viper for configuration management, and Logrus for logging.

## Features

	•	**Modular Command System**: Easily extend functionality by editing `commands.yaml` and adding corresponding handler files.
	•	**Gift Code Management**: Redeem, validate, and list gift codes, with the ability to deploy codes to all users.
	•	**Web Scraping**: Scrapes websites for gift codes using configurable selectors.
	•	**Player ID Management**: Add, edit, remove, and list player IDs.
	•	**Database-Backed**: Uses SQLite for data persistence with GORM as the ORM.
	•	**Configuration**: Centralized configuration using Viper, with YAML support.
	•	**Logging**: Flexible logging using Logrus, with configurable log levels.

## Getting Started

### Prerequisites

	•	[Go](https://golang.org/doc/install) (1.18 or higher)
	•	[SQLite](https://www.sqlite.org/index.html) (pre-installed on most systems)
	•	[Railway](https://railway.app/) or any Docker-based deployment platform

### Installation

	1.	**Clone the repository:**
```bash
git clone https://github.com/yourusername/keeper-bot.git
cd keeper-bot
```
	2.	**Install dependencies:**
Inside the project directory, install the Go dependencies:
```bash
go mod tidy
```
	3.	**Setup the Configuration:**
Rename `config.template.yaml` to `config.yaml` and adjust the settings for your environment:
```bash
cp config.template.yaml config.yaml
```
Ensure you update:
	•	Discord tokens and client details (`DISCORD_BOT_TOKEN`, `DISCORD_CLIENT_ID`, etc.)
	•	Database file location (`/app/data2/keeper.db`)
	•	Scraping configuration (selectors and URLs)
	4.	**Run Migrations:**
Ensure your SQLite database is set up with the necessary schema by running any migrations:
```bash
go run migrations.go
```
	5.	**Run the Bot:**
Start the bot locally:
```bash
go run main.go
```

### Docker Setup

	1.	**Build the Docker image:**
```bash
docker build -t keeper-bot .
```
	2.	**Run the container:**
```bash
docker run -d –name keeper-bot -v $(pwd)/data:/app/data2 keeper-bot
```

### Railway Deployment

	1.	**Install the Railway CLI:**
Follow the instructions at [Railway.app](https://railway.app/) to install the CLI.
	2.	**Link your project:**
Inside the project folder:
```bash
railway link
```
	3.	**Deploy your project:**
```bash
railway up
```

### Configuration

The bot’s behavior is controlled by the `config.yaml` file. Key configurations include:

```yaml
discord:
token: ${DISCORD_BOT_TOKEN}
client_id: ${DISCORD_CLIENT_ID}
command_prefix: “!”
RoleID: ${DISCORD_ROLE_ID}

database:
volumeMountPath: “/app/data2”
name: “keeper.db”

logging:
log_level: “info”

scrape:
sites:
- name: “VG247”
url: “https://www.vg247.com/whiteout-survival-codes”
selector: “ul li strong”
```

	•	**Discord**: Settings related to your bot’s Discord integration.
	•	**Database**: File path and name for the SQLite database.
	•	**Logging**: Log level (`info`, `debug`, `warn`, etc.).
	•	**Scraping**: Configure websites and selectors for scraping gift codes.

### Commands

The bot’s commands are defined in `commands.yaml`. You can add new commands or modify existing ones by editing this file and creating the corresponding handler.

Example of `commands.yaml`:

```yaml
commands:
id:
description: “Manage player IDs”
usage: “!id  [arguments]”
handler: “handleIDCommand”

giftcode:
description: “Manage gift codes”
usage: “!giftcode  [arguments]”
handler: “handleGiftCodeCommand”
```

### Contributing

	1.	Fork the repository
	2.	Create your feature branch (`git checkout -b feature/your-feature`)
	3.	Commit your changes (`git commit -m ‘Add some feature’`)
	4.	Push to the branch (`git push origin feature/your-feature`)
	5.	Open a Pull Request

### License

Distributed under the MIT License. See `LICENSE` for more information.

### Contact

Project Link: [https://github.com/yourusername/keeper-bot](https://github.com/yourusername/keeper-bot)
