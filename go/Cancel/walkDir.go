/*
 并发目录遍历 https://books.studygolang.com/gopl-zh/ch8/ch8-08.html
 并发的退出 https://books.studygolang.com/gopl-zh/ch8/ch8-09.html
 code https://github.com/adonovan/gopl.io/blob/master/ch8/du4/main.go
 Using context cancellation in Go https://www.sohamkamani.com/golang/2018-06-17-golang-using-context-cancellation/
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var verbose = flag.Bool("v", false, "显示详细进度")

func main() {
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}
	fileSizes := make(chan int64)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background()) // 上下文
	go func() {
		os.Stdin.Read(make([]byte, 1)) // 回车换行退出
		cancel()                       // Done() <-chan struct{}
	}()
	start := time.Now()
	for _, root := range roots {
		wg.Add(1)
		go walkDir(ctx, root, &wg, fileSizes) // 并发遍历目录
	}
	go func() {
		wg.Wait()
		close(fileSizes)
	}()

	var tick <-chan time.Time
	if *verbose {
		tick = time.Tick(1 * time.Second)
	}
	var nfiles, nbytes int64
loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancel!")
			break loop
		case size, ok := <-fileSizes: // 从通道中获取文件大小
			if !ok {
				break loop
			}
			nfiles++
			nbytes += size
		case <-tick: // 进度
			printDiskUsage(nfiles, nbytes)
		}
	}
	printDiskUsage(nfiles, nbytes)
	fmt.Println("Spent " + time.Since(start).String())
}
func printDiskUsage(nfiles, nbytes int64) {
	fmt.Printf("%d files  %.1f GB\n", nfiles, float64(nbytes)/1e9)
}
func walkDir(ctx context.Context, dir string, wg *sync.WaitGroup, fileSizes chan<- int64) {
	defer wg.Done()
	select {
	case <-ctx.Done():
		return
	default:
	}
	for _, entry := range dirents(ctx, dir) {
		if entry.IsDir() {
			wg.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go walkDir(ctx, subdir, wg, fileSizes) // 递归目录
		} else {
			fileSizes <- entry.Size()
		}
	}
}

var sema = make(chan struct{}, runtime.NumCPU()) // 限制并发
func dirents(ctx context.Context, dir string) []os.FileInfo {
	select {
	case sema <- struct{}{}:
	case <-ctx.Done():
		return nil
	}
	defer func() { <-sema }()

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "du: %v\n", err)
		return nil
	}
	return entries
}
