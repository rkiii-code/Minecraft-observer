package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rkiii-code/Minecraft-observer/bot/commands"
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN is empty")
	}
	guildID := os.Getenv("DISCORD_GUILD_ID")
	channelID := os.Getenv("DISCORD_CHANNEL_ID")
	if channelID == "" {
		log.Println("[WARN] DISCORD_CHANNEL_ID is empty; log notifications will be disabled")
	}
	container := os.Getenv("MC_CONTAINER_NAME")
	if container == "" {
		container = "mc_bedrock"
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("new session: %v", err)
	}

	// スラッシュコマンド
	commands.Register(commands.NewSay())
	commands.Register(commands.NewPing())
	dg.AddHandler(commands.DispatchInteraction)
	dg.Identify.Intents = 0

	if err := dg.Open(); err != nil {
		log.Fatalf("open session: %v", err)
	}
	defer dg.Close()

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

	// === ここから追加：Dockerログ監視を開始 ===
	if channelID != "" {
		go func() {
			// 監視したいパターンはここで増やせる
			re := regexp.MustCompile(`Player connected`)
			// 失敗しても再試行するループ
			for {
				if err := streamDockerLogsAndNotify(dg, container, channelID, re); err != nil {
					log.Printf("[logwatch] error: %v (retry in 3s)", err)
					time.Sleep(3 * time.Second)
				}
			}
		}()
	}
	// === 追加ここまで ===

	log.Println("Bot is running. Press Ctrl+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}

// docker logs -f を実行し、正規表現にヒットした行をDiscordへ送る
func streamDockerLogsAndNotify(dg *discordgo.Session, container, channelID string, re *regexp.Regexp) error {
	// --since=5m で再起動時のドバッと通知をある程度抑制
	cmd := exec.Command("docker", "logs", "-f", "--since=5m", container)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("StdoutPipe: %w", err)
	}
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start docker logs: %w", err)
	}

	sc := bufio.NewScanner(stdout)
	errSc := bufio.NewScanner(stderr)
	go func() {
		for errSc.Scan() {
			log.Printf("[docker-err] %s", errSc.Text())
		}
	}()

	log.Printf("[logwatch] following container=%s", container)
	for sc.Scan() {
		line := sc.Text()
		if re.MatchString(line) {
			msg := fmt.Sprintf("**[MC]** %s  %s", time.Now().Format("2006-01-02 15:04:05"), line)
			// ここで @here にしたいなら msg = "@here " + msg （サーバ側権限に依存）
			if _, err := dg.ChannelMessageSend(channelID, msg); err != nil {
				log.Printf("[logwatch] send failed: %v", err)
			} else {
				log.Printf("[logwatch] posted: %s", line)
			}
		}
	}
	// scannerエラー
	if err := sc.Err(); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("scanner: %w", err)
	}
	// docker logs プロセス終了（コンテナ停止など）→ 呼び出し元でリトライ
	return fmt.Errorf("docker logs exited")
}
