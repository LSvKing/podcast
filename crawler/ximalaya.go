package crawler

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"time"
)

func Ximalaya(id string) []byte {

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

	link := "http://www.ximalaya.com/1000262/album/" + id

	resp, err := http.Get(link)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return []byte("资源不存在")
	}

	doc, _ := goquery.NewDocumentFromResponse(resp)

	resp.Body.Close()

	image, _ := doc.Find(".albumface180 img").Attr("src")

	h, err := doc.Find(".personal_header .username").Html()

	if err != nil {
		fmt.Println(err)
	}

	re, _ := regexp.Compile(`(?s)^(.*)\<i`)

	nickname := strings.TrimSpace(re.FindAllStringSubmatch(h, 1)[0][1])


	rss := rss{
		Title: doc.Find(".detailContent_title h1").Text(),
		Author: []string{
			nickname,
		},
		Summary:     TrimHtml(doc.Find(".detailContent_intro").Text()),
		Description: TrimHtml(doc.Find(".detailContent_intro").Text()),
		Subtitle:    doc.Find(".detailContent_title h1").Text(),
		Version:     "2",
		Itunes:      "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Link:        link,
		Language:    "zh-cn",
		Image: Image{
			Href: image,
		},
		Owner: Owner{
			Name:  nickname,
			Email: "LSvKing@Gmail.com",
		},
	}

	if doc.HasClass(".mgr5") {
		mgr5 := doc.Find(".mgr5").Text()
		pubdateArr := strings.Split(mgr5,":")

		t, err := time.Parse("2006-01-02", pubdateArr[1])

		if err != nil {
			fmt.Println(err.Error())
		}

		rss.PubDate = t.Format(rfc2822)

	}

	fmt.Println(rss)

	var items []Item

	respList, _ := http.Get("http://m.ximalaya.com/album/more_tracks?aid=" + id + "&page=1")

	body, _ := ioutil.ReadAll(respList.Body)

	respList.Body.Close()

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
