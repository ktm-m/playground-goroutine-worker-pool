package helper

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type WorkerPoolPort interface {
	SubmitJob(job Job)
	GracefullyShutdown()
}

var (
	defaultNumWorker     = 100
	defaultSleepDuration = 10 * time.Millisecond
)

type Job func(ctx context.Context)

type WorkerPoolConfig struct {
	NumWorkers    int
	SleepDuration time.Duration
}

type workerPool struct {
	numWorker     int
	sleepDuration time.Duration
	jobQueue      chan Job
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewWorkerPool(workerPoolConfig *WorkerPoolConfig) WorkerPoolPort {
	ctx, cancel := context.WithCancel(context.Background())

	workerPoolInstance := &workerPool{
		numWorker:     defaultNumWorker,
		sleepDuration: defaultSleepDuration * time.Millisecond,
		jobQueue:      make(chan Job, 1*defaultNumWorker),
		ctx:           ctx,
		cancel:        cancel,
	}

	workerPoolInstance.setWorkerPoolConfig(workerPoolConfig)
	workerPoolInstance.start()

	return workerPoolInstance
}

func (w *workerPool) setWorkerPoolConfig(workerPoolConfig *WorkerPoolConfig) {
	if workerPoolConfig.NumWorkers > 0 {
		w.numWorker = workerPoolConfig.NumWorkers
	}

	if workerPoolConfig.SleepDuration > 0 {
		w.sleepDuration = workerPoolConfig.SleepDuration
	}
}

func (w *workerPool) start() {
	for i := 0; i < w.numWorker; i++ {
		go w.workerLoop()
	}
}

func (w *workerPool) workerLoop() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case job := <-w.jobQueue:
			func() {
				defer w.wg.Done()
				defer func() {
					if r := recover(); r != nil {
					}
				}()
				job(w.ctx)
			}()
		}
	}
}

func (w *workerPool) SubmitJob(job Job) {
	w.wg.Add(1)

	for {
		select {
		case <-w.ctx.Done():
			w.wg.Done()
			return
		case w.jobQueue <- job:
			return
		default:
			fmt.Println("job queue is full, waiting to retry...")
			time.Sleep(w.sleepDuration)
			w.jobQueue <- job
		}
	}
}

func (w *workerPool) GracefullyShutdown() {
	w.wg.Wait()
	w.cancel()
}
