package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

type OneDir struct {
	dir         string
	filePref    string
	fileExt     string
	fileIdStart int
	fileNum     int
	fileMin     int64
	fileMax     int64
}

func CreateDirs(wg *sync.WaitGroup, dirCh <-chan OneDir) {
	defer wg.Done()

	for one := range dirCh {
		err := os.MkdirAll(one.dir, os.ModePerm)
		if err != nil {
			log.Println(err)
			return
		}
		for j := 0; j < one.fileNum; j++ {
			fileId := j + one.fileIdStart
			filePref := fmt.Sprintf("%v%v", one.filePref, fileId)
			fileName := fmt.Sprintf("%v.%v", filePref, one.fileExt)
			allPath := path.Join(one.dir, fileName)

			f, err := os.Create(allPath)
			if err != nil {
				log.Println(err)
				return
			}

			rand.Seed(time.Now().UnixNano())
			sizeRand := rand.Int63n(one.fileMax-one.fileMin) + one.fileMin
			if err := f.Truncate(sizeRand); err != nil {
				f.Close()
				log.Println(err)
				return
			}

			f.Close()
		}
		log.Printf("%v ok", one.dir)
	}
	return
}

func InputDirs(wg *sync.WaitGroup, dirCh chan<- OneDir,
	filePref, fileExt, dayFrom string,
	days, fileIdStart, fileNum int, fileMin, fileMax int64) {

	defer wg.Done()

	var one OneDir
	var fromTime time.Time
	var err error
	var pwd string
	var i int

	fromTime, err = time.Parse("01/02/2006", dayFrom)
	if err != nil {
		log.Println(err)
		return
	}
	pwd, err = os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}

	for i = 0; i < days; i++ {
		to := fromTime.AddDate(0, 0, i)
		year := to.Format("2006")
		month := to.Format("01")
		day := to.Format("02")
		name := fmt.Sprintf("%v%02v%02v", year, month, day)
		dir := path.Join(pwd, name)

		one.dir = dir
		one.filePref = filePref
		one.fileExt = fileExt
		one.fileIdStart = fileIdStart
		one.fileNum = fileNum
		one.fileMin = fileMin
		one.fileMax = fileMax

		dirCh <- one
	}
	close(dirCh)
}

func main_files() {
	dayFrom := flag.String("t", "01/01/2016", "Time from(mm/dd/yyyy)")
	days := flag.Int("d", 365, "Days ")
	fileNum := flag.Int("n", 10240, "Number of files in one directory")
	threadNum := flag.Int("c", -1, "Concurrency, -1|0 : number of logical CPUs")
	fileMin := flag.Int64("i", 1024, "MinSize of files")
	fileMax := flag.Int64("a", 1024000, "MaxSize of files")
	filePref := flag.String("p", "", "Prefix of files")
	fileIdStart := flag.Int("s", 10000000, "Start ID fo files")
	fileExt := flag.String("e", "bob", "Extension of files")
	flag.Parse()

	if *threadNum <= 0 {
		*threadNum = runtime.NumCPU() //number of logical CPUs
	}
	dirCh := make(chan OneDir, *threadNum)
	wg := sync.WaitGroup{}
	wg.Add(1)

	// producer
	go InputDirs(&wg, dirCh, *filePref, *fileExt, *dayFrom, *days, *fileIdStart, *fileNum, *fileMin, *fileMax)

	// consumer
	for i := 0; i < *threadNum; i++ { // fan-out
		wg.Add(1)
		go CreateDirs(&wg, dirCh)
	}

	wg.Wait()
}
