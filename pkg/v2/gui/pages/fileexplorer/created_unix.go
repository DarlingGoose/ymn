//go:build linux

package fileexplorer

import (
	"io/fs"
	"syscall"
	"time"
)

func createdTime(info fs.FileInfo) time.Time {
	if info == nil {
		return time.Time{}
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return info.ModTime()
	}

	return time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
}
