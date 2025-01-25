# Minecraft Discord RCON Bot

Short guide:

## Features
- Monitors Minecraft players joining/leaving.
- Multiple slash commands (`/list`, `/stats`, `/whitelist`, `/difficulty`, `/setcustommodeldata`).
- RCON integration to execute commands.

## Requirements
- Go 1.18+
- Running Minecraft server with RCON enabled.
- Discord bot token.

## Installation
1. **Obtain** your Discord token and server's RCON details.
2. **Edit** `config.toml`.
3. **Build** and run the bot: `go build && ./yourbot`.

## Config (config.toml)
```toml
BotChannelID = "DISCORD_CHANNEL_ID"
GuildID = "YOUR_GUILD_ID"
Token = "DISCORD_BOT_TOKEN"
RconAddr = "127.0.0.1:25575"
RconPass = "YOUR_RCON_PASSWORD"
```

## Usage
- `/list`: Show online players.
- `/stats`: Show a player's stat. Usage: `/stats <stat> <player>`.
- `/whitelist`: Add a player to the whitelist.
- `/difficulty`: Change difficulty (disallows "peaceful").
- `/setcustommodeldata`: Advanced usage for custom item models.

## Notes
- RCON config must match your server's `server.properties`.
- `usercache.json` and `world/stats` must be accessible.
