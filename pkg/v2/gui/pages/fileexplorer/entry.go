package fileexplorer

import (
	"io/fs"
	"strings"
	"time"
)

type entry struct {
	Name        string
	Path        string
	IsDir       bool
	Size        int64
	ModTime     time.Time
	CreatedTime time.Time
	Mode        fs.FileMode
}

type CommonPlace struct {
	Label string
	Path  string
	Icon  string
}

type SortBy int

const (
	SortByName SortBy = iota
	SortBySize
	SortByModified
	SortByCreated
	SortByKind
	SortByExtension
)

func (s SortBy) String() string {
	switch s {
	case SortBySize:
		return "Size"
	case SortByModified:
		return "Modified"
	case SortByCreated:
		return "Created"
	case SortByKind:
		return "Kind"
	case SortByExtension:
		return "Extension"
	default:
		return "Name"
	}
}

func sortByFromValue(value string) SortBy {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "size":
		return SortBySize
	case "modified":
		return SortByModified
	case "created":
		return SortByCreated
	case "kind":
		return SortByKind
	case "extension":
		return SortByExtension
	default:
		return SortByName
	}
}
