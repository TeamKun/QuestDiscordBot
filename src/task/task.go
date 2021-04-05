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

type Channel struct {
	name      string
	channelId string
}

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

	channelNames := [...]string{quest.NOT_ORDERD,
		quest.WAITING_FOR_REVIEW,
		quest.WAITING_FOR_FINAL_REVIEW}

	var channels []Channel

	for _, channelName := range channelNames {
		var channel Channel

		channel.name = channelName

		if channelName == quest.NOT_ORDERD {
			channel.channelId = config.NotOrderdChannel

		} else if channelName == quest.WAITING_FOR_REVIEW {
			channel.channelId = config.WaitingForReviewChannel

		} else if channelName == quest.WAITING_FOR_FINAL_REVIEW {
			channel.channelId = config.WaitingForFinalReviewChannel
		}

		channels = append(channels, channel)
	}

	// 定期実行タスク
	for {

		for _, channel := range channels {

			channelId = channel.channelId

			// 実行時の時刻を取得
			currentTime = time.Now()

			// Notionから未受注クエストを取得
			notionQuests, err = quest.GetQuestByStatus(config.NotionPageId, channel.name)
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
			deleteQuest(discordSession, channelId, postedQuests, notionQuests, currentTime, channel.name)

			// 新規未受注クエストを投稿
			messageCreateNewQuest(discordSession, channelId, notionQuests, postedQuests, currentTime, channel.name)

			//　タイトルが変更されていれば修正
			renameQuestTitle(discordSession, channelId, postedQuests, notionQuests)
		}

		// コンフィグで設定した時間待機
		time.Sleep(time.Duration(config.ProcessingSpan) * time.Second)
	}
}

/**
タイトルが変更されていないかチェックし、変更されていれば投稿を修正

@params ディスコードセッション

@params チャンネルID

@params チャンネルに投稿されているクエスト

@params Notionから取得したクエスト
*/
func renameQuestTitle(discordSessin *discordgo.Session,
	channelId string,
	postedQuests []quest.Quest,
	notionQuests []quest.Quest) {

	for _, postedQuest := range postedQuests {
		for _, notionQuest := range notionQuests {

			// ページURLが一致しない
			if postedQuest.PageUrl != notionQuest.PageUrl {
				continue
			}

			// タイトルが一致
			if postedQuest.Title == notionQuest.Title {
				continue
			}

			// 投稿を修正
			editQuestMessage(discordSessin, channelId, postedQuest.MessageId, notionQuest.Title, postedQuest.PageUrl)
		}
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
func sendQuestMessage(discordSessin *discordgo.Session,
	channelId string,
	quest quest.Quest) {
	baseMessage := "**TITLE**\nURL"

	// 文字列を置き換え
	message := strings.Replace(baseMessage, "TITLE", quest.Title, 1)
	message = strings.Replace(message, "URL", quest.PageUrl, 1)

	// クエストメッセージを送信
	discordSessin.ChannelMessageSend(channelId, message)
}

/**
クエストメッセージを編集する

@params ディスコードセッション

@params メッセージID

@params チャンネルID

@params 修正後タイトル

@params NotionのURL
*/
func editQuestMessage(discordSession *discordgo.Session, channelId string, messageId string, newTitle string, notionUrl string) {

	baseMessage := "**TITLE**\nURL"

	// 文字列を置き換え
	message := strings.Replace(baseMessage, "TITLE", newTitle, 1)
	message = strings.Replace(message, "URL", notionUrl, 1)

	// クエストメッセージを送信
	discordSession.ChannelMessageEdit(channelId, messageId, message)

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
	notionQuests []quest.Quest,
	currntTime time.Time,
	channelName string) {

	for _, postedQuest := range postedQuests {
		if !isQuestsArrayContains(postedQuest, notionQuests) {
			discordSession.ChannelMessageDelete(channnelId, postedQuest.MessageId)

			// ログを出力
			fmt.Printf("%v [%vクエスト][削除] %v\n", currntTime, channelName, postedQuest.Title)
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
	notionQuests []quest.Quest,
	postedQuests []quest.Quest,
	currntTime time.Time,
	channelName string) {

	for _, notionQuest := range notionQuests {
		if !isQuestsArrayContains(notionQuest, postedQuests) {
			sendQuestMessage(discordSession, channnelId, notionQuest)

			// ログを出力
			fmt.Printf("%v [%vクエスト][追加] %v\n", currntTime, channelName, notionQuest.Title)
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

/**
チャンネルから投稿されたクエストリストを取得

@params ディスコードセッション

@params チャンネルID

@return 投稿されているクエストリスト

@return エラー
*/
func getQuestsByChannel(discordSession *discordgo.Session, channelId string) ([]quest.Quest, error) {
	// チャンネルからメッセージを取得
	messages, err := discordSession.ChannelMessages(channelId, 100, "1", "1", "1")
	if err != nil {
		log.Println(err)
		time.Sleep(30 * time.Second)
		return nil, err
	}
	// 取得したメッセージリストをクエスト型リストに変換
	return parceMessagesToQuests(messages), nil
}
