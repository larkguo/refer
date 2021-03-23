package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

func fileCopy(fw io.Writer, fr io.Reader) (written int64, err error) {
	return io.Copy(fw, fr) //回调fr.Read()读取byte并回调fw.Write()把读的byte写入
}
func pipeCopy(fw io.Writer, fr io.Reader) (written int64, err error) {
	pr, pw := io.Pipe() // 增加一层pipe(reader和writer)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer func() {
			wg.Done()
		}()
		io.Copy(pw, fr) //回调fr.Read()读取byte并回调pw.Write()把读的byte写入
		pw.Close()
	}()
	go func() {
		defer func() {
			wg.Done()
		}()
		written, err = io.Copy(fw, pr) //回调pr.Read()读取byte并回调fw.Write()把读的byte写入
		pr.Close()
	}()
	wg.Wait()
	return written, err
}

type myWriter struct {
	w io.Writer
}

func (mw myWriter) Write(p []byte) (n int, err error) {
	//TODO: Implement your own processing
	return mw.w.Write(p)
}
func myWriterCopy(fw io.Writer, fr io.Reader) (written int64, err error) {
	mw := myWriter{w: fw}
	return io.Copy(mw, fr) //回调fr.Read()读取byte并回调myw.Write()把读的byte写入
}

type myReader struct {
	r io.Reader
}

func (m myReader) Read(p []byte) (n int, err error) {
	//TODO: Implement your own processing
	return m.r.Read(p)
}
func myReaderCopy(fw io.Writer, fr io.Reader) (written int64, err error) {
	mr := myReader{r: fr}
	return io.Copy(fw, mr.r) //回调mr.r.Read()读取byte并回调fw.Write()把读的byte写入
}

func main() {
	fr, err := os.Open("copy.go")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fr.Close()
	fw, err := os.Create("copy")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fw.Close()

	fr.Seek(0, io.SeekStart)
	fw.Seek(0, io.SeekStart)
	n1, err1 := fileCopy(fw, fr)
	if err1 != nil {
		fmt.Println(err1)
	}
	fmt.Println("fileCopy:", n1)

	fr.Seek(0, io.SeekStart)
	n2, err2 := pipeCopy(fw, fr)
	if err2 != nil {
		fmt.Println(err2)
	}
	fmt.Println("pipeCopy:", n2)

	fr.Seek(0, io.SeekStart)
	n3, err3 := myWriterCopy(fw, fr)
	if err3 != nil {
		fmt.Println(err3)
	}
	fmt.Println("myWriterCopy:", n3)

	fr.Seek(0, io.SeekStart)
	n4, err4 := myReaderCopy(fw, fr)
	if err4 != nil {
		fmt.Println(err4)
	}
	fmt.Println("myReaderCopy:", n4)
}
