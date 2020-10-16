package worker

import (
	"fmt"
	"strings"

	"github.com/go-kit/kit/log"

	"github.com/qubole/edith/internal/logger"
	"github.com/qubole/edith/pkg/command"
)

// PushInterface to enqueue to redis
type PushInterface interface {
	Push(command.Command)
}

// Codec manages pool of workers
type Manager struct {
	workers []*Worker
}

// Option to set params.
type Option func(*Worker)

// NewManager is constructor for workers.
func NewManager(servers []string, cmdManager command.Codec,options ...Option) *Manager {
	m := &Manager{}

	// apply
	for idx, s := range servers {
		w := newWorker(idx, s, cmdManager, options...)
		m.workers = append(m.workers, w)
	}

	return m
}

// Logger sets logger.
// labels are key, val pair .. so even in number always.
func Logger(level string, labels ...interface{}) Option {
	return func(w *Worker) {
		w.logger = logger.Create(level)

		labels := append(labels, "worker_id", fmt.Sprintf("%d", w.id))
		w.logger = log.With(w.logger, labels...)
	}
}

// CmdQueue sets cmdQueue.
func CmdQueue(queue string) Option {
	return func(w *Worker) {
		w.cmdQueue = strings.ToLower(queue) + fmt.Sprintf("_%d", w.id)
		w.cmdQueueBackup = "_" + w.cmdQueue + "_backup_"
	}
}

// CmdQueueTimeout sets cmdQueueTimeout.
func CmdQueueTimeout(timeout int64) Option {
	return func(w *Worker) {
		w.cmdQueueTimeOut = timeout
	}
}

// RunFn type to encapsulate a goroutine.
type RunFn func(<-chan struct{}) error

// Runnables returns all goroutines parallel in parallel for all workers.
func (m *Manager) Runnables() []RunFn {
	var fs []RunFn

	for _, worker := range m.workers {
		w := worker
		fs = append(fs, func(stop <-chan struct{}) error {
			w.loopPushCmd(stop)
			return fmt.Errorf("worker.%d.loopPushCmd.interrupted", w.id)
		})
		fs = append(fs, func(stop <-chan struct{}) error {
			_ = w.loopProcessCmd(stop)
			return fmt.Errorf("worker.%d.loopProcessCmd.interrupted", w.id)
		})
	}
	return fs
}

// Push a command to worker.
func (m *Manager) Push(cmd command.Command) {
	idx := cmd.GetID() % uint64(len(m.workers))
	m.workers[idx].push(cmd)
}
