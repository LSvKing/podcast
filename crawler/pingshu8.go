package crawler

import (
	"net/http"
	"fmt"
	"encoding/xml"

	"podcast/cache"

	"golang.org/x/net/html/charset"
	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req"
	"regexp"
	"time"
)

func PingShu8(id string) []byte{
	siteUrl := "http://m.pingshu8.com"

	link:= "https://m.pingshu8.com/MusicList/mmc_" + id +".html"

	resp, err := http.Get(link)

	if err != nil {
		panic(err.Error)
	}

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return []byte("资源不存在")
	}

	utfBody, err := charset.NewReader(resp.Body, "gbk")

	if err != nil {
		// handler error
		fmt.Println(err)
	}

	doc, _ := goquery.NewDocumentFromReader(utfBody)
	Time:= doc.Find(".info div").Eq(2).Find("span").Text()

	if body, err := cache.Get("pingshu-" + id + "|" + Time); err == nil {
		return body
	}

	resp.Body.Close()

	Author := doc.Find(".info div").Eq(0).Find("span").First().Text()
	owner:= doc.Find(".info div").Eq(0).Find("span").Last().Text()

	t, err := time.Parse("2006/1/02 15:04:05", Time)

	if err != nil {
		fmt.Println(err)
	}


	rss := rss{
		Title: doc.Find(".bookname").Text(),
		Author: []string{
			Author,
		},
		Summary:     TrimHtml(doc.Find(".book_intro").Text()),
		Description: TrimHtml(doc.Find(".book_intro").Text()),
		Subtitle:    doc.Find(".bookname").Text(),
		Version:     "2",
		Itunes:      "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Link:        link,
		Language:    "zh-cn",
		Image: Image{
			Href: siteUrl + doc.Find(".bookimg img").AttrOr("src","null"),
		},
		PubDate: t.Format(time.RFC1123),
		Owner: Owner{
			Name:  owner,
			Email: "LSvKing@Gmail.com",
		},
	}

	var items []Item

	doc.Find("#playlist li").Each(func(i int, selection *goquery.Selection) {

		fmt.Println("items start",i)

		h:= selection.Find("a").AttrOr("href","null")

		re,_ := regexp.Compile(`\d+`)

		id := re.FindString(h)

		down_url := "http://www.pingshu8.com/bzmtv_Inc/download.asp?fid=" + id

		header := req.Header{
			"Referer": "http://www.pingshu8.com/down_228133.html",
			"Host": "www.pingshu8.com",
			"User-Agent": "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3269.3 Safari/537.36",
		}

		r, _ :=req.Head(down_url,header)

		items = append(items, Item{
			Title:    selection.Find("a").Text(),
			Subtitle: selection.Find("a").Text(),
			Author:   "LSvking",
			PubDate:  t.Format(time.RFC1123),
			Summary:  "",
			Guid: Guid{
				IsPermaLink: "true",
			},
			Image: Image{
				Href: siteUrl + doc.Find(".bookimg img").AttrOr("src","null"),
			},
			Enclosure: Enclosure{
				Url:  r.Response().Request.URL.String(),
				Type: "audio/mpeg",
			},
			//Duration: ,
		})

		fmt.Println("items end",i)
	})

	rss.Item = items

	output, err := xml.MarshalIndent(rss, "  ", "    ")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	o := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n" + string(output))

	cache.Set("pingshu-"+id+"|"+Time, o)

	return o
}
