package crawler

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/LSvKing/podcast/cache"

	"github.com/PuerkitoBio/goquery"
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

	// link := "http://www.ximalaya.com/album/" + id

	link := "https://www.ximalaya.com/youshengshu/20836133/"

	//resp, err := http.Get(link)

	//resty.SetDebug(true)
	cookies := []*http.Cookie{
		{
			Name:  "1&_token",
			Value: "82732660&F96F6A790ADC4NdV00658C2EBAD6967E63ACD97800D6127805D46E1D569708E047A1B47AD5337DD3",
		}, {
			Name:  "device_id",
			Value: "xm_1555047403033_judn2nfdv2d3d6",
		},
	}

	//client.SetCookies(cookies)
	resp, err := httpclient.
		WithCookie(cookies[0], cookies[1]).
		Get(link, nil)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return []byte("资源不存在")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
	}

	//realLink := doc.Url.RawPath
	//fmt.Println(realLink)

	//if doc.Find(".mgr-5").Size() > 0 {
	mgr5 := doc.Find(".mgr-5").Text()

	pubdateArr := strings.Split(mgr5, ":")

	date := strings.TrimSpace(pubdateArr[1])

	t, err := time.Parse("2006-01-02", date)

	if err != nil {
		fmt.Println(err.Error())
	}

	if body, err := cache.Get("xima-" + id + "|" + date); err == nil {
		return body
	}
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
		u := link + "/p" + strconv.Itoa(i) + "/"

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

	cache.Set("xima-"+id+"|"+date, o)

	return o
	// fmt.Println(string(o))

}
