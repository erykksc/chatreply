package providers

type Message struct {
	ID        string
	ChannelID string
	Content   string
}

type Reaction struct {
	MessageID string
	ChannelID string
	Content   string
}

type MsgProvider interface {
	Init() error
	Close()
	ListenToMessages() chan Message
	ListenToReactions() chan Reaction
	SendMessage(msg string) (sentMsg Message, err error)
	AddReaction(msgID, reaction string) error
	RemoveReaction(msgID, reaction string) error
}
