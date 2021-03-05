package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type FileDownloader struct {
	writer       io.WriterAt // 接口,实现Write方法
	fileSize     int64       // 文件大小
	url          string      //下载地址
	concurrency  int64       // 并发数
	doneFilePart []FilePart  // 所有数据分片结构
}
type FilePart struct {
	writer io.WriterAt // 继承自FileDownloader
	id     int
	from   int64 // 数据分片的base基准
	to     int64 // 数据分片的top
	cur    int64 // 一片数据内当前的写位置
}

func (part *FilePart) Write(p []byte) (n int, err error) {
	if part.cur > part.to {
		return 0, io.EOF
	}
	cur := part.from + part.cur
	n, err = part.writer.WriteAt(p, cur) // 每段数据按顺序多次回调,需要记录当前位置
	part.cur += int64(n)

	log.Printf("job[%v] get range[%v-%v) \n", part.id, cur, part.cur)
	return
}
func NewFileDownloader(url string, concurrency int64) *FileDownloader {
	file, err := os.Create(filepath.Base(url))
	if err != nil {
		return nil
	}
	return &FileDownloader{
		fileSize:     0,
		url:          url,
		writer:       file,
		concurrency:  concurrency,
		doneFilePart: make([]FilePart, concurrency),
	}
}
func (down *FileDownloader) head() (int64, error) {
	r, err := http.NewRequest("HEAD", down.url, nil) // 获取文件大小等元信息
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode > 299 {
		return 0, errors.New(fmt.Sprintf("Can't process, %v", resp.Status))
	}
	if resp.Header.Get("Accept-Ranges") != "bytes" { // 不支持分段
		return 0, errors.New("Server not support of partial requests")
	}
	n, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	return n, err
}
func (down *FileDownloader) Run() error {
	fileSize, err := down.head()
	if err != nil {
		return err
	}
	down.fileSize = fileSize

	jobs := make([]FilePart, down.concurrency)
	pageSize := fileSize / down.concurrency // 分片大小

	for i := range jobs { // 创建多个任务
		jobs[i].id = i
		if i == 0 {
			jobs[i].from = 0
		} else {
			jobs[i].from = jobs[i-1].to + 1
		}
		if i < int(down.concurrency-1) {
			jobs[i].to = jobs[i].from + pageSize
		} else {
			jobs[i].to = fileSize - 1 // 最后一片
		}
	}

	var wg sync.WaitGroup
	for _, j := range jobs { // 多个任务同时下载
		wg.Add(1)
		go func(job FilePart) {
			defer wg.Done()
			_, err := down.downPart(&job) // 分段分页下载
			if err != nil {
				log.Println("downPart:", err, job)
			}
		}(j)
	}
	wg.Wait() // 所有任务执行完成
	return nil
}
func (d *FileDownloader) downPart(part *FilePart) (int64, error) {
	r, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		return 0, err
	}
	part.writer = d.writer // 继承writer
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", part.from, part.to))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	n, err := io.Copy(part, resp.Body) // 回调FilePart.Write()
	if err != nil {
		return 0, err
	}
	return n, nil
}
func main() {
	startTime := time.Now()

	url := flag.String("u", "https://mojotv.cn/go/go-range-download", "http(s)://... ")
	jobs := flag.Int64("j", 3, "Concurrency jobs ")
	flag.Parse()
	if *url == "" {
		fmt.Println("Specify a Url(-u 'http(s)://...' ) Jobs (-j JOBS) ")
		return
	}
	downloader := NewFileDownloader(*url, *jobs)
	if err := downloader.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Took: %f second\n", time.Now().Sub(startTime).Seconds())
}
