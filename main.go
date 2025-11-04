package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	// go.mod の module に合わせて import を調整してください
	// 例: module github.com/rkiii-code/Minecraft-observer
	"github.com/rkiii-code/Minecraft-observer/bot/commands"
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN is empty")
	}
	guildID := os.Getenv("DISCORD_GUILD_ID") // 開発中は入れると即反映。空ならグローバル登録

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("new session: %v", err)
	}

	// コマンド登録（ファイルを増やしたらここに1行足すだけ）
	commands.Register(commands.NewSay())
	commands.Register(commands.NewPing())

	// 共通ディスパッチをハンドラに
	dg.AddHandler(commands.DispatchInteraction)

	// スラッシュコマンドには Intent は不要だが、将来のために最小限
	dg.Identify.Intents = 0

	if err := dg.Open(); err != nil {
		log.Fatalf("open session: %v", err)
	}
	defer dg.Close()

	// アプリIDはセッションから取得（Client ID と同じ）
	appID := dg.State.User.ID
	log.Printf("appID=%s", appID)

	// コマンド作成
	for _, c := range commands.All() {
		cmd, err := dg.ApplicationCommandCreate(appID, guildID, c.Def)
		if err != nil {
			log.Fatalf("cannot create command %s: %v", c.Def.Name, err)
		}
		log.Printf("created command: %s (%s)", cmd.Name, cmd.ID)
	}

	log.Println("Bot is running. Press Ctrl+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// （任意）終了時にコマンド削除したい場合
	// cmds, _ := dg.ApplicationCommands(appID, guildID)
	// for _, c := range cmds { _ = dg.ApplicationCommandDelete(appID, guildID, c.ID) }
}
