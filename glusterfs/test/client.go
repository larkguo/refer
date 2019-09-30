package main

import (
	"flag"
	"fmt"

	"github.com/gluster/gogfapi/gfapi"
)

func main() {
	var host, volume, writeFile string

	/* paramParse */
	flag.Bool("?|h", false, "./GlusterClient -h localhost -v vol_distributed -w write_file")
	flag.StringVar(&host, "h", "localhost", "Host")
	flag.StringVar(&volume, "v", "", "Volume")
	flag.StringVar(&writeFile, "w", "writefile", "WriteFileName")
	flag.Parse()
	parsed := flag.Parsed()
	if parsed {
		if volume == "" {
			flag.Usage()
			return
		}
	}

	fmt.Println("\nGlusterClient: ")
	fmt.Printf(" -h\tHost: '%s'\n", host)
	fmt.Printf(" -v\tVolume: '%s'\n", volume)
	fmt.Printf(" -w\tWriteFileName: '%s'\n", writeFile)

	vol := &gfapi.Volume{}
	if err := vol.Init(volume, host); err != nil {
		fmt.Println("Init err:", err)
		return
	}
	
	if err := vol.Mount(); err != nil {
		fmt.Println("Mount err:", err)
		return
	}
	defer vol.Unmount()
	
	f, err := vol.Create(writeFile)
	if err != nil {
		fmt.Println("Create err:", err)
		return
	}
	defer f.Close()
	
	if _, err := f.Write([]byte("hello")); err != nil {
		fmt.Println("Write err:", err)
		return
	}
	
	fmt.Printf("write to host[%s] volume[%s] file[%s] ok!\n", host, volume, writeFile)
	return
}
