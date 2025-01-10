package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/bwmarrin/discordgo"
	"github.com/gorcon/rcon"
)

type Config struct {
	BotChannelID string
	GuildID      string
	Token        string
	RconAddr     string
	RconPass     string
}

// Global Variables
var config Config
var online = make(map[string]bool)

// Structures
type UserCacheEntry struct {
	Name      string `json:"name"`
	UUID      string `json:"uuid"`
	ExpiresOn string `json:"expiresOn"`
}

type StatsFile struct {
	Stats       map[string]map[string]int `json:"stats"`
	DataVersion int                       `json:"DataVersion"`
}

// ParseUserCache reads the usercache.json and returns a map of UUID to Username
func ParseUserCache(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []UserCacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, e := range entries {
		m[e.UUID] = e.Name
	}
	return m, nil
}

// ParseStats reads a player's stats JSON file and returns a StatsFile struct
func ParseStats(path string) (*StatsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sf StatsFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, err
	}
	return &sf, nil
}

// ParsePlayers processes the RCON 'list' command output
func ParsePlayers(s string) []string {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) < 2 {
		return nil
	}
	after := strings.TrimSpace(parts[1])
	if after == "" {
		return nil
	}
	sps := strings.Split(after, ",")
	var ps []string
	for _, p := range sps {
		ps = append(ps, strings.TrimSpace(p))
	}
	return ps
}

// changes detects joined and left players
func changes(s string) ([]string, []string) {
	ps := ParsePlayers(s)
	cur := make(map[string]bool)
	for _, p := range ps {
		cur[p] = true
	}
	var joined, left []string
	for p := range cur {
		if !online[p] {
			joined = append(joined, p)
		}
	}
	for p := range online {
		if !cur[p] {
			left = append(left, p)
		}
	}
	for _, p := range joined {
		online[p] = true
	}
	for _, p := range left {
		delete(online, p)
	}
	return joined, left
}

// Activity monitors player activity via RCON
func Activity(s *discordgo.Session) {
	conn, err := rcon.Dial(config.RconAddr, config.RconPass)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		resp, err := conn.Execute("list")
		if err != nil {
			log.Println(err)
			time.Sleep(30 * time.Second)
			continue
		}

		j, l := changes(resp)
		switch {
		case len(j) > 0 && len(l) > 0:
			s.ChannelMessageSend(config.BotChannelID, fmt.Sprintf("%s joined, %s left", strings.Join(j, ", "), strings.Join(l, ", ")))
		case len(j) > 0:
			s.ChannelMessageSend(config.BotChannelID, fmt.Sprintf("%s joined", strings.Join(j, ", ")))
		case len(l) > 0:
			s.ChannelMessageSend(config.BotChannelID, fmt.Sprintf("%s left", strings.Join(l, ", ")))
		}
		time.Sleep(2 * time.Second)
	}
}

// respond is a helper function to create a Discord interaction response
func respond(content string, extraContent ...map[string]int) *discordgo.InteractionResponse {
	var extra string
	if len(extraContent) > 0 && extraContent[0] != nil {
		for key, value := range extraContent[0] {
			extra += fmt.Sprintf("%s: %d\n", key, value)
		}
	}
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content + extra,
		},
	}
}

