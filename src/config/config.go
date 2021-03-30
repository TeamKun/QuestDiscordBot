package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	ChannelId      string `json:"channelId"`
	NotionPageId   string `json:"notionPageId"`
	ProcessingSpan int    `json:"processingSpan"`
}

/**
コンフィグを取得する.
@return コンフィグ
*/
func LoadConfig() Config {
	jsonString, err := ioutil.ReadFile("../config.json")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	config := new(Config)
	err = json.Unmarshal([]byte(jsonString), config)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	return *config
}
