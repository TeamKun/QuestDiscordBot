package task

import (
	"github.com/TeamKun/QuestDiscordBot/src/config"
	"github.com/TeamKun/QuestDiscordBot/src/quest"
)

type Channel struct {
	Name      string
	ChannelId string
}

/**
チャンネルの情報を取得する.

@params コンフィグ

@return チャンネル情報
*/
func getChannel(config config.Config) []Channel {

	var channels []Channel

	// コンフィグからチャンネルIDを取得
	channelNames := [...]string{quest.NOT_ORDERD,
		quest.WAITING_FOR_REVIEW,
		quest.WAITING_FOR_FINAL_REVIEW}

	// チャンネルIDとステータスを紐づける
	for _, channelName := range channelNames {
		var channel Channel

		channel.Name = channelName

		// 未受注チャンネル
		if channelName == quest.NOT_ORDERD {
			channel.ChannelId = config.NotOrderdChannel

			// レビュー待ち
		} else if channelName == quest.WAITING_FOR_REVIEW {
			channel.ChannelId = config.WaitingForReviewChannel

			// 最終レビュー待ち
		} else if channelName == quest.WAITING_FOR_FINAL_REVIEW {
			channel.ChannelId = config.WaitingForFinalReviewChannel
		}

		channels = append(channels, channel)
	}

	return channels
}
