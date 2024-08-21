package util

import (
	"sync"
)

type Task func()

type WorkerPool struct {
	taskQueue chan Task
	noWorkers int
	wg        sync.WaitGroup
}

func NewWorkerPool(noWorkers int) *WorkerPool {
	return &WorkerPool{
		taskQueue: make(chan Task),
		noWorkers: noWorkers,
	}
}

func (pool *WorkerPool) Start() {
	for i := 0; i < pool.noWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}
}

func (pool *WorkerPool) worker(id int) {
	defer pool.wg.Done()
	for task := range pool.taskQueue {
		task()
	}
}

func (pool *WorkerPool) Submit(task Task) {
	pool.taskQueue <- task
}

func (pool *WorkerPool) Stop() {
	close(pool.taskQueue)
	pool.wg.Wait()
}
