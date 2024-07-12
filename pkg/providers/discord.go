package providers

import (
	"errors"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/erykksc/chatreply/pkg/configuration"
)

// CreateDiscord is a factory function implementing ProviderFactoryFunc signature
// It is used by CreateProvider
func CreateDiscord(conf configuration.Configuration) (MsgProvider, error) {
	if conf.Discord.Token == "" {
		return nil, errors.New("discord token not provided")
	}

	if conf.Discord.UserID == "" {
		return nil, errors.New("discord user ID not provided")
	}

	return &Discord{
		Token:           conf.Discord.Token,
		UserID:          conf.Discord.UserID,
		messageChannel:  make(chan Message),
		reactionChannel: make(chan Reaction),
	}, nil
}

// Discord is a struct implementing MsgProvider interface
type Discord struct {
	s               *discordgo.Session
	UserChannel     *discordgo.Channel
	messageChannel  chan Message
	reactionChannel chan Reaction
	Token           string
	UserID          string
}

func (d *Discord) Init() error {
	session, err := discordgo.New("Bot " + d.Token)
	if err != nil {
		return err
	}
	d.s = session

	// Register the messageCreate func as a callback for MessageCreate events.
	d.s.AddHandler(d.onMessageCreate)
	d.s.AddHandler(d.onReactionAdd)

	d.s.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsDirectMessageReactions |
		discordgo.PermissionAddReactions

	// Open a websocket connection to Discord and begin listening.
	err = d.s.Open()
	if err != nil {
		slog.Error("error opening connection", "error", err)
		return err
	}

	channel, err := d.s.UserChannelCreate(d.UserID)
	if err != nil {
		slog.Error("error creating DM channel", "error", err)
		return err
	}
	d.UserChannel = channel
	return nil

}

func (d *Discord) Close() {
	close(d.messageChannel)
	close(d.reactionChannel)
	d.s.Close()
}

func (d *Discord) MessagesChannel() chan Message {
	return d.messageChannel
}

func (d *Discord) ReactionsChannel() chan Reaction {
	return d.reactionChannel
}

type MessageSendFunc func(text string) (sentMsgID string, err error)

func (d *Discord) SendMessage(msg string, asString bool) (sentMsgID string, err error) {
	var sendFunc MessageSendFunc

	if asString {
		slog.Debug("sending message as string (forced by argument)", "message", msg)
		sendFunc = d.SendStringHandler
	} else {
		sendFunc = d.chooseMessageSendFunc(msg)
	}

	// Use chosen handler
	sentMsgID, err = sendFunc(msg)

	return sentMsgID, err
}

func (d *Discord) chooseMessageSendFunc(msg string) MessageSendFunc {
	info, err := os.Stat(msg)

	// If msg isn't a filepath
	if err != nil {
		slog.Debug("sending message as string (message isn't a filepath)", "message", msg)
		return d.SendStringHandler
	}

	// Check if msg's filepath is a directory
	if info.IsDir() {
		slog.Debug("sending message as string (message is a directory path)", "message", msg)
		return d.SendStringHandler
	}

	// It is a filepath and not a directory
	return d.SendMessageWithFile
}

func (d *Discord) AddReaction(msgID, reaction string) error {
	err := d.s.MessageReactionAdd(d.UserChannel.ID, msgID, reaction)
	return err
}
func (d *Discord) RemoveReaction(msgID, reaction string) error {
	err := d.s.MessageReactionRemove(d.UserChannel.ID, msgID, reaction, "@me")
	return err
}

func (d *Discord) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	slog.Debug("message received", "message", m.Content)
	if m.Author.ID == s.State.User.ID {
		return
	}

	var refMsgID string
	if m.ReferencedMessage != nil {
		refMsgID = m.ReferencedMessage.ID
	}

	msg := Message{
		ID:              m.ID,
		ChatID:          m.ChannelID,
		ReferencedMsgID: refMsgID,
		Content:         m.Content,
	}

	d.messageChannel <- msg
}

func (d *Discord) onReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	slog.Debug("reaction added", "reaction", r.Emoji.Name)
	// Skip if the reaction is from the bot
	if r.UserID == s.State.User.ID {
		return
	}

	reaction := Reaction{
		MessageID: r.MessageID,
		ChatID:    r.ChannelID,
		Content:   r.Emoji.Name,
	}

	d.reactionChannel <- reaction
}
