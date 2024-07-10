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
	// ListenToMessages returns a channel with messages from other users (not bot's own)
	ListenToMessages() chan Message
	// ListenToReactions returns a channel with reactions of other users (not bot's own)
	ListenToReactions() chan Reaction
	SendMessage(msg string) (sentMsg Message, err error)
	AddReaction(msgID, reaction string) error
	RemoveReaction(msgID, reaction string) error
}
