package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	fastPredictMtx = sync.Mutex{}
)

func FastPredictProtected(data string) string {
	fastPredictMtx.Lock()
	defer fastPredictMtx.Unlock()
	return FastPredict(data)
}

func SlowPredictAsync(data string) <-chan string {
	ch := make(chan string)
	go func() {
		ch <- SlowPredict(data)
	}()
	return ch
}

func Advertise(freeFlowJobs ...job) {
	chans := make([]chan interface{}, len(freeFlowJobs))
	for i := range chans {
		chans[i] = make(chan interface{})
	}

	wg := sync.WaitGroup{}
	wg.Add(len(freeFlowJobs))

	for i, job := range freeFlowJobs {
		i := i
		job := job
		go func() {
			if i == 0 {
				job(nil, chans[i])
			} else {
				job(chans[i-1], chans[i])
			}
			close(chans[i])
			wg.Done()
		}()
	}

	wg.Wait()
}

func GetProfile(in, out chan interface{}) {
	wg := sync.WaitGroup{}
	for data := range in {
		userId := strconv.FormatInt(int64(data.(int)), 10)
		wg.Add(1)
		go func() {
			sCh := SlowPredictAsync(userId)
			sfCh := SlowPredictAsync(FastPredictProtected(userId))

			out <- <-sCh + "-" + <-sfCh
			wg.Done()
		}()
	}
	wg.Wait()
}

func GetGroup(in, out chan interface{}) {
	wg := sync.WaitGroup{}
	for data := range in {
		profile := data.(string)

		wg.Add(1)
		go func() {
			groups := make([]string, 6)

			groupsWg := sync.WaitGroup{}
			groupsWg.Add(6)
			for i := range groups {
				i := i
				go func() {
					groups[i] = SlowPredict(strconv.FormatInt(int64(i), 10) + profile)
					groupsWg.Done()
				}()
			}
			groupsWg.Wait()

			out <- strings.Join(groups, "")
			wg.Done()
		}()
	}

	wg.Wait()
}

func ConcatProfiles(in, out chan interface{}) {
	var profiles []string
	for data := range in {
		profiles = append(profiles, data.(string))
	}

	sort.Strings(profiles)

	out <- strings.Join(profiles, "_")
}
