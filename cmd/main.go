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
	// Setup logging
	logOptions := slog.HandlerOptions{}
	if Verbose {
		logOptions.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &logOptions))
	slog.SetDefault(logger)

	// Setup provider
	discordP := providers.CreateDiscord(Token, UserID)
	var provider providers.MsgProvider
	provider = &discordP
	provider.Init()
	defer provider.Close()

	// Send messages and add watch reactions
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

	slog.Info("Bot is now running.  Press CTRL-C to exit.")
	var sc = make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Wait here until:
	// * All messages get reactions
	// * CTRL-C or other term signal is received.
EventsLoop:
	for len(unresolvedMsgs) > 0 {
		select {
		case msg := <-provider.ListenToMessages():
			onNewMessage(provider, msg)
		case reaction := <-provider.ListenToReactions():
			onNewReaction(provider, reaction)
		case <-sc:
			slog.Info("shutting down...")
			break EventsLoop
		}
	}

	// Cleanup unresolved Messages
	for messageID := range unresolvedMsgs {
		provider.RemoveReaction(messageID, WatchEmoji)
	}
}

func onNewMessage(_ providers.MsgProvider, msg providers.Message) {
	slog.Debug("handling msg", "msg", msg.Content)

	// Check if referenced
	if msg.ReferencedMsgID == "" {
		slog.Info("message does not reference another message, skipping", "messageID", msg.ID)
		return
	}

	// Get referenced message
	refMsg, ok := unresolvedMsgs[msg.ReferencedMsgID]
	if !ok {
		slog.Info("message not found in unresolved messages, skipping", "messageID", msg.ID)
		return
	}

	outputMsg(refMsg, msg.Content)
}

func onNewReaction(p providers.MsgProvider, r providers.Reaction) {
	slog.Debug("handling r", "r", r.Content)
	msg, ok := unresolvedMsgs[r.MessageID]
	if !ok {
		slog.Info("message not found in unresolved messages, skipping", "messageID", r.MessageID)
		return
	}

	outputMsg(msg, r.Content)

	p.RemoveReaction(r.MessageID, WatchEmoji)
	delete(unresolvedMsgs, r.MessageID)
}

func outputMsg(originalMsg, reaction string) {
	b := strings.Builder{}
	b.WriteString(originalMsg)
	b.WriteString(Separator)
	b.WriteString(reaction)
	b.WriteString(OutSeparator)

	fmt.Fprint(os.Stdout, b.String())
}
