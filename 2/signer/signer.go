package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	md5Permits := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}

	for input := range in {
		inputInt, ok := input.(int)
		if !ok {
			panic("encountered non-int data")
		}

		inputString := strconv.Itoa(inputInt)

		wg.Add(1)
		go singleHashItem(wg, inputString, out, md5Permits)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for input := range in {
		inputString, ok := input.(string)
		if !ok {
			panic("encountered non-string data")
		}

		wg.Add(1)
		go multiHashItem(wg, inputString, out)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	inputs := make([]string, 0)
	for input := range in {
		inputString, ok := input.(string)
		if !ok {
			panic("encountered non-string data")
		}

		inputs = append(inputs, inputString)
	}

	sort.Strings(inputs)

	result := strings.Builder{}
	for idx, input := range inputs {
		result.WriteString(input)
		if idx != len(inputs)-1 {
			result.WriteString("_")
		}
	}

	out <- result.String()
}

func singleHashItem(
	wg *sync.WaitGroup,
	item string,
	out chan interface{},
	md5Permits chan struct{},
) {
	defer wg.Done()

	md5C := deferredMd5(item, md5Permits)
	crc32C := deferredCrc32(item)

	md5Result := <-md5C

	crc32Md5C := deferredCrc32(md5Result)

	result := <-crc32C + "~" + <-crc32Md5C
	out <- result
}

func deferredCrc32(data string) chan string {
	result := make(chan string, 1)
	go func() {
		result <- DataSignerCrc32(data)
	}()
	return result
}

func deferredMd5(data string, permits chan struct{}) chan string {
	result := make(chan string, 1)
	go func() {
		permits <- struct{}{}
		result <- DataSignerMd5(data)
		<-permits
	}()
	return result
}

func multiHashItem(
	wg *sync.WaitGroup,
	item string,
	out chan interface{},
) {
	defer wg.Done()

	itemWg := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	hashes := make([]string, 6)

	for i := 0; i < 6; i++ {
		itemWg.Add(1)

		go func(idx int, item string) {
			defer itemWg.Done()

			crc32Result := <-deferredCrc32(strconv.Itoa(idx) + item)

			mutex.Lock()
			hashes[idx] = crc32Result
			mutex.Unlock()
		}(i, item)
	}

	itemWg.Wait()

	result := strings.Builder{}
	for _, hash := range hashes {
		result.WriteString(hash)
	}

	out <- result.String()
}
