package bot

import (
	"fmt"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/bwmarrin/discordgo"
)

func (bs *BotService) handleEcho(args string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	if len(args) == 0 {
		return nil
	}
	return bs.US.EmbedMessage(args)
}

func (bs *BotService) handleHelp(args string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	helpText := gl.MsgHelp

	for _, command := range bs.commandNames {
		helpText += fmt.Sprintf(gl.MsgUnorderedList, bs.US.FormatHelp(command, bs.handlersMap[command]))
	}

	return bs.US.EmbedMessage(helpText)
}
