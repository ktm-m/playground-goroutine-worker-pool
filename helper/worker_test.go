package helper_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ktm-m/playground-goroutine-worker-pool/helper"
)

func TestSubmitAndProcessJobs(t *testing.T) {
	pool := helper.NewWorkerPool(&helper.WorkerPoolConfig{
		NumWorkers:         5, // Small number for testing
		MaxQueueMultiplier: 2,
	})

	var counter int32 = 0
	numJobs := 20

	// Submit jobs
	for i := 0; i < numJobs; i++ {
		pool.SubmitJob(func(ctx context.Context) {
			atomic.AddInt32(&counter, 1)
			time.Sleep(10 * time.Millisecond) // Simulate work
		})
	}

	// Allow jobs to process
	time.Sleep(500 * time.Millisecond)
	pool.GracefullyShutdown()

	if int(atomic.LoadInt32(&counter)) != numJobs {
		t.Fatalf("Expected %d jobs to complete, but got %d", numJobs, counter)
	}

	t.Log("Expected jobs completed:", numJobs)
	t.Log("Num jobs completed:", counter)
}

func TestQueueFullRetry(t *testing.T) {
	pool := helper.NewWorkerPool(&helper.WorkerPoolConfig{
		NumWorkers:         1,
		MaxQueueMultiplier: 1, // Very small queue (1 job)
		SleepDuration:      10 * time.Millisecond,
	})

	// Create a slow job to fill the queue
	slowJobDone := make(chan struct{})
	pool.SubmitJob(func(ctx context.Context) {
		time.Sleep(200 * time.Millisecond) // Block queue for a while
		close(slowJobDone)
	})

	var jobsCompleted int32 = 0

	// Submit another job that will need to retry
	go func() {
		pool.SubmitJob(func(ctx context.Context) {
			atomic.AddInt32(&jobsCompleted, 1)
		})
	}()

	// Wait for first job to finish
	<-slowJobDone

	// Allow second job to process
	time.Sleep(100 * time.Millisecond)
	pool.GracefullyShutdown()

	if atomic.LoadInt32(&jobsCompleted) != 1 {
		t.Fatal("Second job should have eventually completed after retrying")
	}
}
