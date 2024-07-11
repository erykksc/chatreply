package providers

import (
	"bufio"
	"mime"
	"os"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
)

func (d *Discord) SendStringHandler(text string) (sentMsgID string, err error) {
	m, err := d.s.ChannelMessageSend(d.UserChannel.ID, text)
	if err != nil {
		return "", err
	}
	return m.ID, nil
}

func (d *Discord) SendMessageWithFile(path string) (sentMsgID string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	baseName := filepath.Base(path)
	reader := bufio.NewReader(file)

	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)

	msg, err := d.s.ChannelMessageSendComplex(d.UserChannel.ID,
		&discordgo.MessageSend{
			Content: path,
			Files: []*discordgo.File{
				{
					Name:        baseName,
					ContentType: contentType,
					Reader:      reader,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return msg.ID, nil
}

func (d *Discord) SendImageMessage(path string) (sentMsgID string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	baseName := filepath.Base(path)
	reader := bufio.NewReader(file)

	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)

	msg, err := d.s.ChannelMessageSendComplex(d.UserChannel.ID,
		&discordgo.MessageSend{
			Files: []*discordgo.File{
				{
					Name:        baseName,
					ContentType: contentType,
					Reader:      reader,
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: path,
					Type:        discordgo.EmbedTypeImage,
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + baseName,
					},
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return msg.ID, nil
}
