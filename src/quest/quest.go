package quest

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
)

type Quest struct {
	Title     string
	PageUrl   string
	Status    string
	MessageId string
}

/**
メッセージをクエスト型に変換する

@params メッセージ

@return クエスト
*/
func ParseMessageToQuest(message *discordgo.Message) Quest {
	var quest Quest

	slice := strings.Split(message.Content, "\n")

	quest.Title = strings.Replace(slice[0], "*", "", -1)
	quest.PageUrl = slice[1]
	quest.MessageId = message.ID

	return quest
}

/**
テーブルからクエストすべてを取得する.
*/
func GetQuestAll(pageId string) ([]Quest, error) {
	var quests []Quest
	client := &notionapi.Client{}

	page, err := client.DownloadPage(pageId)

	// ページ取得失敗
	if err != nil {
		return nil, err
	}

	block := page.Root().Content
	html := string(tohtml.ToHTML(block[0].Page))

	trReg := regexp.MustCompile(`<tr.*?</tr>`)
	trList := trReg.FindAllString(html, -1)
	statusIndex := getStatusColumnIndex(trList[0])

	for i := 0; i < len(trList); i++ {

		// ヘッダー行を無視
		if i == 0 {
			continue
		}

		tdReg := regexp.MustCompile(`<td.*?</td>`)
		tdList := tdReg.FindAllString(trList[i], -1)

		// タイトルを取得
		title := getTitle(tdList[0])
		// urlを取得
		url := getPageUrl(trList[i])
		// ステータス
		status := getStatus(tdList[statusIndex])

		var quest Quest
		quest.Title = title
		quest.PageUrl = url
		quest.Status = status

		quests = append(quests, quest)
	}
	return quests, nil
}

/**
テーブルから指定したステータスのクエストを取得する.
*/
func GetQuestByStatus(pageId string, status string) ([]Quest, error) {

	var quests []Quest

	// すべてのクエストを取得
	allQuests, err := GetQuestAll(pageId)

	// 取得失敗
	if err != nil {
		return nil, err
	}

	// ステータスが一致しているもののみ取得
	for i := 0; i < len(allQuests); i++ {
		if allQuests[i].Status == status {
			quests = append(quests, allQuests[i])
		}
	}

	return quests, nil

}

/**
ステータスカラムのインデックスを取得する.
*/
func getStatusColumnIndex(tr string) int {
	thReg := regexp.MustCompile(`<th.*?</th>`)
	statusReg := regexp.MustCompile(`ステータス`)
	thList := thReg.FindAllString(tr, -1)
	for i := 0; i < len(thList); i++ {
		if statusReg.MatchString(thList[i]) {
			return i
		}
	}
	return -1
}

/**
td要素からクエストタイトルを取得する.
*/
func getTitle(td string) string {
	reg := regexp.MustCompile(`<a.*?</a>`)
	title := reg.FindString(td)

	// aタグがない場合
	if title == "" {
		reg = regexp.MustCompile(`>.*?<`)
		title = reg.FindString(td)
		title = strings.Replace(title, `>`, "", -1)
		title = strings.Replace(title, `<`, "", -1)
		return title
	}

	reg = regexp.MustCompile(`>.*?<`)
	title = reg.FindString(title)

	title = strings.Replace(title, `>`, "", -1)
	title = strings.Replace(title, `<`, "", -1)
	return title
}

/**
td要素からステータスを取得する.
*/
func getStatus(td string) string {
	reg := regexp.MustCompile(`>.*?<`)
	status := reg.FindString(td)
	status = strings.Replace(status, `>`, "", -1)
	status = strings.Replace(status, `<`, "", -1)
	return status
}

/**
tr要素のid属性からクエストページのURLを取得する.
*/
func getPageUrl(tr string) string {

	urlBase := "https://www.notion.so/"
	// id属性部分を抽出
	idReg := regexp.MustCompile(`id=".*?"`)

	// 不要部分を排除
	id := strings.Replace(idReg.FindString(tr), `id="`, "", -1)
	id = strings.Replace(id, `"`, "", -1)
	id = strings.Replace(id, `-`, "", -1)

	// urlベースと結合してreturn
	return urlBase + id
}
