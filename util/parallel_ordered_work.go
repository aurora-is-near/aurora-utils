package util

import (
	"context"
	"sync"
)

func ProcessInParallel[A any, B any](
	ctx context.Context,
	input <-chan A,
	output chan<- B,
	processFn func(A) B,
	workersCount int,
	bufferSize int, // at least 100 (recommended something like 512)
) error {

	type internalJob struct {
		in   A
		out  B
		done chan struct{}
	}

	var wg sync.WaitGroup
	internalInput := make(chan *internalJob, bufferSize)
	internalOutput := make(chan *internalJob, bufferSize)

	// Inputting
	inputFullyDone := false
	wg.Go(func() {
		defer close(internalInput)
		defer close(internalOutput)

		it := NewChanIterator(input)
		for in := range it.Iterate(ctx) {
			job := &internalJob{
				in:   in,
				done: make(chan struct{}),
			}
			// Writing to input first, because workers are not blocked on outputter
			select {
			case <-ctx.Done():
				return
			case internalInput <- job:
			}
			// Only then writing to output, because outputter can block on workers
			select {
			case <-ctx.Done():
				return
			case internalOutput <- job:
			}
		}
		inputFullyDone = it.Closed()
	})

	// Workers
	for range workersCount {
		wg.Go(func() {
			for job := range NewChanIterator(internalInput).Iterate(ctx) {
				job.out = processFn(job.in)
				close(job.done) // Notifying outputter that processing is done
			}
		})
	}

	// Outputting
	outputFullyDone := false
	wg.Go(func() {
		it := NewChanIterator(internalOutput)
		for job := range it.Iterate(ctx) {
			// Waiting for job to be processed
			select {
			case <-ctx.Done():
				return
			case <-job.done:
			}
			// Writing to output
			select {
			case <-ctx.Done():
				return
			case output <- job.out:
			}
		}
		outputFullyDone = it.Closed()
	})

	wg.Wait()
	if inputFullyDone && outputFullyDone {
		return nil
	}
	return ctx.Err()
}
