package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
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

type MountString struct {
	FileSystem string
	Type       string
	Size       string 
	Used       string 
	Avail      string 
	UsePercent string 
	MountPoint string 
}

func BlockSizeToString(blockSize uint64) (rtn string) {
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

func ParseMounts(mounts map[string]Mount, reader io.Reader) {
	br := bufio.NewReader(reader)
	for s, err := br.ReadString('\n'); err == nil; s, err = br.ReadString('\n') {
		mnt := Mount{}
		if _, err := fmt.Sscanf(s, "%s %s %s %s %d %d", &mnt.FileSystem, &mnt.MountPoint, &mnt.Type, &mnt.MntOps, &mnt.DumpFrequency, &mnt.PassNo); err != nil {
			continue
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
			mounts[mnt.FileSystem] = mnt
		}
	}
}

func main() {
	f, err := os.Open("/etc/mtab")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer f.Close()

	mounts := make(map[string]Mount)

	ParseMounts(mounts, f)
	for _, mount := range mounts {
		var item MountString
		item.FileSystem = mount.FileSystem
		item.Size = BlockSizeToString(mount.FSStats.Blocks)
		item.Used = BlockSizeToString(mount.FSStats.BlockUsed)
		item.Avail = BlockSizeToString(mount.FSStats.BlockAvail)
		item.UsePercent = fmt.Sprintf("%.1f", mount.FSStats.UsePercent) + "%"
		item.MountPoint = mount.MountPoint
		item.Type = mount.Type

		fmt.Println(item)
	}
}