func main() {
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		log.Fatal("Failed to open config file", err)
	}

	// Create a new Discord session
	s, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	// Set intents
	s.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	// Open the Discord session
	err = s.Open()
	if err != nil {
		log.Fatal("Error opening Discord session:", err)
	}
	defer s.Close()

	// Get bot user ID
	botUser, err := s.User("@me")
	if err != nil {
		log.Fatal("Error fetching bot user:", err)
	}

	// Register /list command
	cmdList := &discordgo.ApplicationCommand{
		Name:        "list",
		Description: "Show online players",
	}
	_, err = s.ApplicationCommandCreate(botUser.ID, config.GuildID, cmdList)
	if err != nil {
		log.Fatal("Error registering /list command:", err)
	}

	// Register /stats command
	cmdStats := &discordgo.ApplicationCommand{
		Name:        "stats",
		Description: "Show a specific stat for a given player",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "stat",
				Description: "The stat to query (e.g. custom:jump), for help type 'help:key' or 'category:help'",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "player",
				Description: "The player's name",
				Required:    true,
			},
		},
	}
	_, err = s.ApplicationCommandCreate(botUser.ID, config.GuildID, cmdStats)
	if err != nil {
		log.Fatal("Error registering /stats command:", err)
	}

	// Handle interactions
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		switch i.ApplicationCommandData().Name {
		case "list":
			// Handle /list command
			conn, err := rcon.Dial(config.RconAddr, config.RconPass)
			if err != nil {
				s.InteractionRespond(i.Interaction, respond("Error connecting to RCON."))
				return
			}
			defer conn.Close()

			resp, err := conn.Execute("list")
			if err != nil {
				s.InteractionRespond(i.Interaction, respond("Error executing RCON command."))
				return
			}

			// Parse players and respond
			players := ParsePlayers(resp)
			if len(players) == 0 {
				s.InteractionRespond(i.Interaction, respond("No players are currently online."))
				return
			}
			s.InteractionRespond(i.Interaction, respond(fmt.Sprintf("Online Players: %s", strings.Join(players, ", "))))

		case "stats":
			// Handle /stats command
			options := i.ApplicationCommandData().Options
			if len(options) < 2 {
				s.InteractionRespond(i.Interaction, respond("Please provide both stat and player name."))
				return
			}

			statName := options[0].StringValue()
			playerName := options[1].StringValue()

			uuidMap, err := ParseUserCache("usercache.json")
			if err != nil {
				s.InteractionRespond(i.Interaction, respond("Error reading usercache."))
				return
			}

			var uuid string
			for k, v := range uuidMap {
				if strings.EqualFold(v, playerName) {
					uuid = k
					break
				}
			}

			if uuid == "" {
				s.InteractionRespond(i.Interaction, respond("Player not found in usercache."))
				return
			}

			// Parse stats
			statsPath := fmt.Sprintf("world/stats/%s.json", uuid)
			sf, err := ParseStats(statsPath)
			if err != nil {
				s.InteractionRespond(i.Interaction, respond("Error reading stats file."))
				return
			}

			parts := strings.SplitN(statName, ":", 2)
			if len(parts) != 2 {
				s.InteractionRespond(i.Interaction, respond("Stat must be in 'category:key' format, e.g., 'used:torch' or 'custom:jumps'."))
				return
			}

			if parts[0] == "help" {
				s.InteractionRespond(i.Interaction, respond("Options are: 'custom', 'broken', 'used', 'picked_up', 'killed', 'dropped', 'mined', 'crafted', 'killed_by'."))
				return
			}

			if parts[1] == "help" {
				categoryStats := sf.Stats["minecraft:"+parts[0]]
				s.InteractionRespond(i.Interaction, respond("Options for this category are:", categoryStats))
				return
			}

			category := "minecraft:" + parts[0]
			key := "minecraft:" + parts[1]

			categoryStats, categoryExists := sf.Stats[category]
			if !categoryExists {
				s.InteractionRespond(i.Interaction, respond(fmt.Sprintf("Category '%s' not found.", category)))
				return
			}

			val, keyExists := categoryStats[key]
			if !keyExists {
				s.InteractionRespond(i.Interaction, respond(fmt.Sprintf("Stat '%s' not found in category '%s'.", key, category)))
				return
			}

			response := fmt.Sprintf("%s's %s (%s): %d", playerName, key, category, val)
			s.InteractionRespond(i.Interaction, respond(response))
		}
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID != s.State.User.ID && strings.ToLower(m.Content) == "hi" {
			s.ChannelMessageSend(m.ChannelID, "chat")
		}
	})

	go Activity(s)

	fmt.Println("Bot is running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
