package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ktm-m/playground-goroutine-worker-pool/helper"
)

type AsyncServicePort interface {
	RunAsyncJob()
}

type asyncService struct {
	async      helper.AsyncPort
	workerPool helper.WorkerPoolPort
}

func NewAsyncService(async helper.AsyncPort, workerPool helper.WorkerPoolPort) AsyncServicePort {
	return &asyncService{
		async:      async,
		workerPool: workerPool,
	}
}

func (a *asyncService) RunAsyncJob() {
	var (
		wg                                     sync.WaitGroup
		firstJobErr, secondJobErr, thirdJobErr error
	)

	wg.Add(3)
	a.workerPool.SubmitJob(func(ctx context.Context) {
		a.async.RunAsyncJob(func() {
			fmt.Println("running first job")
			time.Sleep(1 * time.Second)
			fmt.Println("first job done")
		}, &wg, &firstJobErr)
	})

	a.workerPool.SubmitJob(func(ctx context.Context) {
		a.async.RunAsyncJob(func() {
			fmt.Println("running second job")
			time.Sleep(1 * time.Second)
			fmt.Println("second job done")
		}, &wg, &secondJobErr)
	})

	a.workerPool.SubmitJob(func(ctx context.Context) {
		a.async.RunAsyncJob(func() {
			fmt.Println("running third job")
			time.Sleep(1 * time.Second)
			fmt.Println("third job done")
		}, &wg, &thirdJobErr)
	})
	wg.Wait()
}
