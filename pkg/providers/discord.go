package providers

import (
	"errors"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/erykksc/chatreply/pkg/configuration"
)

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
		MessageChannel:  make(chan Message),
		ReactionChannel: make(chan Reaction),
	}, nil
}

type Discord struct {
	s               *discordgo.Session
	UserChannel     *discordgo.Channel
	MessageChannel  chan Message
	ReactionChannel chan Reaction
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
	d.s.AddHandler(d.messageCreate)
	d.s.AddHandler(d.reactionAdd)

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
	close(d.MessageChannel)
	close(d.ReactionChannel)
	d.s.Close()
}

func (d *Discord) ListenToMessages() chan Message {
	return d.MessageChannel
}

func (d *Discord) ListenToReactions() chan Reaction {
	return d.ReactionChannel
}

func (d *Discord) SendMessage(msg string) (sentMsg Message, err error) {
	m, err := d.s.ChannelMessageSend(d.UserChannel.ID, msg)
	if err != nil {
		return Message{}, err
	}
	return Message{
		ID:        m.ID,
		ChannelID: m.ChannelID,
	}, nil
}
func (d *Discord) AddReaction(msgID, reaction string) error {
	err := d.s.MessageReactionAdd(d.UserChannel.ID, msgID, reaction)
	return err
}
func (d *Discord) RemoveReaction(msgID, reaction string) error {
	err := d.s.MessageReactionRemove(d.UserChannel.ID, msgID, reaction, "@me")
	return err
}

func (d *Discord) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
		ChannelID:       m.ChannelID,
		ReferencedMsgID: refMsgID,
		Content:         m.Content,
	}

	d.MessageChannel <- msg
}

func (d *Discord) reactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	slog.Debug("reaction added", "reaction", r.Emoji.Name)
	// Skip if the reaction is from the bot
	if r.UserID == s.State.User.ID {
		return
	}

	reaction := Reaction{
		MessageID: r.MessageID,
		ChannelID: r.ChannelID,
		Content:   r.Emoji.Name,
	}

	d.ReactionChannel <- reaction
}
