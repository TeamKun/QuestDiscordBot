package task

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/TeamKun/QuestDiscordBot/src/quest"
	"github.com/bwmarrin/discordgo"
)

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
未受注として取得されなかったクエストをチャンネルから削除する

@params ディスコードセッション

@params チャンネルID

@params 投稿されているクエストリスト

@params 未受注クエストリスト
*/
func deleteQuestMessage(discordSession *discordgo.Session,
	channnelId string,
	postedQuests []quest.Quest,
	notionQuests []quest.Quest,
	currntTime time.Time,
	channelName string) {

	for _, postedQuest := range postedQuests {
		if !isQuestsArrayContains(postedQuest, notionQuests) {
			discordSession.ChannelMessageDelete(channnelId, postedQuest.MessageId)

			// ログを出力
			outPutLog(currntTime, channelName, "削除", postedQuest.Title)
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
			outPutLog(currntTime, channelName, "追加", notionQuest.Title)
		}
	}
}

/**
ログを出力する

@params 処理時刻

@params チャンネル名

@params 処理名

@params クエスト名
*/
func outPutLog(currentTime time.Time,
	channelName string,
	actionName string,
	questName string) {

	fmt.Printf("%v [%vクエスト][%v] %v\n", currentTime, channelName, actionName, questName)
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
クエストメッセージを編集する

@params ディスコードセッション

@params メッセージID

@params チャンネルID

@params 修正後タイトル

@params NotionのURL
*/
func editMessage(discordSession *discordgo.Session, channelId string, messageId string, newTitle string, notionUrl string) {

	baseMessage := "**TITLE**\nURL"

	// 文字列を置き換え
	message := strings.Replace(baseMessage, "TITLE", newTitle, 1)
	message = strings.Replace(message, "URL", notionUrl, 1)

	// クエストメッセージを送信
	discordSession.ChannelMessageEdit(channelId, messageId, message)
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
			editMessage(discordSessin, channelId, postedQuest.MessageId, notionQuest.Title, postedQuest.PageUrl)
		}
	}
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
