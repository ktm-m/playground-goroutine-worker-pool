package main

import (
	"github.com/ktm-m/playground-goroutine-worker-pool/config"
	"github.com/ktm-m/playground-goroutine-worker-pool/helper"
	"github.com/ktm-m/playground-goroutine-worker-pool/service"
)

func main() {
	appConfig := config.GetConfig()

	workerPool := helper.NewWorkerPool(&helper.WorkerPoolConfig{
		NumWorkers:    appConfig.WorkerPool.NumWorker,
		SleepDuration: appConfig.WorkerPool.SleepDuration,
	})
	async := helper.NewAsync()

	asyncService := service.NewAsyncService(async, workerPool)
	testAsyncService(asyncService)

	workerPool.GracefullyShutdown()
}

func testAsyncService(asyncService service.AsyncServicePort) {
	for i := 0; i < 100000; i++ {
		asyncService.RunAsyncJob()
	}
}
