package providers

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/erykksc/chatreply/pkg/configuration"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func CreateTelegram(config configuration.Configuration) (MsgProvider, error) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	slog.Debug("config", "chatID", config.Telegram.ChatID)
	return &Telegram{
		Token:           config.Telegram.Token,
		ChatID:          config.Telegram.ChatID,
		context:         ctx,
		cancelCtx:       cancelCtx,
		messageChannel:  make(chan Message),
		reactionChannel: make(chan Reaction),
	}, nil
}

type Telegram struct {
	bot             *bot.Bot
	context         context.Context
	cancelCtx       context.CancelFunc
	Token           string
	ChatID          string
	messageChannel  chan Message
	reactionChannel chan Reaction
}

func (t *Telegram) Init() error {

	opts := []bot.Option{
		bot.WithAllowedUpdates(bot.AllowedUpdates{
			"message",
			"message_reaction",
		}),
		bot.WithDefaultHandler(t.handler),
	}

	b, err := bot.New(t.Token, opts...)
	if err != nil {
		return err
	}

	t.bot = b

	slog.Debug("Starting Telegram bot")
	go b.Start(t.context)
	slog.Debug("Telegram bot started")

	return nil
}

func (t *Telegram) Close() {
	slog.Debug("Closing Telegram bot")
	t.cancelCtx()
	close(t.messageChannel)
	close(t.reactionChannel)
}

func (t *Telegram) MessagesChannel() chan Message {
	return t.messageChannel
}
func (t *Telegram) ReactionsChannel() chan Reaction {
	return t.reactionChannel
}
func (t *Telegram) SendMessage(msg string, asText bool) (sentMsgID string, err error) {
	sentMsg, err := t.bot.SendMessage(t.context, &bot.SendMessageParams{
		ChatID: t.ChatID,
		Text:   msg,
	})
	if err != nil {
		return "", err
	}

	return strconv.Itoa(sentMsg.ID), err
}
func (t *Telegram) AddReaction(msgID, reaction string) error {
	intMsgID, err := strconv.Atoi(msgID)
	if err != nil {
		return err
	}

	_, err = t.bot.SetMessageReaction(t.context, &bot.SetMessageReactionParams{
		ChatID:    t.ChatID,
		MessageID: intMsgID,
		Reaction: []models.ReactionType{{
			Type: models.ReactionTypeTypeEmoji,
			ReactionTypeEmoji: &models.ReactionTypeEmoji{
				Type:  models.ReactionTypeTypeEmoji,
				Emoji: reaction,
			},
		}},
	})
	return err
}
func (t *Telegram) RemoveReaction(msgID, reaction string) error {
	return errors.ErrUnsupported
}

func (t *Telegram) handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	switch {
	case update.Message != nil:
		slog.Debug("Telegram message received", "chatID", update.Message.Chat.ID)
		t.messageHandler(update.Message)
	case update.MessageReaction != nil:
		slog.Debug("Telegram message reaction received")
		t.reactionHandler(update.MessageReaction)
	case update.MessageReactionCount != nil:
		slog.Debug("Telegram message reaction count received")
	}
}

func (t *Telegram) messageHandler(msg *models.Message) {
	if msg == nil {
		slog.Error("Called messageHandler with a nil ")
		return
	}

	var refMsgID string
	if msg.ReplyToMessage != nil {
		refMsgID = strconv.Itoa(msg.ReplyToMessage.ID)
	}

	t.messageChannel <- Message{
		ID:              strconv.Itoa(msg.ID),
		ChatID:          strconv.FormatInt(msg.Chat.ID, 10),
		ReferencedMsgID: refMsgID,
		Content:         msg.Text,
	}
}

func (t *Telegram) reactionHandler(reaction *models.MessageReactionUpdated) {
	if reaction == nil {
		slog.Error("Called reactionHandler with a nil reaction")
		return
	}

	reactions := reaction.NewReaction
	var reactionContent string

	newReaction := reactions[0]

	switch newReaction.Type {
	case models.ReactionTypeTypeEmoji:
		reactionContent = newReaction.ReactionTypeEmoji.Emoji
	case models.ReactionTypeTypeCustomEmoji:
		reactionContent = newReaction.ReactionTypeCustomEmoji.CustomEmojiID
	default:
		slog.Error("Unknown reaction type", "type", newReaction.Type)
		return
	}

	t.reactionChannel <- Reaction{
		MessageID: strconv.Itoa(reaction.MessageID),
		ChatID:    strconv.FormatInt(reaction.Chat.ID, 10),
		Content:   reactionContent,
	}
}
