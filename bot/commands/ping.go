package commands

import "github.com/bwmarrin/discordgo"

func NewPing() Command {
	def := &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Pongを返す",
	}
	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong!",
			},
		})
	}
	return Command{Def: def, Handler: handler}
}
