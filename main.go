// main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	appCommands = []*discordgo.ApplicationCommand{
		{
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
		},
	}
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	guildID := os.Getenv("DISCORD_GUILD_ID")
	if token == "" || guildID == "" {
		log.Fatal("DISCORD_BOT_TOKEN と DISCORD_GUILD_ID を設定してください")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("New:", err)
	}
	dg.Identify.Intents = discordgo.IntentsGuilds

	// コマンドのハンドラ
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		if i.ApplicationCommandData().Name == "say" {
			text := i.ApplicationCommandData().Options[0].StringValue()
			// まずはエフェメラルで受付
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "送るね！",
					Flags:   discordgo.MessageFlagsEphemeral, // 呼び出し者にだけ見える
				},
			})
			// 実際のチャンネルに投稿（コマンドが来たチャンネルへ）
			channelID := i.ChannelID
			if _, err := s.ChannelMessageSend(channelID, text); err != nil {
				log.Println("send error:", err)
			}
		}
	})

	if err := dg.Open(); err != nil {
		log.Fatal("Open:", err)
	}
	log.Println("Bot started.")

	// コマンド登録（ギルドコマンドなら即時反映）
	registered := make([]*discordgo.ApplicationCommand, 0, len(appCommands))
	for _, ac := range appCommands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildID, ac)
		if err != nil {
			log.Fatalf("cmd create error: %v", err)
		}
		registered = append(registered, cmd)
	}
	log.Println("Commands registered.")

	// 終了待ち
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// クリーンアップ（コマンド削除）
	for _, cmd := range registered {
		_ = dg.ApplicationCommandDelete(dg.State.User.ID, guildID, cmd.ID)
	}
	_ = dg.Close()
}
