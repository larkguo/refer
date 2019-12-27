package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// context方式取消
func cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
func main() {
	ctx, cancelfunc := context.WithCancel(context.Background())
	// Determine the initial directories.
	roots := os.Args[1:]
	if len(roots) == 0 {
		roots = []string{"."}
	}

	go func() { //触发取消
		os.Stdin.Read(make([]byte, 1)) // read a single byte
		cancelfunc()
	}()

	// Traverse each root of the file tree in parallel.
	fileSizes := make(chan int64)
	var n sync.WaitGroup
	for _, root := range roots {
		n.Add(1)
		go walkDir(ctx, root, &n, fileSizes)
	}
	go func() {
		n.Wait()
		close(fileSizes)
	}()
	tick := time.Tick(500 * time.Millisecond) // Print the results periodically.
	var nfiles, nbytes int64
loop:
	for {
		select {
		case <-ctx.Done():
			// Drain fileSizes to allow existing goroutines to finish.
			for range fileSizes {
				// Do nothing.
			}
			return
		case size, ok := <-fileSizes:
			if !ok {
				break loop // fileSizes was closed
			}
			nfiles++
			nbytes += size
		case <-tick:
			printDiskUsage(nfiles, nbytes)
		}
	}
	printDiskUsage(nfiles, nbytes) // final totals
}

func printDiskUsage(nfiles, nbytes int64) {
	fmt.Printf("%d files  %.1f GB\n", nfiles, float64(nbytes)/1e9)
}

// walkDir recursively walks the file tree rooted at dir
// and sends the size of each found file on fileSizes.
func walkDir(ctx context.Context, dir string, n *sync.WaitGroup, fileSizes chan<- int64) {
	defer n.Done()
	if cancelled(ctx) {
		return // 已经取消
	}
	for _, entry := range dirents(ctx, dir) {
		if entry.IsDir() {
			n.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go walkDir(ctx, subdir, n, fileSizes) // 递归遍历
		} else {
			fileSizes <- entry.Size() // 递归返回
		}
	}
}

var sema = make(chan struct{}, 20) //concurrency-limiting counting semaphore,并发限制计数

// dirents returns the entries of directory dir.
func dirents(ctx context.Context, dir string) []os.FileInfo {
	select {
	case sema <- struct{}{}: // acquire token,大于并发限制则阻塞等待
	case <-ctx.Done(): //取消判断
		return nil // cancelled
	}
	defer func() { <-sema }() // release token, 释放一个并发计数

	// ...read directory...
	f, err := os.Open(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "du: %v\n", err)
		return nil
	}
	defer f.Close()

	entries, err := f.Readdir(0) // 0 => no limit; read all entries
	if err != nil {
		fmt.Fprintf(os.Stderr, "du: %v\n", err)
		// Don't return: Readdir may return partial results.
	}
	return entries
}
