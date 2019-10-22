package main

import (
	"fmt"
	"strings"
)

type DevEntry struct {
	name   string
	state  string
	read   string
	write  string
	cksum  string
	rest   string
	level  int //级别,标识设备层次关系
	device int //1:是最底层设备
}

func poolConfParse(confstr string) (devs []DevEntry, err error) {
	minIndent := 0
	for _, origline := range strings.Split(confstr, "\n") {
		if origline == "" {
			continue
		}
		origlen := len(origline)
		line := strings.TrimLeft(origline, "\t")
		line = strings.TrimLeft(line, " ")
		indent := (origlen - len(line)) / 2
		f := strings.Fields(line)
		lenf := len(f)
		if f == nil || lenf <= 0 {
			continue
		}

		if f[0] == "config:" {
			continue
		}
		if len(f) >= 5 && f[0] == "NAME" && f[1] == "STATE" && f[2] == "READ" && f[3] == "WRITE" && f[4] == "CKSUM" {
			minIndent = indent
			continue
		}

		var dev DevEntry
		dev.level = indent - minIndent
		dev.device = 0

		if lenf > 0 {
			dev.name = f[0]
		}
		if lenf > 1 {
			dev.state = f[1]
		}
		if lenf > 2 {
			dev.read = f[2]
		}
		if lenf > 3 {
			dev.write = f[3]
		}
		if lenf > 4 {
			dev.cksum = f[4]
		}
		if lenf > 5 {
			dev.rest = strings.Join(f[5:], " ")
		}
		if dev.level > 0 {
			lowerName := strings.ToLower(dev.name)
			if !strings.Contains(lowerName, "mirror") &&
				//!strings.Contains(lowerName, "spares") &&
				//!strings.Contains(lowerName, "cache") &&
				//!strings.Contains(lowerName, "logs") &&
				!strings.Contains(lowerName, "raidz") &&
				!strings.Contains(lowerName, "replacing") {
				dev.device = 1
			}
		}
		devs = append(devs, dev)
	}
	return devs, nil
}

/*
{"name":"pool2","state":"DEGRADED","read":"0","write":"0","cksum":"0","rest":"","level":"0","device":"0"}
{"name":"mirror-0","state":"DEGRADED","read":"0","write":"0","cksum":"0","rest":"","level":"1","device":"0"}
{"name":"replacing-0","state":"DEGRADED","read":"0","write":"0","cksum":"0","rest":"","level":"2","device":"0"}
{"name":"wwn-0x6660","state":"UNAVAIL","read":"0","write":"0","cksum":"0","rest":"cannot open","level":"3","device":"1"}
{"name":"wwn-0x6634","state":"ONLINE","read":"0","write":"0","cksum":"0","rest":"8M resilvered","level":"3","device":"1"}
{"name":"wwn-0x6668","state":"ONLINE","read":"0","write":"0","cksum":"0","rest":"","level":"2","device":"1"}
{"name":"cache","state":"","read":"","write":"","cksum":"","rest":"","level":"0","device":"0"}
{"name":"wwn-0x66c0","state":"ONLINE","read":"0","write":"0","cksum":"0","rest":"","level":"1","device":"1"}
{"name":"spares","state":"","read":"","write":"","cksum":"","rest":"","level":"0","device":"0"}
{"name":"wwn-0x66c2","state":"AVAIL","read":"","write":"","cksum":"","rest":"","level":"1","device":"1"}
*/
func main() {
	var confStr string

	confStr = `
		config:

		NAME              STATE     READ WRITE CKSUM
		pool2             DEGRADED     0     0     0
		  mirror-0        DEGRADED     0     0     0
		    replacing-0   DEGRADED     0     0     0
		      wwn-0x6660  UNAVAIL      0     0     0 cannot open
		      wwn-0x6634  ONLINE       0     0     0 8M resilvered
		    wwn-0x6668    ONLINE       0     0     0
		cache
		  wwn-0x66c0      ONLINE       0     0     0
		spares
		  wwn-0x66c2      AVAIL

		`

	devs, err := poolConfParse(confStr)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, dev := range devs {
			devStr := fmt.Sprintf(`{"name":"%s","state":"%s","read":"%s","write":"%s","cksum":"%s","rest":"%s","level":"%d","device":"%d"}`,
				dev.name, dev.state, dev.read, dev.write, dev.cksum, dev.rest, dev.level, dev.device)
			fmt.Println(devStr)
		}
	}

}
