//go:build linux && !js

package hostmem

import (
	"bufio"
	"bytes"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

// Total returns the kernel's MemTotal: the usable physical memory in bytes.
func Total() (uint64, error) {
	return meminfo("MemTotal:")
}

// Available returns the kernel's MemAvailable estimate in bytes: free memory
// plus the page cache and slab it can reclaim without swapping. Free memory
// alone excludes the page cache, which any host reading files fills.
func Available() (uint64, error) {
	return meminfo("MemAvailable:")
}

// meminfo returns the /proc/meminfo field named key, converted from KiB to
// bytes.
func meminfo(key string) (uint64, error) {
	// Read the kernel's memory report.
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	// Find the field and convert its KiB value to bytes.
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		fields := bytes.Fields(scanner.Bytes())
		if len(fields) != 3 || string(fields[0]) != key {
			continue
		}
		kib, err := strconv.ParseUint(string(fields[1]), 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "parse "+key)
		}
		return kib << 10, nil
	}
	return 0, errors.Errorf("%s not found in /proc/meminfo", key)
}
