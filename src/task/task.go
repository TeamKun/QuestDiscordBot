package task

import (
	"log"
	"time"

	"github.com/TeamKun/QuestDiscordBot/src/config"
	"github.com/TeamKun/QuestDiscordBot/src/quest"
	"github.com/bwmarrin/discordgo"
)

/** Notionから取得したクエスト */
var notionQuests []quest.Quest

/* チャンネルに投稿中のクエスト*/
var postedQuests []quest.Quest

/* エラー */
var err error

/* 実行時の時刻 */
var currentTime time.Time

/* チャンネルID */
var channelId string

/**
メインロジック
@param ディスコードセッション
@param コンフィグ
*/
func Task(discordSession *discordgo.Session, config config.Config) {

	// discordのチャンネル情報を取得
	channels := getChannel(config)

	// 定期実行タスク
	for {

		for _, channel := range channels {

			channelId = channel.ChannelId

			// 実行時の時刻を取得
			currentTime = time.Now()

			// Notionから未受注クエストを取得
			notionQuests, err = quest.GetQuestByStatus(config.NotionPageId, channel.Name)
			if err != nil {
				log.Println(err)
				time.Sleep(30 * time.Second)
				continue
			}

			// チャンネルに投稿されているクエストを取得
			postedQuests, err = getQuestsByChannel(discordSession, channelId)
			if err != nil {
				log.Println(err)
				time.Sleep(30 * time.Second)
				continue
			}

			// 未受注クエストとして取得されなかったメッセージを削除
			deleteQuestMessage(discordSession, channelId, postedQuests, notionQuests, currentTime, channel.Name)

			// 新規未受注クエストを投稿
			messageCreateNewQuest(discordSession, channelId, notionQuests, postedQuests, currentTime, channel.Name)

			//　タイトルが変更されていれば修正
			renameQuestTitle(discordSession, channelId, postedQuests, notionQuests)
		}

		// コンフィグで設定した時間待機
		time.Sleep(time.Duration(config.ProcessingSpan) * time.Second)
	}
}
