package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"strconv"

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
	Email string `xml:"cemail"`
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

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", HomeHandler)

	r.HandleFunc("/feed/{type}/{id}", FeedHandler)

	r.Headers("Content-Type", "application/xml")

	log.Fatal(http.ListenAndServe(":8071", r))
}

func enclosure(u string) string {
	return "http://od.qingting.fm/" + u
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

func FeedHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	fmt.Println(vars["type"])

	var output []byte

	switch vars["type"] {
	case "qingting":
		output = qingting(vars["id"])
	default:
		output = []byte("类型不存在")
	}

	w.Write(output)
	w.WriteHeader(http.StatusOK)
	// w.
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

	body, _ := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	json := gjson.ParseBytes(body)

	rss := rss{
		Version:     "2",
		Itunes:      "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Title:       json.Get("data.name").String(),
		PubDate:     json.Get("data.update_time").String(),
		Description: json.Get("data.desc").String(),
		Language:    "zh-cn",
		Link:        link,
		Author: []string{
			json.Get("data.podcasters.name").String(),
		},
		Image: Image{
			Href: json.Get("data.img_url").String(),
		},
		Subtitle: json.Get("data.name").String(),
		Summary:  json.Get("data.desc").String(),
		Owner: Owner{
			Name:  json.Get("data.podcasters.name").String(),
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

		url := "http://i.qingting.fm/wapi/channels/5142813/programs/page/" + strconv.Itoa(i)

		resp, _ := http.Get(url)

		// "http://i.qingting.fm/wapi/channels/5142813/programs/page/1"

		body, _ := ioutil.ReadAll(resp.Body)

		ijsonArr := gjson.ParseBytes(body).Get("data").Array()

		for _, ijson := range ijsonArr {
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
				PubDate:  ijson.Get("update_time").String(),
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
