package main

import (
	"fmt"
	"time"

	"code.gitea.io/gitea/modules/graceful"
	"code.gitea.io/gitea/services/webhook"
)

func main() {
	fmt.Println("Starting Gitea Webhook Service...")
	store := webhook.NewFileStore("webhook_tasks.json")
	queue := webhook.NewWebhookQueue(store, 10)
	queue.Start(2)

	queue.Push(webhook.WebhookTask{
		ID:        "task-1",
		URL:       "https://httpbin.org/delay/1",
		Payload:   `{"event": "push"}`,
		State:     "pending",
		CreatedAt: time.Now(),
	})

	time.Sleep(500 * time.Millisecond)
	queue.Shutdown()
	fmt.Println("Gitea Webhook Service stopped.")
}
