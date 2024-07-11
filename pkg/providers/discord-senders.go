package providers

func (d *Discord) SendTextMessage(text string) (sentMsgID string, err error) {
	m, err := d.s.ChannelMessageSend(d.UserChannel.ID, text)
	if err != nil {
		return "", err
	}
	return m.ID, nil
}
