package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"
)

type rss struct {
	Link        string   `xml:"channel>link"`
	PubDate     string   `xml:"channel>pubDate"`
	Description string   `xml:"channel>description"`
	Itunes      string   `xml:"xmlns:itunes,attr"`
	Author      []string `xml:"channel>itunes:author"`
	Language    string   `xml:"channel>language"`
	Title       string   `xml:"channel>title"`
	Version     string   `xml:"version,attr"`
	Image       Image    `xml:"channel>itunes:image"`
	Summary     string   `xml:"channel>itunes:summary"`
	Subtitle    string   `xml:"channel>itunes:subtitle"`
	Owner       Owner    `xml:"channel>owner"`
	Item        []Item   `xml:"channel>item"`
}

type Image struct {
	Href string `xml:"href,attr"`
}

type Owner struct {
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

type Enclosure struct {
	Url  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}
type Guid struct {
	Text        string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
}

type Item struct {
	Title     string    `xml:"title" json:"name"`
	Subtitle  string    `xml:"itunes:subtitle" json:"name"`
	Author    string    `xml:"itunes:author" json:"name"`
	PubDate   string    `xml:"pubDate"`
	Summary   string    `xml:"itunes:summary"`
	Guid      Guid      `xml:"guid"`
	Image     Image     `xml:"itunes:image"`
	Enclosure Enclosure `xml:"enclosure"`
	Duration  int64     `xml:"itunes:duration" json:"duration"`
}

const rfc2822 = "Mon, 2 Jan 2006 15:04:05 UTC+8"

func main() {

	// ximalaya("3008")
	r := mux.NewRouter()

	r.HandleFunc("/", HomeHandler)

	r.HandleFunc("/feed/{type}/{id}.xml", FeedHandler)

	r.Headers("Content-Type", "application/xml")

	http.ListenAndServe(":8071", r)
}

func enclosure(u string) string {
	return "http://od.qingting.fm/" + u
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

func FeedHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	var output []byte

	switch vars["type"] {
	case "qingting":
		output = qingting(vars["id"])
	case "ximalaya":
		output = ximalaya(vars["id"])
	default:
		output = []byte("类型不存在")
	}

	w.Write(output)
	w.WriteHeader(http.StatusOK)
}

func qingting(id string) []byte {

	link := "http://i.qingting.fm/wapi/channels/" + id

	resp, err := http.Get(link)

	if err != nil {
		panic(err.Error)
	}

	// if resp.StatusCode != 200 {
	// 	panic("Error")
	// }

	// "http://i.qingting.fm/wapi/channels/5142813/programs/page/1"

	fmt.Println(resp.StatusCode)

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

	podcasters := json.Get("data.podcasters").Array()

	rss := rss{
		Version:     "2",
		Itunes:      "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Title:       json.Get("data.name").String(),
		PubDate:     t.Format(rfc2822),
		Description: json.Get("data.desc").String(),
		Language:    "zh-cn",
		Link:        link,
		Author: []string{
			podcasters[0].Get("name").String(),
		},
		Image: Image{
			Href: json.Get("data.img_url").String(),
		},
		Subtitle: json.Get("data.name").String(),
		Summary:  json.Get("data.desc").String(),
		Owner: Owner{
			Name:  podcasters[0].Get("name").String(),
			Email: "LSvKing@Gmail.com",
		},
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

func ximalaya(id string) []byte {

	type (
		xiMaList struct {
			Res      bool  `json:"res"`
			NextPage int   `json:"next_page"`
			SoundIds []int `json:"sound_ids"`
		}

		xiMaItem struct {
			ID       int
			PlayPath string `json:"play_path_64"`
			Duration int64  `json:"duration"`
			Title    string `json:"title"`
			NickName string `json:"nickname"`
			Intro    string `json:"intro"`
			CoverURL string `json:"cover_url"`
		}
	)

	link := "http://m.ximalaya.com/1000262/album/" + id

	resp, err := http.Get(link)

	if err != nil && resp.StatusCode == http.StatusNotFound {
		return []byte("资源不存在")
	}

	doc, _ := goquery.NewDocumentFromResponse(resp)

	resp.Body.Close()

	image, _ := doc.Find(".album-face img").Attr("src")

	rss := rss{
		Title: doc.Find(".album-tit").Text(),
		Author: []string{
			strings.TrimSpace(doc.Find(".nickname").Text()),
		},
		Summary:     strings.TrimSpace(doc.Find(".intro-breviary").Text()),
		Description: strings.TrimSpace(doc.Find(".intro-breviary").Text()),
		Subtitle:    doc.Find(".album-tit").Text(),
		Version:     "2",
		Itunes:      "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Link:        link,
		Language:    "zh-cn",
		Image: Image{
			Href: image,
		},
		Owner: Owner{
			Name:  strings.TrimSpace(doc.Find(".nickname").Text()),
			Email: "LSvKing@Gmail.com",
		},
		//PubDate:     doc.Find("span.mgr-5").Text(),
	}

	var items []Item

	resp, _ := http.Get("http://m.ximalaya.com/album/more_tracks?aid=" + id + "&page=1")

	body, _ := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	var xiList xiMaList

	json.Unmarshal(body, &xiList)

	var soundIds []int

	soundIds = append(soundIds, xiList.SoundIds...)

	for xiList.NextPage != 0 {
		resp, _ := http.Get("http://m.ximalaya.com/album/more_tracks?aid=" + id + "&page=" + strconv.Itoa(int(xiList.NextPage)))

		body, _ := ioutil.ReadAll(resp.Body)

		err := json.Unmarshal(body, &xiList)

		if err != nil {
			fmt.Println(err)
		}

		resp.Body.Close()

		soundIds = append(soundIds, xiList.SoundIds...)
	}

	//http://www.ximalaya.com/tracks/199237.json

	for _, id := range soundIds {

		resp, _ := http.Get("http://www.ximalaya.com/tracks/" + strconv.Itoa(int(id)) + ".json")

		body, _ := ioutil.ReadAll(resp.Body)

		var xiItem xiMaItem

		// fmt.Println(string(body))
		json.Unmarshal(body, &xiItem)

		items = append(items, Item{
			Title:    xiItem.Title,
			Subtitle: xiItem.Title,
			Author:   xiItem.NickName,
			//PubDate:,
			Summary: xiItem.Intro,
			Guid: Guid{
				IsPermaLink: "true",
			},
			Image: Image{
				Href: xiItem.CoverURL,
			},
			Enclosure: Enclosure{
				Url:  xiItem.PlayPath,
				Type: "audio/mpeg",
			},
			Duration: xiItem.Duration,
		})

		resp.Body.Close()
	}

	rss.Item = items

	output, err := xml.MarshalIndent(rss, "  ", "    ")

	// fmt.Printf("%+v\n", rss)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	o := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n" + string(output))

	return o
	// fmt.Println(string(o))

}
