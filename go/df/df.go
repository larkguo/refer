package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var SIZE_KB uint64 = 1024
var SIZE_MB uint64 = 1048576
var SIZE_GB uint64 = 1073741824
var SIZE_TB uint64 = 1099511627776
var SIZE_PB uint64 = 1125899906842624

type FSStats struct {
	Blocks     uint64 // Total number of data blocks in a file system.
	BlockFree  uint64 // Free blocks in a file system.
	BlockAvail uint64
	BlockUsed  uint64
	UsePercent float64
	Type       string
}

// Mount describes a mounted filesytem. Please see man fstab for further details.
type Mount struct {
	FileSystem    string  // The field describes the block special device or remote filesystem to be mounted.
	MountPoint    string  // Describes the mount point for the filesytem.
	Type          string  // Describes the type of the filesystem.
	MntOps        string  // Describes the mount options associated with the filesystem.
	DumpFrequency int     // Dump frequency in days.
	PassNo        int     // Pass number on parallel fsck.
	FSStats       FSStats // Filesystem data, may be nil.
}

type DiskFree struct {
	// param
	MountMatch string

	// result
	Mounts map[string]Mount
}

func (df *DiskFree) BlockSizeToString(blockSize uint64) (rtn string) {
	var tempFloat float64

	if blockSize > SIZE_PB {
		tempFloat = float64(blockSize) / float64(SIZE_PB)
		rtn = fmt.Sprintf("%.1f", tempFloat) + "P"
		return
	}
	if blockSize > SIZE_TB {
		tempFloat = float64(blockSize) / float64(SIZE_TB)
		rtn = fmt.Sprintf("%.1f", tempFloat) + "T"
		return
	}
	if blockSize > SIZE_GB {
		tempFloat = float64(blockSize) / float64(SIZE_GB)
		rtn = fmt.Sprintf("%.1f", tempFloat) + "G"
		return
	}
	if blockSize > SIZE_MB {
		tempFloat = float64(blockSize) / float64(SIZE_MB)
		rtn = fmt.Sprintf("%.1f", tempFloat) + "M"
		return
	}
	if blockSize > SIZE_KB {
		tempFloat = float64(blockSize) / float64(SIZE_KB)
		rtn = fmt.Sprintf("%.1f", tempFloat) + "K"
		return
	}
	rtn = fmt.Sprintf("%d", blockSize)
	return
}

func (df *DiskFree) AddItem(line string) {
	mnt := Mount{}
	var err error
	if _, err = fmt.Sscanf(line, "%s %s %s %s %d %d", &mnt.FileSystem, &mnt.MountPoint, &mnt.Type, &mnt.MntOps, &mnt.DumpFrequency, &mnt.PassNo); err != nil {
		return
	}

	statfs := syscall.Statfs_t{}
	if err = syscall.Statfs(mnt.MountPoint, &statfs); err == nil {
		fsStats := FSStats{}
		fsStats.Blocks = statfs.Blocks * (uint64)(statfs.Bsize)
		fsStats.BlockFree = statfs.Bfree * (uint64)(statfs.Bsize)
		fsStats.BlockAvail = statfs.Bavail * (uint64)(statfs.Bsize)
		fsStats.BlockUsed = (statfs.Blocks - statfs.Bavail) * (uint64)(statfs.Bsize)
		UsePercent := (1 - (float64)(statfs.Bavail)/(float64)(statfs.Blocks)) * 100
		fsStats.UsePercent, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", UsePercent), 64)
		fsStats.Type = mnt.Type
		mnt.FSStats = fsStats
	}
	if mnt.FSStats.Blocks > 0 {
		df.Mounts[mnt.MountPoint] = mnt
	}
}

func (df *DiskFree) ParseMounts() {
	reader, err := os.Open("/etc/mtab")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer reader.Close()

	br := bufio.NewReader(reader)
	for line, err := br.ReadString('\n'); err == nil; line, err = br.ReadString('\n') {
		if df.MountMatch != "" {
			item := strings.ToLower(line)
			if strings.Contains(item, df.MountMatch) {
				df.AddItem(line)
			}
		} else {
			df.AddItem(line)
		}

	}
}

func NewDF(mountMatch string) *DiskFree {
	mounts := make(map[string]Mount)
	return &DiskFree{
		MountMatch: strings.ToLower(mountMatch),
		Mounts:     mounts,
	}
}

func main() {
	var dfMatch, dfType, dfFileSystem, dfSize, dfUsed, dfAvail, dfUsePercent, dfMountPoint string
	paramCount := len(os.Args)
	if paramCount > 1 {
		dfMatch = os.Args[1]
	}

	df := NewDF(dfMatch)
	df.ParseMounts()
	fmt.Printf("%-38s %-16s %8s %8s %6s %6s %-s\n", "Filesystem", "Type", "Size", "Used", "Avail", "Use%", "Mounted on")
	for _, mount := range df.Mounts {
		dfFileSystem = mount.FileSystem
		dfType = mount.Type
		dfSize = df.BlockSizeToString(mount.FSStats.Blocks)
		dfUsed = df.BlockSizeToString(mount.FSStats.BlockUsed)
		dfAvail = df.BlockSizeToString(mount.FSStats.BlockAvail)
		dfUsePercent = fmt.Sprintf("%.1f", mount.FSStats.UsePercent) + "%"
		dfMountPoint = mount.MountPoint
		fmt.Printf("%-38s %-16s %8s %8s %6s %6s %-s\n",
			dfFileSystem, dfType, dfSize, dfUsed, dfAvail, dfUsePercent, dfMountPoint)
	}
}
