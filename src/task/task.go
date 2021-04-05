package task

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/TeamKun/QuestDiscordBot/src/config"
	"github.com/TeamKun/QuestDiscordBot/src/quest"
	"github.com/bwmarrin/discordgo"
)

/** 未受注クエスト */
var notOrderdQuests []quest.Quest

/* チャンネルに投稿中のクエスト*/
var postedQuests []quest.Quest

/* エラー */
var err error

/* 実行時の時刻 */
var currentTime time.Time

/**
メインロジック
@param ディスコードセッション
@param コンフィグ
*/
func Task(discordSession *discordgo.Session, config config.Config) {

	/* チャンネルID */
	channelId := config.ChannelId

	// 初回起動時、チャンネル内のメッセージをクリア
	// messageDeleteAll(discordSession, channelId)

	// 定期実行タスク
	for {

		// 実行時の時刻を取得
		currentTime = time.Now()

		// Notionから未受注クエストを取得
		notOrderdQuests, err = quest.GetQuestByStatus(config.NotionPageId, quest.NOT_ORDERD)
		if err != nil {
			log.Println(err)
			time.Sleep(30 * time.Second)
			continue
		}

		// チャンネルからメッセージを取得
		messages, err := discordSession.ChannelMessages(channelId, 100, "1", "1", "1")
		if err != nil {
			log.Println(err)
			time.Sleep(30 * time.Second)
			continue
		}
		// 取得したメッセージリストをクエスト型リストに変換
		postedQuests = parceMessagesToQuests(messages)

		// Notionから取得されなかったメッセージを削除
		deleteQuest(discordSession, channelId, postedQuests, notOrderdQuests, currentTime)

		// Notionから新たに取得されたクエストを投稿
		messageCreateNewQuest(discordSession, channelId, notOrderdQuests, postedQuests, currentTime)

		// コンフィグで設定した時間待機
		time.Sleep(time.Duration(config.ProcessingSpan) * time.Second)
	}
}

/**
チャンネル内のメッセージをすべて削除する

@param ディスコードセッション

@param チャンネルID
*/
func messageDeleteAll(discordSessin *discordgo.Session, channelId string) {
	messages, err := discordSessin.ChannelMessages(channelId, 100, "1", "1", "1")
	if err != nil {
		return
	}

	for _, message := range messages {
		discordSessin.ChannelMessageDelete(channelId, message.ID)
	}
}

/**
クエストメッセージを送信する

@param ディスコードセッション

@param チャンネルID

@param クエスト
*/
func sendQuestMessage(discordSessin *discordgo.Session, channelId string, quest quest.Quest) {
	baseMessage := "**TITLE**\nURL"

	// 文字列を置き換え
	message := strings.Replace(baseMessage, "TITLE", quest.Title, 1)
	message = strings.Replace(message, "URL", quest.PageUrl, 1)

	// クエストメッセージを送信
	discordSessin.ChannelMessageSend(channelId, message)
}

/**
メッセージリストをクエスト型に変換する

@params メッセージリスト

@return クエストリスト
*/
func parceMessagesToQuests(messages []*discordgo.Message) []quest.Quest {

	var quests []quest.Quest
	for _, message := range messages {

		// BOT以外が発言したメッセージを無視
		if !message.Author.Bot {
			continue
		}
		quests = append(quests, quest.ParseMessageToQuest(message))
	}
	return quests
}

/**
未受注として取得されなかったクエストをチャンネルから削除する

@params ディスコードセッション

@params チャンネルID

@params 投稿されているクエストリスト

@params 未受注クエストリスト

*/
func deleteQuest(discordSession *discordgo.Session,
	channnelId string,
	postedQuests []quest.Quest,
	notOrderdQuests []quest.Quest,
	currntTime time.Time) []quest.Quest {

	for _, quespostedQuest := range postedQuests {
		if !isQuestsArrayContains(quespostedQuest, notOrderdQuests) {
			discordSession.ChannelMessageDelete(channnelId, quespostedQuest.MessageId)

			// ログを出力
			fmt.Printf("%v [削除] %v\n", currntTime, quespostedQuest.Title)
		}
	}
}

/**
未投稿の未受注クエストをチャンネルに投稿する.

@params ディスコードセッション

@params チャンネルID

@params 未受注クエストリスト

@params 投稿されているクエストリスト
*/
func messageCreateNewQuest(discordSession *discordgo.Session,
	channnelId string,
	notOrderdQuests []quest.Quest,
	postedQuests []quest.Quest,
	currntTime time.Time) {

	for _, notOrderdQuest := range notOrderdQuests {
		if !isQuestsArrayContains(notOrderdQuest, postedQuests) {
			sendQuestMessage(discordSession, channnelId, notOrderdQuest)

			// ログを出力
			fmt.Printf("%v [追加] %v\n", currntTime, notOrderdQuest.Title)
		}
	}
}

/**
対象のクエストがクエストリストに含まれているか判定する.

@params クエスト

@params 検索対象リスト

@return true: 一致 false: 不一致
*/
func isQuestsArrayContains(quest quest.Quest, targetQuests []quest.Quest) bool {

	if len(targetQuests) == 0 {
		return false
	}
	for _, targetQuest := range targetQuests {
		if targetQuest.PageUrl == quest.PageUrl {
			return true
		}
	}
	return false
}
