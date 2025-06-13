package helper

import (
	"fmt"
	"sync"
)

type AsyncPort interface {
	RunAsyncJob(fn func(), wg *sync.WaitGroup, err *error)
}

type async struct{}

func NewAsync() AsyncPort {
	return &async{}
}

func (a *async) RunAsyncJob(fn func(), wg *sync.WaitGroup, err *error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				*err = fmt.Errorf("go routine is panic: %v", r)
			}
			wg.Done()
		}()
		fn()
	}()
}
