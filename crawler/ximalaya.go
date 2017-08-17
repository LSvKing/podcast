package crawler

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"io/ioutil"

	"github.com/PuerkitoBio/goquery"
	"google.golang.org/appengine"
	"google.golang.org/appengine/memcache"
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

	link := "http://www.ximalaya.com/album/" + id

	resp, err := http.Get(link)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return []byte("资源不存在")
	}

	// fmt.Println(string(resp.Body))

	doc, _ := goquery.NewDocumentFromResponse(resp)

	resp.Body.Close()

	realLink := doc.Url.String()

	//if doc.Find(".mgr-5").Size() > 0 {
	mgr5 := doc.Find(".mgr-5").Text()

	pubdateArr := strings.Split(mgr5, ":")

	data := strings.TrimSpace(pubdateArr[1])

	t, err := time.Parse("2006-01-02", data)

	if err != nil {
		fmt.Println(err.Error())
	}

	c := appengine.BackgroundContext()

	if item, err := memcache.Get(c, "xima-"+id+"|"+data); err == nil {
		return item.Value
	}

	//if cache.New().Has("xima-" + id + "|" + data) {
	//	c, _ := cache.New().Get("xima-" + id + "|" + data)
	//	return c
	//}
	//}

	image, _ := doc.Find(".albumface180 img").Attr("src")

	h, err := doc.Find(".personal_header .username").Html()

	if err != nil {
		fmt.Println(err)
	}

	re, _ := regexp.Compile(`(?Us)^(.*)\<`)

	fmt.Println(re.FindAllStringSubmatch(h, 1))
	nickname := strings.TrimSpace(re.FindAllStringSubmatch(h, 1)[0][1])

	rss := rss{
		Title: doc.Find(".detailContent_title h1").Text(),
		Author: []string{
			nickname,
		},
		Summary:     TrimHtml(doc.Find(".detailContent_intro article").Text()),
		Description: TrimHtml(doc.Find(".detailContent_intro article").Text()),
		Subtitle:    doc.Find(".detailContent_title h1").Text(),
		Version:     "2",
		Itunes:      "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Link:        link,
		Language:    "zh-cn",
		Image: Image{
			Href: image,
		},
		PubDate: t.Format(time.RFC1123),
		Owner: Owner{
			Name:  nickname,
			Email: "LSvKing@Gmail.com",
		},
	}

	if doc.Find(".mgr-5").Size() > 0 {
		mgr5 := doc.Find(".mgr-5").Text()

		pubdateArr := strings.Split(mgr5, ":")

		t, err := time.Parse("2006-01-02", strings.TrimSpace(pubdateArr[1]))

		if err != nil {
			fmt.Println(err.Error())
		}

		rss.PubDate = t.Format(time.RFC1123)
	}

	var items []Item
	page := 1

	if doc.Find(".pagingBar .pagingBar_page").Length() > 0 {
		pageCount := doc.Find(".pagingBar .pagingBar_page").Last().Prev().Text()
		page, _ = strconv.Atoi(pageCount)
	}

	for i := 1; i <= page; i++ {
		u := realLink + "?page=" + strconv.Itoa(i)

		docList, _ := goquery.NewDocument(u)

		docList.Find(".album_soundlist ul li").Each(func(i int, selection *goquery.Selection) {
			sound_id, _ := selection.Attr("sound_id")

			resp, _ := http.Get("http://www.ximalaya.com/tracks/" + sound_id + ".json")

			body, _ := ioutil.ReadAll(resp.Body)

			var xiItem xiMaItem

			json.Unmarshal(body, &xiItem)

			t, err := time.Parse("2006-01-02", strings.TrimSpace(selection.Find(".operate span").Text()))

			if err != nil {
				fmt.Println(err.Error())
			}

			pubDate := t.Format(time.RFC1123)

			items = append(items, Item{
				Title:    xiItem.Title,
				Subtitle: xiItem.Title,
				Author:   xiItem.NickName,
				PubDate:  pubDate,
				Summary:  xiItem.Intro,
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
		})
	}

	rss.Item = items

	fmt.Println(len(items))

	output, err := xml.MarshalIndent(rss, "  ", "    ")

	//fmt.Printf("%+v\n", rss)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	o := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n" + string(output))

	//cache.New().Set("xima-"+id+"|"+data, o, 365*24*time.Hour)

	it := &memcache.Item{
		Key:   "xima-" + id + "|" + data,
		Value: o,
	}

	if err := memcache.Add(c, it); err == memcache.ErrNotStored {
		fmt.Println("it with key  already exists", it.Key)
	} else if err != nil {
		//log.Errorf(ctx, "error adding item: %v", err)
	}

	return o
	// fmt.Println(string(o))

}
