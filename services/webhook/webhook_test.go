package webhook

import (
	"testing"
)

type MockStore struct {
	tasks []WebhookTask
}

func (m *MockStore) Save(tasks []WebhookTask) error {
	m.tasks = tasks
	return nil
}

func (m *MockStore) LoadPending() ([]WebhookTask, error) {
	return m.tasks, nil
}

func TestWebhookQueueGracefulShutdown(t *testing.T) {
	store := &MockStore{}
	queue := NewWebhookQueue(store, 10)

	queue.Start(2)

	task1 := WebhookTask{ID: "1", URL: "http://localhost:8080", State: "pending"}
	task2 := WebhookTask{ID: "2", URL: "http://localhost:8080", State: "pending"}

	queue.Push(task1)
	queue.Push(task2)

	queue.Shutdown()

	pending, _ := store.LoadPending()
	if len(pending) == 0 {
		t.Errorf("Expected tasks to be persisted, got 0")
	}
}
