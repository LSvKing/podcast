package crawler

import (
	"net/http"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req"

	"strconv"
	"regexp"
	"log"
)

func Ivoix(id string)[]byte  {

	siteUrl := "http://m.ivoix.cn/book" + id

	resp, err := req.Get(siteUrl)

	mp3Url := "http://m.ivoix.cn/inc/audio.asp"

	if err != nil {
		log.Fatalln(err.Error)
	}

	if err != nil || resp.Response().StatusCode == http.StatusNotFound {
		return []byte("资源不存在")
	}

	mp3Resp, err := req.Get("http://m.ivoix.cn/js/h7.js")

	re,_ := regexp.Compile(`ahead=(\"|\')(.*)(\"|\');`)

	mp3Arr := re.FindAllStringSubmatch(mp3Resp.String(),-1)

	mp3Prefix:= mp3Arr[len(mp3Arr) - 1][2]

	fmt.Println(mp3Arr[len(mp3Arr) - 1][2])

	if err != nil {
		log.Fatalln(err)
	}

	doc, _ := goquery.NewDocumentFromResponse(resp.Response())

	pageNum := doc.Find(".pgsel option").Length()

	info:= doc.Find("#bookinfo")

	author:= info.Find("p").Eq(1).Text()
	owner := info.Find("p").Eq(0).Text()
	image := info.Find(".bookimg").AttrOr("src","null")

	t:=time.Now()

	rss := rss{
		Title: info.Find("h3").Text(),
		Author: []string{
			author,
		},
		Summary:     TrimHtml(info.Find("p").Eq(5).Text()),
		Description: TrimHtml(info.Find("p").Eq(5).Text()),
		Subtitle:    info.Find(".bookname").Text(),
		Version:     "2",
		Itunes:      "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Link:        siteUrl,
		Language:    "zh-cn",
		Image: Image{
			Href: image,
		},
		PubDate: t.Format(time.RFC1123),
		Owner: Owner{
			Name:  owner,
			Email: "LSvKing@Gmail.com",
		},
	}


	header := req.Header{
		"Cookie":        "safedog-flow-item=; lygusername=lsvking; userid=427591; ASPSESSIONIDQSTSTBCS=FKCGENOCPGGAJJCPNGEAFMIH; apwd=lsv324000; userid=427591; aname=lsvking; lyguserpwd=lsv324000; hisArt=%5B%7B%22title%22%3A%22%E5%A4%A9%E4%BD%93%E6%82%AC%E6%B5%AE%22%2C%22url%22%3A%22%2Fbook23549%22%7D%2C%7B%22title%22%3A%22%E5%87%AF%E5%8F%94%E8%A5%BF%E6%B8%B8%E8%AE%B0_1-5%E9%83%A8%E5%85%A8%E9%9B%86%22%2C%22url%22%3A%22%2Fbook23536%22%7D%2C%7B%22title%22%3A%22%E5%86%92%E6%AD%BB%E8%AE%B0%E5%BD%95%E7%A5%9E%E7%A7%98%E4%BA%8B%E4%BB%B64_%E9%9D%92%E9%9B%AA%E6%95%85%E4%BA%8B%22%2C%22url%22%3A%22%2Fbook22737%22%7D%2C%7B%22title%22%3A%22undefined%22%2C%22url%22%3A%22undefined%22%7D%2C%7B%22title%22%3A%22%E6%9D%91%E4%B8%8A%E6%98%A5%E6%A0%91_1Q84%22%2C%22url%22%3A%22%2Fbook23556%22%7D%2C%7B%22title%22%3A%22%E8%8B%8F%E9%BA%BB%E5%96%87%E5%A7%91%E4%BC%A0%22%2C%22url%22%3A%22%2Fbook23543%22%7D%2C%7B%22title%22%3A%22%E5%88%9D%E4%B8%AD%E7%94%9F%E5%BF%85%E8%83%8C%E5%8F%A4%E8%AF%97%E6%96%87%E6%A0%87%E5%87%86%E6%9C%97%E8%AF%B5%22%2C%22url%22%3A%22%2Fbook23513%22%7D%5D",
	}

	param := req.Param{
		"uname": "lsvking",
	}

	fmt.Println(pageNum)

	var items []Item

	for i:=1;i < pageNum;i++{

		siteUrl := siteUrl + "p" + strconv.Itoa(i)

		resp, err := req.Get(siteUrl)

		if err != nil {
			fmt.Println(err.Error)
		}


		if err != nil || resp.Response().StatusCode == http.StatusNotFound {
			return []byte("资源不存在")
		}

		doc, _ := goquery.NewDocumentFromResponse(resp.Response())

		doc.Find("#sortedList li").Each(func(i int, selection *goquery.Selection) {
			aid := selection.Find("span").Eq(0).AttrOr("kv","null")

			title:=  selection.Find("span").Eq(0).AttrOr("kt","null")

			param["aid"] = aid

			resp,err := req.Post(mp3Url,header,param)

			if err != nil {
				fmt.Println(err)
			}


			items = append(items, Item{
				Title:    title,
				Subtitle: title,
				Author:   author,
				PubDate:  t.Format(time.RFC1123),
				Summary:  "",
				Guid: Guid{
					IsPermaLink: "true",
				},
				Image: Image{
					Href: image,
				},
				Enclosure: Enclosure{
					Url:  mp3Prefix + resp.String(),
					Type: "audio/mpeg",
				},
				//Duration: ,
			})

		})
	}


	rss.Item = items
	output, err := xml.MarshalIndent(rss, "  ", "    ")

	o := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n" + string(output))

	//cache.Set("pingshu-"+id+"|"+Time, o)

	return o
}
