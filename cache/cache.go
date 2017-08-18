package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adelowo/onecache"
	"github.com/adelowo/onecache/filesystem"

	"cloud.google.com/go/datastore"
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

type Cache struct {
	Key  string
	Body []byte
}

func Get(key string) ([]byte, error) {
	ctx := context.Background()

	// Set your Google Cloud Platform project ID.
	projectID := "lsvproxyv1"

	// Creates a client.
	client, err := datastore.NewClient(ctx, projectID)

	if err != nil {
		fmt.Println("Failed to create client:", err)
	}

	// Sets the kind for the new entity.
	kind := "Podcast"
	// Sets the name/ID for the new entity.
	//name := key

	var c Cache

	k := datastore.NameKey(kind, key, nil)

	if err := client.Get(ctx, k, &c); err != nil {
		return []byte(""), err
	} else {
		return c.Body, err
	}
}

func Set(key string, body []byte) {
	ctx := context.Background()

	// Set your Google Cloud Platform project ID.
	projectID := "lsvproxyv1"

	// Creates a client.
	client, err := datastore.NewClient(ctx, projectID)

	if err != nil {
		fmt.Println("Failed to create client:", err)
	}

	kind := "Podcast"

	// Creates a Key instance.
	k := datastore.NameKey(kind, key, nil)

	c := Cache{
		Key:  key,
		Body: body,
	}

	// k
	if _, err := client.Put(ctx, k, &c); err != nil {
		log.Fatalf("Failed to save task: %v", err)
	}

	log.Printf("Saved %v: %v\n", k, c.Body)

}
