package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var (
	badLinks []string
)

func prettyPrint(str interface{}) {
	strJson, _ := json.MarshalIndent(str, "", "  ")
	log.Printf("%s\n", string(strJson))
}

func main() {

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("Fatal error config file:", err)
	}

	discord, err := discordgo.New("Bot " + viper.GetString("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Error creating Discord session: %s\n", err)
	}
	badLinks = viper.GetStringSlice("BAD_LINKS")

	discord.AddHandler(messageCreate)
	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildBans

	err = discord.Open()
	if err != nil {
		log.Fatalln("Cant connect using DISCORD_TOKEN", viper.GetString("DISCORD_TOKEN"))
		os.Exit(1)
	}

	// Join servers
	log.Println("Invite link: ", fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s&scope=bot&permissions=%d", discord.State.User.ID, discord.Identify.Intents))

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	// if m.Message is in badLinks
	for _, link := range badLinks {
		if strings.Contains(m.Content, link) {
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			err := s.GuildBanCreate(m.GuildID, m.Author.ID, 2)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintln("Tried banning", m.Author.Username, "for", link, "but", err))
			}

			s.ChannelMessageSend(viper.GetString("LOG_CHANNEL"), fmt.Sprintln("Banning", m.Author.Username, "for", link))

			return
		}
	}
}
