package crawler

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

func Qingting(id string) []byte {

	link := "http://i.qingting.fm/wapi/channels/" + id

	resp, err := http.Get(link)

	if err != nil {
		panic(err.Error)
	}

	// if resp.StatusCode != 200 {
	// 	panic("Error")
	// }

	// "http://i.qingting.fm/wapi/channels/5142813/programs/page/1"

	body, _ := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	json := gjson.ParseBytes(body)

	if json.Get("code").Int() == 1 {
		return []byte("不存在")
	}

	t, err := time.Parse("2006-01-02 15:04:05", json.Get("data.update_time").String())

	if err != nil {
		fmt.Println(err.Error())
	}

	rss := rss{
		Version:     "2",
		Itunes:      "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Title:       json.Get("data.name").String(),
		PubDate:     t.Format(rfc2822),
		Description: json.Get("data.desc").String(),
		Language:    "zh-cn",
		Link:        link,
		Image: Image{
			Href: json.Get("data.img_url").String(),
		},
		Subtitle: json.Get("data.name").String(),
		Summary:  json.Get("data.desc").String(),
		Owner: Owner{
			Email: "LSvKing@Gmail.com",
		},
	}

	if json.Get("data.podcasters").Exists() {
		podcasters := json.Get("data.podcasters").Array()
		rss.Author = []string{
			podcasters[0].Get("name").String(),
		}
		rss.Owner.Name = podcasters[0].Get("name").String()

	}

	program_count := json.Get("data.program_count").Int()

	var items []Item

	pageCount := 1

	if program_count > 30 {
		if program_count%30 != 0 {
			pageCount = int(program_count/30) + 1
		} else {
			pageCount = int(program_count / 30)
		}

	}

	for i := 1; i <= pageCount; i++ {

		url := "http://i.qingting.fm/wapi/channels/" + id + "/programs/page/" + strconv.Itoa(i)

		resp, _ := http.Get(url)

		body, _ := ioutil.ReadAll(resp.Body)

		ijsonArr := gjson.ParseBytes(body).Get("data").Array()

		for _, ijson := range ijsonArr {

			t, err := time.Parse("2006-01-02 15:04:05", ijson.Get("update_time").String())

			if err != nil {
				fmt.Println(err.Error(), ijson.Get("update_time").String())
			}

			items = append(items, Item{
				Title:    ijson.Get("name").String(),
				Subtitle: ijson.Get("name").String(),
				Summary:  ijson.Get("desc").String(),
				Enclosure: Enclosure{
					Url:  enclosure(ijson.Get("file_path").String()),
					Type: "audio/mpeg",
				},
				Image: Image{
					Href: json.Get("data.img_url").String(),
				},
				PubDate:  t.Format(rfc2822),
				Duration: ijson.Get("duration").Int(),
				Guid: Guid{
					IsPermaLink: "true",
				},
			})
		}

		resp.Body.Close()
	}

	rss.Item = items

	output, err := xml.MarshalIndent(rss, "  ", "    ")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	o := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n" + string(output))

	return o
}

func enclosure(u string) string {
	return "http://od.qingting.fm/" + u
}
