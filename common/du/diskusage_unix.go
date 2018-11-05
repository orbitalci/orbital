// +build darwin dragonfly freebsd linux netbsd openbsd

package du

//https://github.com/StalkR/goircbot/blob/master/lib/disk/space_unix.go

import (
	"syscall"
)

// Space returns total and free bytes available in a directory, e.g. `/`.
// Think of it as "df" UNIX command.
func Space(path string) (total, free uint64, err error) {
	s := syscall.Statfs_t{}
	err = syscall.Statfs(path, &s)
	if err != nil {
		return
	}
	total = uint64(s.Bsize) * uint64(s.Blocks)
	free = uint64(s.Bsize) * uint64(s.Bfree)
	return
}