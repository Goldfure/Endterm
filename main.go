package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sort"
	"sync"
	"time"
)

type record struct {
	word    []byte
	counter int
}
type job func(in, out chan interface{})

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 10)
	out := make(chan interface{}, 10)

	for i := 0; i < len(jobs); i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, in, out chan interface{}, j job) {
			j(in, out)
			close(out)
			wg.Done()
		}(wg, in, out, jobs[i])
		in = out
		out = make(chan interface{}, 77)
	}
	wg.Wait()
}

func producer(out chan interface{}, file []byte, readingBuf []byte) {
	var bytes []byte
	for _, v := range file {
		if (v >= 65 && v <= 90) || (v >= 97 && v <= 122) {
			if v >= 65 && v <= 90 {
				v += 32
			}
			bytes = append(bytes, v)
		} else if len(bytes) != 0 {
			out <- bytes
			bytes = []byte{}
		}
	}
}

func consumer(in, out chan interface{}) {
	var arr []record
	for val := range in {
		if len(arr) == 0 {
			arr = append(arr, record{word: val.([]byte), counter: 1})
		} else {
			check := false
			for index, v := range arr {
				if bytes.Equal(v.word, val.([]byte)) {
					arr[index].counter = arr[index].counter + 1
					check = true
					break
				}
			}
			if !check {
				arr = append(arr, record{word: val.([]byte), counter: 1})
			}
		}
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].counter > arr[j].counter
	})
	for i := 0; i < 20; i++ {
		out <- arr[i]
	}
}

func main() {
	start := time.Now()
	file, err := ioutil.ReadFile("mobydick.txt")
	if err != nil {
		panic(err)
		return
	}
	readingBuf := make([]byte, 1)

	jobs := []job{
		job(func(in, out chan interface{}) {
			producer(out, file, readingBuf)
		}),
		job(func(in, out chan interface{}) {
			consumer(in, out)
		}),
		job(func(in, out chan interface{}) {
			for val := range in {
				fmt.Printf("%v %v\n", val.(record).counter, string(val.(record).word))
			}
		}),
	}

	ExecutePipeline(jobs...)

	fmt.Printf("Process took %s\n", time.Since(start))
}
