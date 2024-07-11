package main

import (
	"bufio"
	"flag"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
	"github.com/erykksc/chatreply/pkg/configuration"
	"github.com/erykksc/chatreply/pkg/providers"
)

// Variables used for command line parameters
var (
	ConfigPath   string
	Separator    string
	MsgSeparator string
	OutSeparator string
	SkipReplies  bool
	WatchEmoji   string
	Verbose      bool
)

func init() {
	defaultConfPath := "$XDG_CONFIG_HOME/chatreply/conf.toml"
	flag.StringVar(&ConfigPath, "f", defaultConfPath, "Filepath of the config .toml file")
	flag.StringVar(&Separator, "s", ":", "Separator between message and emoji")
	flag.StringVar(&MsgSeparator, "msg-sep", "\n", "Separator between messages")
	flag.StringVar(&OutSeparator, "out-sep", "\n", "Separator between output messages")
	flag.BoolVar(&SkipReplies, "skip-replies", false, "Do not wait for replies, just send the messages")
	flag.StringVar(&WatchEmoji, "watch-emoji", "👀", "Emoji used to indicate the program is watching the message for a reply")
	flag.BoolVar(&Verbose, "v", false, "Sets logging level to Debug")
	flag.Parse()

	if ConfigPath == defaultConfPath {
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		ConfigPath = filepath.Join(xdgConfigHome, "chatreply", "conf.toml")
	}
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

	config, err := configuration.LoadConfiguration(ConfigPath)
	if err != nil {
		log.Fatalf("error loading configuration: %s", err)
	}

	provider := providers.Discord{
		Token:  config.Discord.Token,
		UserID: config.Discord.UserID,
	}

	provider.Init()
	defer provider.Close()

	imgPath := "./ketchup.jpeg"
	file, err := os.Open(imgPath)
	if err != nil {
		slog.Error("error opening file", "error", err)
		os.Exit(1)
	}
	baseName := filepath.Base(imgPath)
	reader := bufio.NewReader(file)

	provider.S.ChannelMessageSendComplex(provider.UserChannel.ID,
		&discordgo.MessageSend{
			// Content: "Test Message",
			Files: []*discordgo.File{
				{
					Name:        baseName,
					ContentType: "image/jpeg",
					Reader:      reader,
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Ketchup image",
					Type:  "image/jpeg",
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + baseName,
					},
				},
			},
		},
	)
}
