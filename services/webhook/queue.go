package webhook

import (
	"net/http"
	"sync"
	"time"

	"code.gitea.io/gitea/modules/graceful"
)

type WebhookQueue struct {
	store      Store
	queue      chan WebhookTask
	workerWg   sync.WaitGroup
	httpClient *http.Client
}

func NewWebhookQueue(store Store, bufferSize int) *WebhookQueue {
	return &WebhookQueue{
		store:      store,
		queue:      make(chan WebhookTask, bufferSize),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (wq *WebhookQueue) Start(numWorkers int) {
	pendingTasks, err := wq.store.LoadPending()
	if err == nil {
		for _, task := range pendingTasks {
			select {
			case wq.queue <- task:
			default:
			}
		}
	}

	for i := 0; i < numWorkers; i++ {
		wq.workerWg.Add(1)
		go wq.workerLoop()
	}
}

func (wq *WebhookQueue) workerLoop() {
	defer wq.workerWg.Done()
	manager := graceful.GetManager()
	shutdownCtx := manager.ShutdownContext()

	for {
		select {
		case <-shutdownCtx.Done():
			return
		case task, ok := <-wq.queue:
			if !ok {
				return
			}
			manager.RegisterTask()
			wq.deliver(task)
			manager.DoneTask()
		}
	}
}

func (wq *WebhookQueue) deliver(task WebhookTask) {
	task.State = "processing"
	req, err := http.NewRequestWithContext(graceful.GetManager().ShutdownContext(), "POST", task.URL, nil)
	if err != nil {
		task.State = "failed"
		return
	}
	resp, err := wq.httpClient.Do(req)
	if err != nil {
		task.State = "failed"
	} else {
		resp.Body.Close()
		task.State = "completed"
	}
}

func (wq *WebhookQueue) Push(task WebhookTask) bool {
	select {
	case <-graceful.GetManager().ShutdownContext().Done():
		return false
	case wq.queue <- task:
		return true
	}
}

func (wq *WebhookQueue) Shutdown() {
	graceful.GetManager().DoGracefulShutdown()

	close(wq.queue)
	wq.workerWg.Wait()

	var remainingTasks []WebhookTask
	for task := range wq.queue {
		task.State = "pending"
		remainingTasks = append(remainingTasks, task)
	}

	if len(remainingTasks) > 0 {
		_ = wq.store.Save(remainingTasks)
	} else {
		_ = wq.store.Save(nil)
	}
}
