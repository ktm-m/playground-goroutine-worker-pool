package helper_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ktm-m/playground-goroutine-worker-pool/helper"
)

func TestWorkerPool_SubmitJobAndGracefulShutdown(t *testing.T) {
	var processedCount int32 = 0
	numJobs := 50
	numWorkers := 10

	wp := helper.NewWorkerPool(&helper.WorkerPoolConfig{
		NumWorkers:    numWorkers,
		SleepDuration: 5 * time.Millisecond,
	})

	for i := 0; i < numJobs; i++ {
		jobNum := i
		wp.SubmitJob(func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond) // simulate some work
			fmt.Printf("Job %d done\n", jobNum)
			atomic.AddInt32(&processedCount, 1)
		})
	}

	wp.GracefullyShutdown()

	if processedCount != int32(numJobs) {
		t.Errorf("Expected %d jobs processed, but got %d", numJobs, processedCount)
	}
}

func TestWorkerPool_SubmitJob_WaitsWhenQueueIsFull(t *testing.T) {
	wp := helper.NewWorkerPool(&helper.WorkerPoolConfig{
		NumWorkers:    1, // worker เดียว
		SleepDuration: 100 * time.Millisecond,
	})

	// Job แรกใช้เวลานานเพื่อบล็อค queue
	wp.SubmitJob(func(ctx context.Context) {
		time.Sleep(500 * time.Millisecond)
	})

	start := time.Now()

	// Job ที่สองจะเข้า queue ไม่ได้ทันที (queue size = 1, job แรกยังไม่เสร็จ)
	wp.SubmitJob(func(ctx context.Context) {
		// no-op
	})

	duration := time.Since(start)

	if duration < 90*time.Millisecond {
		t.Errorf("Expected SubmitJob to wait due to full queue, but it returned too quickly: %v", duration)
	} else {
		t.Logf("SubmitJob waited as expected: %v", duration)
	}

	wp.GracefullyShutdown()
}
