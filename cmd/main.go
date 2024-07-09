package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/erykksc/notifr/internal/utils"
)

const WatchEmoji = "👀"

// Variables used for command line parameters
var (
	Token        string
	UserID       string
	Separator    string
	MsgSeparator string
	OutSeparator string
	Verbose      bool
)
var sc = make(chan os.Signal, 1)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&UserID, "u", "", "User's discord ID")
	flag.StringVar(&Separator, "s", ":", "Separator between message and emoji")
	flag.StringVar(&MsgSeparator, "msg-sep", "\n", "Separator between messages")
	flag.StringVar(&OutSeparator, "out-sep", "\n", "Separator between output messages")
	flag.BoolVar(&Verbose, "v", false, "Sets logging level to Debug")
	flag.Parse()
}

var unresolvedMsgs = make(map[string]string)

func main() {
	logOptions := slog.HandlerOptions{}
	if Verbose {
		logOptions.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &logOptions))
	slog.SetDefault(logger)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("error creating Discord session: %s", err)
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	dg.AddHandler(reactionAdd)

	// Just like the ping pong example, we only care about receiving message
	// events in this example.
	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsDirectMessageReactions |
		discordgo.PermissionAddReactions

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatalf("error opening connection: %s", err)
	}
	defer dg.Close()

	channel, err := dg.UserChannelCreate(UserID)
	if err != nil {
		log.Fatalf("error creating DM channel: %s", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(utils.SplitBySeparator([]byte(MsgSeparator)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		m, err := dg.ChannelMessageSend(channel.ID, line)
		if err != nil {
			log.Fatalf("error sending message: %s", err)
		}
		dg.MessageReactionAdd(channel.ID, m.ID, WatchEmoji)

		unresolvedMsgs[m.ID] = line
	}

	// Wait here until CTRL-C or other term signal is received.
	slog.Debug("Bot is now running.  Press CTRL-C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	for messageID := range unresolvedMsgs {
		dg.MessageReactionRemove(channel.ID, messageID, WatchEmoji, "@me")
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// We create the private channel with the user who sent the message.
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		slog.Error("error creating channel", "error", err)
		return
	}
	// Then we send the message through the channel we created.
	_, err = s.ChannelMessageSend(channel.ID, "Pong!")
	if err != nil {
		slog.Error("error sending DM message", "error", err)
	}
}

func reactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Skip if the reaction is from the bot
	if r.UserID == s.State.User.ID {
		return
	}

	msg, ok := unresolvedMsgs[r.MessageID]
	if !ok {
		slog.Info("Message not found in unresolved messages, skipping", "messageID", r.MessageID)
		return
	}

	outputMsg(msg, r.Emoji.Name)

	s.MessageReactionRemove(r.ChannelID, r.MessageID, WatchEmoji, "@me")
	delete(unresolvedMsgs, r.MessageID)
	if len(unresolvedMsgs) == 0 {
		sc <- syscall.SIGTERM
	}
}

func outputMsg(originalMsg, reaction string) {
	b := strings.Builder{}
	b.WriteString(originalMsg)
	b.WriteString(Separator)
	b.WriteString(reaction)
	b.WriteString(OutSeparator)

	fmt.Fprint(os.Stdout, b.String())
}
