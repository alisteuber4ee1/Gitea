package graceful

import (
	"context"
	"sync"
)

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var (
	manager     *Manager
	managerOnce sync.Once
)

func GetManager() *Manager {
	managerOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		manager = &Manager{
			ctx:    ctx,
			cancel: cancel,
		}
	})
	return manager
}

func (m *Manager) ShutdownContext() context.Context {
	return m.ctx
}

func (m *Manager) RegisterTask() {
	m.wg.Add(1)
}

func (m *Manager) DoneTask() {
	m.wg.Done()
}

func (m *Manager) DoGracefulShutdown() {
	m.cancel()
	m.wg.Wait()
}
