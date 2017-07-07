package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"strings"

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
	Duration  string    `xml:"itunes:duration" json:"duration"`
}

const rfc2822 = "Mon, 2 Jan 2006 15:04:05 UTC+8"

func main() {

	ximalaya("3008")
	//r := mux.NewRouter()
	//
	//r.HandleFunc("/", HomeHandler)
	//
	//r.HandleFunc("/feed/{type}/{id}.xml", FeedHandler)
	//
	//r.Headers("Content-Type", "application/xml")
	//
	//http.ListenAndServe(":8071", r)
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
				Duration: ijson.Get("duration").String(),
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

func ximalaya(id string) {

	link := "http://m.ximalaya.com/1000262/album/" + id

	doc, _ := goquery.NewDocument(link)

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

	output, err := xml.MarshalIndent(rss, "  ", "    ")

	fmt.Printf("%+v\n", rss)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	o := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n" + string(output))

	fmt.Println(string(o))

	resp, _ := http.Get("http://m.ximalaya.com/album/more_tracks?aid=3008&page=1")

	body, _ := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	json := gjson.ParseBytes(body)

	var sound_ids []int64

	next_page := json.Get("next_page").Int()

	json.Get("sound_ids").ForEach(func(key, value gjson.Result) bool {
		sound_ids = append(sound_ids, value.Int())
		return true
	})

	for next_page != 0 {
		resp, _ := http.Get("http://m.ximalaya.com/album/more_tracks?aid=3008&page=" + strconv.Itoa(int(next_page)))

		body, _ := ioutil.ReadAll(resp.Body)

		json := gjson.ParseBytes(body)

		resp.Body.Close()

		json.Get("sound_ids").ForEach(func(key, value gjson.Result) bool {
			sound_ids = append(sound_ids, value.Int())
			return true
		})

		next_page = json.Get("next_page").Int()
	}

	//fmt.Println(sound_ids)

	//http://www.ximalaya.com/tracks/199237.json
	for _, id := range sound_ids {

		resp, _ := http.Get("http://www.ximalaya.com/tracks/" + strconv.Itoa(int(id)) + ".json")

		body, _ := ioutil.ReadAll(resp.Body)

		json := gjson.ParseBytes(body)

		//t := time.Parse("")

		items = append(items,Item{
			Title:json.Get("title").String(),
			Subtitle:json.Get("title").String(),
			Author:json.Get("nickname").String(),
			PubDate:,
			Summary:json.Get("intro").String(),
			Guid:Guid{
				IsPermaLink:"true",
			},
			Image:Image{},
			Enclosure:Enclosure{
				Url:json.Get("play_path_64").String(),
				Type:"audio/mpeg",
			},
			Duration:json.Get("duration").String(),
		})
		resp.Body.Close()

	}
}
