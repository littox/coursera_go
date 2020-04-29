package main

import (
	"fmt"
	//"runtime"
	"strconv"
	"sync"
	//"time"
)

func ExecutePipeline(jobs ...job) {
	wgm := &sync.WaitGroup{}
	wgm.Add(len(jobs))

	var out chan interface{}
	in := make(chan interface{}, 1)

	for _, fn := range jobs {
		out = make(chan interface{})

		go func(f job, wgm *sync.WaitGroup, in, out chan interface{}) {
			defer wgm.Done()
			defer close(out)
			f(in, out)
		}(fn, wgm, in, out)

		in = out
	}
	wgm.Wait()
}

var SingleHash = func(in, out chan interface{}) {
	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			fmt.Printf("cant convert result data to string")
		}
		md5chan := make(chan string)
		crc32md5chan := make(chan string)
		crc32chan := make(chan string)
		go putCrc32ToChan(crc32chan, data)
		go putMd5ToChan(md5chan, data)
		go putCrc32ToChan(crc32md5chan, <-md5chan)

		out <- <-crc32chan + "~" + <-crc32md5chan
	}
}

var MultiHash = func(in, out chan interface{}) {

	const count = 5

	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			fmt.Printf("cant convert result data to string")
		}

		parts := make(chan string)

		go func(cnt int, data string, parts chan string) {
			for i := 0; i <= count; i++ {
				parts <- DataSignerCrc32(strconv.Itoa(i) + data)
			}
			close(parts)
		}(count, data, parts)

		go func(parts chan string, out chan interface{}) {
			var res string
			for val := range parts {
				res += val
			}
			out <- res
		}(parts, out)
	}
}
var MultiHash = func(in, out chan interface{}) {

	dataRaw := <-in
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
}

func main() {

	for i := 0; i < 5; i++ {
		k := make(chan int)
		go func(ch chan int) {
			fmt.Println(i, ch, <-ch)
		}(k)
		k <- i + 1
	}
	//start := time.Now()
	//ch1 := make(chan interface{})
	//ch2 := make(chan interface{})
	////go MultiHash(ch1, ch2)
	//go SingleHash(ch1, ch2)
	//ch1 <- "hash"
	//res := <-ch2

	fmt.Scanln()
	//end := time.Since(start)
	//fmt.Println(res, end)
}

//
//
//type job func(in, out chan interface{})
//
//const (
//	MaxInputDataLen = 100
//)
//
//var (
//	dataSignerOverheat uint32 = 0
//	DataSignerSalt            = ""
//)
//
//var OverheatLock = func() {
//	for {
//		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 0, 1); !swapped {
//			fmt.Println("OverheatLock happend")
//			time.Sleep(time.Second)
//		} else {
//			break
//		}
//	}
//}
//
//var OverheatUnlock = func() {
//	for {
//		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 1, 0); !swapped {
//			fmt.Println("OverheatUnlock happend")
//			time.Sleep(time.Second)
//		} else {
//			break
//		}
//	}
//}
//
//var DataSignerMd5 = func(data string) string {
//	OverheatLock()
//	defer OverheatUnlock()
//	data += DataSignerSalt
//	dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
//	time.Sleep(10 * time.Millisecond)
//	return dataHash
//}
//
//var DataSignerCrc32 = func(data string) string {
//	data += DataSignerSalt
//	crcH := crc32.ChecksumIEEE([]byte(data))
//	dataHash := strconv.FormatUint(uint64(crcH), 10)
//	time.Sleep(time.Second)
//	return dataHash
//}
