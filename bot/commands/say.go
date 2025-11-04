package commands

import "github.com/bwmarrin/discordgo"

func NewSay() Command {
	def := &discordgo.ApplicationCommand{
		Name:        "say",
		Description: "Botに発言させる",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "発言内容",
				Required:    true,
			},
		},
	}
	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		text := i.ApplicationCommandData().Options[0].StringValue()
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: text,
			},
		})
	}
	return Command{Def: def, Handler: handler}
}
