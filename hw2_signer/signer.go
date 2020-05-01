package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	//"runtime"
	"strconv"
	"sync"
	//"time"
)

func ExecutePipeline(jobs ...job) {
	wgm := &sync.WaitGroup{}
	wgm.Add(len(jobs))

	var out chan interface{}
	in := make(chan interface{}, 100)

	for _, fn := range jobs {
		out = make(chan interface{}, 100)

		go func(f job, wgm *sync.WaitGroup, in, out chan interface{}) {
			defer wgm.Done()
			defer close(out)
			fmt.Println("run")
			f(in, out)
		}(fn, wgm, in, out)

		in = out
	}
	wgm.Wait()
}

var SingleHash = func(in, out chan interface{}) {
	for dataRaw := range in {
		//dataRaw := <-in
		data := strconv.Itoa(dataRaw.(int))

		md5chan := make(chan string)
		crc32md5chan := make(chan string)
		crc32chan := make(chan string)
		go func(md5chan,
			crc32md5chan,
			crc32chan chan string) {
			go putCrc32ToChan(crc32chan, data)
			go putMd5ToChan(md5chan, data)
			go putCrc32ToChan(crc32md5chan, <-md5chan)
		}(md5chan,
			crc32md5chan,
			crc32chan)

		out <- <-crc32chan + "~" + <-crc32md5chan
	}
}
var MultiHash = func(in, out chan interface{}) {
	for dataRaw := range in {

		//dataRaw := <-in
		data, ok := dataRaw.(string)
		if !ok {
			fmt.Printf("cant convert result data to string")
		}

		const count = 5
		var counters = map[int]string{}
		mu := &sync.Mutex{}
		wg := &sync.WaitGroup{}

		for i := 0; i <= count; i++ {
			wg.Add(1)
			go putMultihashPartToMap(counters, i, data, mu, wg)
		}

		wg.Wait()

		out <- getMultihash(counters, count)
	}
}

//
func getMultihash(counters map[int]string, count int) string {
	var res string
	for i := 0; i <= count; i++ {
		res += counters[i]
	}
	return res
}

func putMultihashPartToMap(counters map[int]string, ind int, data string, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	crc32 := DataSignerCrc32(strconv.Itoa(ind) + data)
	mu.Lock()
	counters[ind] = crc32
	mu.Unlock()
}

func putCrc32ToChan(ch chan string, data string) {
	ch <- DataSignerCrc32(data)
}

func putMd5ToChan(ch chan string, data string) {
	ch <- DataSignerMd5(data)
}

var CombineResults = func(in, out chan interface{}) {
	var all []string
	start := time.Now()
	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			fmt.Printf("cant convert result data to string CombineResults")
		}
		all = append(all, data)
	}
	end := time.Since(start)
	sort.Slice(all, func(i, j int) bool {
		return all[i] < all[j]
	})
	out <- strings.Join(all, "_")
	fmt.Println("exec time multihash", end)
}
