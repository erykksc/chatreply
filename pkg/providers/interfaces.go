package providers

type Message struct {
	ID              string
	ChannelID       string
	ReferencedMsgID string
	Content         string
}

type Reaction struct {
	MessageID string
	ChannelID string
	Content   string
}

type MsgProvider interface {
	Init() error
	Close()
	// MessagesChannel returns a channel with messages from other users (not bot's own)
	MessagesChannel() chan Message
	// ReactionsChannel returns a channel with reactions of other users (not bot's own)
	ReactionsChannel() chan Reaction
	// asText argument should make the message be sent as pure text, don't try to embed data
	SendMessage(msg string, asText bool) (sentMsgID string, err error)
	AddReaction(msgID, reaction string) error
	RemoveReaction(msgID, reaction string) error
}
