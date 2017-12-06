package main

import (
	"log"
	"net/http"

	"podcast/crawler"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", HomeHandler)

	r.HandleFunc("/feed/{type}/{id}.xml", FeedHandler)

	r.Headers("Content-Type", "application/xml")

	log.Fatal(http.ListenAndServe("localhost:8088", r))
	log.Fatal(http.ListenAndServe("podcast.lsvking.com:80", r))
	//crawler.Ximalaya("2684111")
}

//HomeHandler Home
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Podcast"))
}

//FeedHandler Feed
func FeedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var output []byte

	switch vars["type"] {
	case "qingting":
		output = crawler.Qingting(vars["id"])
	case "ximalaya":
		output = crawler.Ximalaya(vars["id"])
	case "pingshu8":
		output = crawler.PingShu8(vars["id"])
	case "ivoix":
		output = crawler.Ivoix(vars["id"])
	default:
		output = []byte("类型不存在")

		//output = crawler.PingShu8(vars["id"])
	}

	w.Write(output)
	w.WriteHeader(http.StatusOK)
}
