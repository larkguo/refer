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
	concurrency  int         // 并发数
	maxRetries   int         // 重试次数
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
	n, err = part.writer.WriteAt(p, cur) //每段数据按顺序多次回调,需要记录当前位置
	part.cur += int64(n)
	log.Printf("job[%v] get range[%v-%v) \n", part.id, cur, part.cur)
	return
}

func NewFileDownloader(url string, concurrency, retries int) *FileDownloader {
	file, err := os.Create(filepath.Base(url))
	if err != nil {
		return nil
	}
	return &FileDownloader{
		fileSize:     0,
		url:          url,
		writer:       file,
		concurrency:  concurrency,
		maxRetries:   retries,
		doneFilePart: make([]FilePart, concurrency),
	}
}

func (down *FileDownloader) head() (int64, error) {
	r, err := http.NewRequest("HEAD", down.url, nil) //获取文件大小等元信息
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
	if resp.Header.Get("Accept-Ranges") != "bytes" { //不支持分段
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

	parts := make([]FilePart, down.concurrency)
	pageSize := fileSize / int64(down.concurrency) //分片大小

	for i := range parts { //创建多个任务
		parts[i].id = i
		if i == 0 {
			parts[i].from = 0
		} else {
			parts[i].from = parts[i-1].to + 1
		}
		if i < int(down.concurrency-1) {
			parts[i].to = parts[i].from + pageSize
		} else {
			parts[i].to = fileSize - 1 //最后一片
		}
	}

	var wg sync.WaitGroup
	for _, part := range parts { //多片同时下载
		wg.Add(1)
		go func(p FilePart) {
			defer wg.Done()
			for retry := 0; retry <= down.maxRetries; retry++ {
				_, err := down.downPart(&p) // 分段分页下载
				if err == nil {
					break
				}
			}
		}(part)
	}

	wg.Wait() //所有任务执行完成
	return nil
}
func (d *FileDownloader) downPart(part *FilePart) (n int64, err error) {
	r, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		return 0, err
	}
	part.writer = d.writer //继承writer
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", part.from, part.to))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	n, err = io.Copy(part, resp.Body) //回调FilePart.Write()
	if err != nil {
		return 0, err
	}
	return n, nil
}
func main() {
	startTime := time.Now()
	url := flag.String("u", "https://mojotv.cn/go/go-range-download", "http(s)://... ")
	jobs := flag.Int("j", 3, "Concurrency jobs ")
	retries := flag.Int("r", 3, "MaxRetries ")
	flag.Parse()
	downloader := NewFileDownloader(*url, *jobs, *retries)
	if err := downloader.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Took: %f second\n", time.Now().Sub(startTime).Seconds())
}

