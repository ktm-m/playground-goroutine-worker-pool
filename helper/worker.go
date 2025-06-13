package helper

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/samber/lo"
)

type WorkerPoolPort interface {
	SubmitJob(job Job)
	GracefullyShutdown()
}

var (
	defaultNumWorker          = 100
	defaultMaxQueueMultiplier = 10
	defaultSleepDuration      = 10 * time.Millisecond
)

type Job func(ctx context.Context)

type WorkerPoolConfig struct {
	NumWorkers         int
	MaxQueueMultiplier int
	SleepDuration      time.Duration
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

	numWorker := lo.Ternary(workerPoolConfig.NumWorkers > 0, workerPoolConfig.NumWorkers, defaultNumWorker)
	workerPoolInstance := &workerPool{
		numWorker:     numWorker,
		sleepDuration: lo.Ternary(workerPoolConfig.SleepDuration > 0, workerPoolConfig.SleepDuration, defaultSleepDuration),
		ctx:           ctx,
		cancel:        cancel,
	}

	maxQueueMultiplier := lo.Ternary(workerPoolConfig.MaxQueueMultiplier > 0, workerPoolConfig.MaxQueueMultiplier, defaultMaxQueueMultiplier)
	jobQueue := make(chan Job, numWorker*maxQueueMultiplier)

	workerPoolInstance.jobQueue = jobQueue
	workerPoolInstance.start()

	return workerPoolInstance
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
						log.Println("go routine is panic:", r)
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
			time.Sleep(w.sleepDuration)
		}
	}
}

func (w *workerPool) GracefullyShutdown() {
	w.cancel()
	w.wg.Wait()
}
