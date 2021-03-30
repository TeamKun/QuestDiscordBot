package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TeamKun/QuestDiscordBot/src/config"
	"github.com/TeamKun/QuestDiscordBot/src/task"
	"github.com/bwmarrin/discordgo"
)

func main() {

	// Discordセッションを取得
	discordSession, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if isError("error creating Discord session,", err) {
		return
	}

	// Discordセッションをオープン
	err = discordSession.Open()
	if isError("error opening connection,", err) {
		return
	}

	// コンフィグの読み込み
	config := config.LoadConfig()

	//　メインロジックを非同期で起動
	go task.Task(discordSession, config)

	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	// 終了コマンドの受付
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Discordセッションをクローズ
	discordSession.Close()
}

/**
エラーチェックをする.

@param エラーメッセージ

@param エラー

@return true: エラーあり false: エラーなし
*/
func isError(errorMessage string, err error) bool {
	if err != nil {
		fmt.Println(errorMessage, err)
		return true
	}

	return false
}
