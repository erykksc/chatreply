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

	"github.com/erykksc/notifr/internal/providers"
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

	discordP := providers.CreateDiscord(Token, UserID)

	var provider providers.MsgProvider
	provider = &discordP
	provider.Init()
	defer provider.Close()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(utils.SplitBySeparator([]byte(MsgSeparator)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		m, err := provider.SendMessage(line)
		if err != nil {
			log.Fatalf("error sending message: %s", err)
		}
		err = provider.AddReaction(m.ID, WatchEmoji)
		if err != nil {
			log.Fatalf("error adding reaction: %s", err)
		}

		unresolvedMsgs[m.ID] = line
	}

	// Wait here until CTRL-C or other term signal is received.
	slog.Info("Bot is now running.  Press CTRL-C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

EventsLoop:
	for len(unresolvedMsgs) > 0 {
		select {
		case msg := <-provider.ListenToMessages():
			slog.Debug("handling message", "message", msg.Content)
			provider.SendMessage(msg.Content)
		case reaction := <-provider.ListenToReactions():
			slog.Debug("handling reaction", "reaction", reaction.Content)
			msg, ok := unresolvedMsgs[reaction.MessageID]
			if !ok {
				slog.Info("message not found in unresolved messages, skipping", "messageID", reaction.MessageID)
				return
			}

			outputMsg(msg, reaction.Content)

			provider.RemoveReaction(reaction.MessageID, WatchEmoji)
			delete(unresolvedMsgs, reaction.MessageID)
			if len(unresolvedMsgs) == 0 {
				sc <- syscall.SIGTERM
			}
		case <-sc:
			slog.Info("shutting down...")
			break EventsLoop
		}
	}

	for messageID := range unresolvedMsgs {
		provider.RemoveReaction(messageID, WatchEmoji)
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
