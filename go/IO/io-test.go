/*
https://www.jishuwen.com/d/2dac
*/
package main

import (
    "fmt"
    "io"
    "os"
    "strings"
    "sync"
)

// 统一定义一个方法来处理错误，这样不会看到很多 if err != nil {} 这种
func executeIfNoErr(err error, f func()) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "\tERROR: %v\n", err)
        return
    }
    f()
}

func main() {
    comment := "Make the plan. " +
        "Execute the plan. " +
        "Expect the plan to go off the rails. " +
        "Throw away the plan."

    fmt.Println("原生string类型:")
    reader1 := strings.NewReader(comment)
    buf1 := make([]byte, 4)
    n, err := reader1.Read(buf1)
    var offset1, index1 int64
    executeIfNoErr(err, func() {
        fmt.Printf("\tRead(%d): %q\n", n, buf1[:n])
        offset1 = int64(5)
        index1, err = reader1.Seek(offset1, io.SeekCurrent)
    })
    executeIfNoErr(err, func() {
        fmt.Printf("\t偏移量: %d,  %d\n", offset1, index1)
        n, err = reader1.Read(buf1)
    })
    executeIfNoErr(err, func() {
        fmt.Printf("\tRead(%d): %q\n", n, buf1[:n])
    })

    reader1.Reset(comment)
    num2 := int64(15)
    fmt.Printf("LimitReader类型，限制数据量(%d):\n", num2)
    reader2 := io.LimitReader(reader1, num2)
    buf2 := make([]byte, 4)
    for i := 0; i < 6; i++ {
        n, err := reader2.Read(buf2)
        executeIfNoErr(err, func() {
            fmt.Printf("\tRead(%d): %q\n", n, buf2[:n])
        })
    }

    reader1.Reset(comment)
    offset3 := int64(33)
    num3 := int64(37)
    fmt.Printf("SectionReader类型，起始偏移量(%d)，到末端的长度(%d):\n", offset3, num3)
    reader3 := io.NewSectionReader(reader1, offset3, num3)
    buf3 := make([]byte, 15)
    for i := 0; i < 5; i++ {
        n, err := reader3.Read(buf3)
        executeIfNoErr(err, func() {
            fmt.Printf("\tRead(%d): %q\n", n, buf3[:n])
        })
    }

    reader1.Reset(comment)
    writer4 := new(strings.Builder)
    fmt.Printf("teeReader类型，write4现在应该为空(%q):\n", writer4)
    reader4 := io.TeeReader(reader1, writer4)
    buf4 := make([]byte, 33)
    for i := 0; i < 5; i++ {
        n, err := reader4.Read(buf4)
        executeIfNoErr(err, func() {
            fmt.Printf("\tRead(%d): %q\n", n, buf4[:n])
            fmt.Printf("\tWrite: %q\n", writer4)
        })
    }

    reader5a := strings.NewReader("Make the plan.")
    reader5b := strings.NewReader("Execute the plan.")
    reader5c := strings.NewReader("Expect the plan to go off the rails.")
    reader5d := strings.NewReader("Throw away the plan.")
    fmt.Println("multiWriter类型，一共4个readers：")
    reader5 := io.MultiReader(reader5a, reader5b, reader5c, reader5d)
    buf5 := make([]byte, 15)
    for i := 0; i < 10; i++ {
        n, err := reader5.Read(buf5)
        executeIfNoErr(err, func() {
            fmt.Printf("\tRead(%d): %q\n", n, buf5[:n])
        })
    }

    fmt.Println("pipe类型:")
    pReader, pWriter := io.Pipe()
    _ = interface{}(pReader).(io.ReadCloser) // 验证是否实现了 io.ReadCloser 接口
    _ = interface{}(pWriter).(io.WriteCloser)
    var wg sync.WaitGroup
    wg.Add(2)
    go func() {
        defer wg.Done()
        n, err := pWriter.Write([]byte(comment))
        defer pWriter.Close()
        executeIfNoErr(err, func() {
            fmt.Printf("\tWrite(%d)\n", n)
        })
    }()
    go func() {
        defer wg.Done()
        buf6 := make([]byte, 15)
        for i := 0; i < 10; i++ {
            n, err := pReader.Read(buf6)
            executeIfNoErr(err, func() {
                fmt.Printf("\tRead(%d): %q\n", n, buf6[:n])
            })
        }
    }()
    wg.Wait()

    fmt.Println("所有示例完成")
}

