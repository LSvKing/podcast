package cache

import (
	"fmt"

	"time"

	"github.com/adelowo/onecache"
	"github.com/adelowo/onecache/filesystem"
)

var store onecache.Store

func init() {
	store = filesystem.MustNewFSStore("data/cache", time.Hour*24*30)
}

func Get(key string) ([]byte, error) {

	if body, err := store.Get(key); err != nil {
		return []byte(""), err
	} else {
		return body, err
	}
}

func Set(key string, body []byte) {

	err := store.Set(key, body, time.Hour*24*30)

	if err != nil {
		fmt.Println(err)
	}

}
