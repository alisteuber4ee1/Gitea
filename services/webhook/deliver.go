package webhook

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

type WebhookTask struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Payload   string    `json:"payload"`
	State     string    `json:"state"` // "pending", "processing", "completed", "failed"
	CreatedAt time.Time `json:"created_at"`
}

type Store interface {
	Save(tasks []WebhookTask) error
	LoadPending() ([]WebhookTask, error)
}

type FileStore struct {
	filePath string
	mu       sync.Mutex
}

func NewFileStore(filePath string) *FileStore {
	return &FileStore{filePath: filePath}
}

func (fs *FileStore) Save(tasks []WebhookTask) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fs.filePath, data, 0644)
}

func (fs *FileStore) LoadPending() ([]WebhookTask, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		return nil, nil
	}
	data, err := ioutil.ReadFile(fs.filePath)
	if err != nil {
		return nil, err
	}
	var tasks []WebhookTask
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	var pending []WebhookTask
	for _, t := range tasks {
		if t.State == "pending" {
			pending = append(pending, t)
		}
	}
	return pending, nil
}
