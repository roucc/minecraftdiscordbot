# Minecraft Discord Bot

A Discord bot written in Go that connects to a Minecraft server via RCON to monitor player activity, announce player joins/leaves, and provide player statistics through slash commands.

## Features

- **Player Activity Monitoring:** Automatically announces when players join or leave the Minecraft server.
- **Slash Commands:**
  - `/list`: Displays the list of currently online players.
  - `/stats get <stat_name> <player_name>`: Retrieves a specific statistic for a given player.
- **Interactive Responses:** Responds to simple messages like "hi" with a greeting.
- **Configurable:** Uses a `config.toml` file to manage configuration settings.

## Prerequisites

- **Go:** Ensure you have Go installed. You can download it from [the official website](https://golang.org/dl/).
- **Discord Bot Token:** You need a Discord bot token. Follow [Discord's guide](https://discord.com/developers/docs/intro) to create a bot and obtain its token.
- **Minecraft Server with RCON Enabled:** Ensure your Minecraft server is running and RCON is enabled with the correct address and password.
- **Config File:** `config.toml` file with necessary configuration settings.

## Example config.toml

```
BotChannelID = "xxxxxxxxxxxxxx"
GuildID      = "xxxxxxxxxxxxxxxxx"
Token        = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
RconAddr     = "localhost:25575"
RconPass     = "xxxxxxxxxx"
```
