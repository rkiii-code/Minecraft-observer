package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Def     *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

var (
	all      []Command
	handlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){}
)

func Register(cmd Command) {
	all = append(all, cmd)
	if cmd.Def != nil && cmd.Handler != nil {
		handlers[cmd.Def.Name] = cmd.Handler
	}
}

func All() []Command { return all }

func DispatchInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	name := i.ApplicationCommandData().Name
	if h, ok := handlers[name]; ok {
		h(s, i)
		return
	}
	log.Printf("no handler for command %q", name)
}
