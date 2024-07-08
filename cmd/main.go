package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token     string
	UserID    string
	Separator string
)
var sc = make(chan os.Signal, 1)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&UserID, "u", "", "User's discord ID")
	flag.StringVar(&Separator, "s", ":::", "Separator between message and emoji")
	flag.Parse()
}

var unresolvedMsgs = make(map[string]string)

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	dg.AddHandler(reactionAdd)

	// Just like the ping pong example, we only care about receiving message
	// events in this example.
	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsDirectMessageReactions

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer dg.Close()

	channel, err := dg.UserChannelCreate(UserID)
	if err != nil {
		fmt.Println("error creating DM channel:", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		m, err := dg.ChannelMessageSend(channel.ID, line)
		if err != nil {
			os.Stderr.WriteString("Error sending a DM: " + err.Error())
		}

		unresolvedMsgs[m.ID] = line
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// We create the private channel with the user who sent the message.
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		os.Stderr.WriteString("error creating channel: " + err.Error())
		return
	}
	// Then we send the message through the channel we created.
	_, err = s.ChannelMessageSend(channel.ID, "Pong!")
	if err != nil {
		os.Stderr.WriteString("error sending DM message: " + err.Error())
	}
}

func reactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	msg, ok := unresolvedMsgs[r.MessageID]
	if !ok {
		fmt.Fprintf(os.Stderr, "Message %s not found in unresolved messages", r.MessageID)
		return
	}

	fmt.Fprintf(os.Stdout, "%s%s%s", msg, Separator, r.Emoji.Name)

	delete(unresolvedMsgs, r.MessageID)
	if len(unresolvedMsgs) == 0 {
		sc <- syscall.SIGTERM
	}
}
