package main

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestPipeline(t *testing.T) {

	var ok = true
	var recieved uint32
	jobs := []job{
		job(func(in, out chan interface{}) {
			out <- 2
			time.Sleep(11 * time.Millisecond)
			if currRecieved := atomic.LoadUint32(&recieved); currRecieved == 0 {
				ok = false
			}
		}),
		job(func(in, out chan interface{}) {
			for _ = range in {
				atomic.AddUint32(&recieved, 2)
			}
		}),
	}
	Advertise(jobs...)
	if !ok || recieved == 0 {
		t.Errorf("Pipeline failed")
	}
}

func TestSigner(t *testing.T) {

	testExpected := "11257784923703605154281634540993982076713738313033459875609_13368272083497320150287587435788400682715660057473264076397_13368272083497320150287587435788400682715660057473264076397_222248225424714955662854791886317094657436381614863488715918_276300556530131145572329991165264717599741618999014013550013_37935537652110229243114886680256714637440414775981866370624_413687177337871789093636686557348828726128554169253172618397"
	testResult := "NOT_SET"

	var (
		DataSignerSalt     string = "" // на сервере будет другое значение
		OverLockCounter    uint32
		OverUnlockCounter  uint32
		SignerMd5Counter   uint32
		SignerCrc32Counter uint32
	)
	OverLock = func() {
		atomic.AddUint32(&OverLockCounter, 1)
		for {
			if swapped := atomic.CompareAndSwapUint32(&dataOverheat, 0, 1); !swapped {
				fmt.Println("OverheatLock happend")
				time.Sleep(time.Second)
			} else {
				break
			}
		}
	}
	OverUnlock = func() {
		atomic.AddUint32(&OverUnlockCounter, 1)
		for {
			if swapped := atomic.CompareAndSwapUint32(&dataOverheat, 1, 0); !swapped {
				fmt.Println("OverheatUnlock happend")
				time.Sleep(time.Second)
			} else {
				break
			}
		}
	}
	FastPredict = func(data string) string {
		atomic.AddUint32(&SignerMd5Counter, 1)
		OverLock()
		defer OverUnlock()
		data += DataSignerSalt
		dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
		time.Sleep(10 * time.Millisecond)
		return dataHash
	}
	SlowPredict = func(data string) string {
		atomic.AddUint32(&SignerCrc32Counter, 1)
		data += DataSignerSalt
		crcH := crc32.ChecksumIEEE([]byte(data))
		dataHash := strconv.FormatUint(uint64(crcH), 10)
		time.Sleep(time.Second)
		return dataHash
	}

	inputData := []int{0, 1, 1, 2, 3, 5, 8}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(GetProfile),
		job(GetGroup),
		job(ConcatProfiles),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				t.Error("cant convert result data to string")
			}
			testResult = data
		}),
	}

	start := time.Now()

	Advertise(hashSignJobs...)

	end := time.Since(start)

	expectedTime := 3 * time.Second

	if testExpected != testResult {
		t.Errorf("results not match\nGot: %v\nExpected: %v", testResult, testExpected)
	}

	if end > expectedTime {
		t.Errorf("execition too long\nGot: %s\nExpected: <%s", end, time.Second*3)
	}

	if int(OverLockCounter) != len(inputData) ||
		int(OverUnlockCounter) != len(inputData) ||
		int(SignerMd5Counter) != len(inputData) ||
		int(SignerCrc32Counter) != len(inputData)*8 {
		t.Errorf("not enough hash-func calls")
	}

}
