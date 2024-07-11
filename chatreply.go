package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/erykksc/chatreply/internal/configuration"
	"github.com/erykksc/chatreply/internal/providers"
	"github.com/erykksc/chatreply/internal/utils"
)

const WatchEmoji = "👀"

// Variables used for command line parameters
var (
	Separator    string
	MsgSeparator string
	OutSeparator string
	Verbose      bool
)

func init() {
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
	} else {
		logOptions.Level = slog.LevelWarn
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &logOptions))
	slog.SetDefault(logger)

	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	configPath := filepath.Join(xdgConfigHome, "chatreply", "conf.toml")
	config, err := configuration.LoadConfiguration(configPath)
	if err != nil {
		log.Fatalf("error loading configuration: %s", err)
	}

	provider, err := providers.CreateProvider(config)
	if err != nil {
		log.Fatalf("error creating provider: %s", err)
	}

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
			onReply(provider, Reply{
				RefMsgID: msg.ReferencedMsgID,
				Content:  msg.Content,
			})
		case reaction := <-provider.ListenToReactions():
			onReply(provider, Reply{
				RefMsgID: reaction.MessageID,
				Content:  reaction.Content,
			})
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

type Reply struct {
	RefMsgID string // ID of the message the reply was to
	Content  string
}

func (r Reply) String() string {
	return fmt.Sprintf("RefMsgID: %s, Content: %s", r.RefMsgID, r.Content)
}

// onReply is run when an unresolved message receives a response
func onReply(p providers.MsgProvider, reply Reply) {
	slog.Debug("handling reply", "reply", reply)
	// Check if referenced
	if reply.RefMsgID == "" {
		slog.Info("message does not reference another message, skipping", "reply", reply)
		return
	}
	unresolvedMsg, ok := unresolvedMsgs[reply.RefMsgID]
	if !ok {
		slog.Info("message not found in unresolved messages, skipping", "reply", reply)
		return
	}
	b := strings.Builder{}
	b.WriteString(unresolvedMsg)
	b.WriteString(Separator)
	b.WriteString(reply.Content)
	b.WriteString(OutSeparator)

	fmt.Fprint(os.Stdout, b.String())

	delete(unresolvedMsgs, reply.RefMsgID)

	err := p.RemoveReaction(reply.RefMsgID, WatchEmoji)
	if err != nil {
		slog.Error("error removing reaction", "error", err, "messageID", reply.RefMsgID, "reaction", WatchEmoji)
	}
}