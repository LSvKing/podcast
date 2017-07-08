package cache

import (
	"time"

	"github.com/adelowo/onecache"
	"github.com/adelowo/onecache/filesystem"
)

func New() onecache.Store {
	store := filesystem.MustNewFSStore("data/cache", 10*time.Hour)
	//store, err := onecache.Get("fs")
	//
	//if err != nil {
	//	log.Panic(err)
	//}

	return store
}
