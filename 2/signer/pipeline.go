package main

import "sync"

func createChannels(numJobs int, maxInputs int) []chan interface{} {
	channels := make([]chan interface{}, 0, numJobs+1)

	channels = append(channels, make(chan interface{}))
	for i := 1; i < numJobs; i++ {
		channels = append(channels, make(chan interface{}, maxInputs))
	}
	channels = append(channels, make(chan interface{}))

	return channels
}

func ExecutePipeline(jobs ...job) {
	maxInputs := 100
	channels := createChannels(len(jobs), maxInputs)

	wg := &sync.WaitGroup{}

	for idx := range jobs {
		wg.Add(1)
		go func(function job, in, out chan interface{}) {
			function(in, out)
			close(out)
			wg.Done()
		}(jobs[idx], channels[idx], channels[idx+1])
	}

	wg.Wait()
}
