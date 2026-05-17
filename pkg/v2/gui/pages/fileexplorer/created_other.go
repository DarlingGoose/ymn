//go:build !linux && !darwin

package fileexplorer

import (
	"io/fs"
	"time"
)

func createdTime(info fs.FileInfo) time.Time {
	if info == nil {
		return time.Time{}
	}
	return info.ModTime()
}
