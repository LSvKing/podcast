package crawler

import (
	"context"
	"log"

	"fmt"

	"cloud.google.com/go/datastore"
)

type Task struct {
	Description string
}

func Ds(id string) []byte {

	ctx := context.Background()

	// Set your Google Cloud Platform project ID.
	projectID := "lsvproxyv1"

	// Creates a client.
	client, err := datastore.NewClient(ctx, projectID)

	if err != nil {
		fmt.Println("Failed to create client:", err)
	}

	// Sets the kind for the new entity.
	kind := "Task"
	// Sets the name/ID for the new entity.
	name := "id-" + id
	// Creates a Key instance.
	taskKey := datastore.NameKey(kind, name, nil)

	// Creates a Task instance.
	task := Task{
		Description: "ID: is" + id,
	}

	// Saves the new entity.
	if _, err := client.Put(ctx, taskKey, &task); err != nil {
		log.Fatalf("Failed to save task: %v", err)
	}

	fmt.Printf("Saved %v: %v\n", taskKey, task.Description)

	var t Task

	k := datastore.NameKey(kind, name, nil)

	client.Get(ctx, k, &t)

	return []byte(t.Description)
}
