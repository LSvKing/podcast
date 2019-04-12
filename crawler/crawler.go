package crawler

import (
	"regexp"
	"strings"
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
	Subtitle  string    `xml:"itunes:subtitle" json:"subtitle"`
	Author    string    `xml:"itunes:author" json:"author"`
	PubDate   string    `xml:"pubDate"`
	Summary   string    `xml:"itunes:summary"`
	Guid      Guid      `xml:"guid"`
	Image     Image     `xml:"itunes:image"`
	Enclosure Enclosure `xml:"enclosure"`
	Duration  int64     `xml:"itunes:duration" json:"duration"`
}

const rfc2822 = "Mon, 2 Jan 2006 15:04:05 UTC+8"

func TrimHtml(str string) string {
	//将HTML标签全转换成小写
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	str = re.ReplaceAllStringFunc(str, strings.ToLower)

	//去除STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	str = re.ReplaceAllString(str, "")

	//去除SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	str = re.ReplaceAllString(str, "")

	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	str = re.ReplaceAllString(str, "\n")

	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	str = re.ReplaceAllString(str, "\n")

	return strings.TrimSpace(str)
}
