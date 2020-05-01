package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
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
	wg := &sync.WaitGroup{}
	for dataRaw := range in {
		data := strconv.Itoa(dataRaw.(int))
		md5 := DataSignerMd5(data)
		wg.Add(1)
		go func(md5, data string, group *sync.WaitGroup, out chan interface{}) {
			defer wg.Done()
			crc32md5chan := make(chan string)
			crc32chan := make(chan string)
			go putCrc32ToChan(crc32chan, data)
			go putCrc32ToChan(crc32md5chan, md5)
			out <- <-crc32chan + "~" + <-crc32md5chan
		}(md5, data, wg, out)
	}
	wg.Wait()
}
var MultiHash = func(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for dataRaw := range in {
		data := dataRaw.(string)
		const count = 5
		wg.Add(1)
		go func(out chan interface{}) {
			defer wg.Done()
			var counters = map[int]string{}
			mu := &sync.Mutex{}
			wgLoc := &sync.WaitGroup{}

			for i := 0; i <= count; i++ {
				wgLoc.Add(1)
				go putMultihashPartToMap(counters, i, data, mu, wgLoc)
			}

			wgLoc.Wait()

			out <- getMultihash(counters, count)
		}(out)
	}
	wg.Wait()
}

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

var CombineResults = func(in, out chan interface{}) {
	var all []string
	for dataRaw := range in {
		data := dataRaw.(string)
		all = append(all, data)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i] < all[j]
	})
	out <- strings.Join(all, "_")
}
