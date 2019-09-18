package main

import (
	"io"
	"os"
	"strings"
)

type rotReader struct {
	r io.Reader
}

// 转换byte  A-M和a-m前进1位,N-Z和n-z后退1位
func rot(b byte) byte {
	switch {
	case 'A' <= b && b <= 'M':
		b = b + 1
	case 'M' < b && b <= 'Z':
		b = b - 1
	case 'a' <= b && b <= 'm':
		b = b + 1
	case 'm' < b && b <= 'z':
		b = b - 1
	}
	return b
}

// 重写Read方法
func (mr rotReader) Read(b []byte) (int, error) {
	n, e := mr.r.Read(b)
	for i := 0; i < n; i++ {
		b[i] = rot(b[i])
	}
	return n, e
}
func main() {
	s := strings.NewReader("H kpwd zpv!") // I love you!
	r := rotReader{s}
	io.Copy(os.Stdout, &r)
}
