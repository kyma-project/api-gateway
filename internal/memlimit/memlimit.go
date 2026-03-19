package memlimit

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
)

const (
	cgroupV2MemMax = "/sys/fs/cgroup/memory.max"
	cgroupV1MemMax = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	maxReasonable  = int64(1) << 60
)

// SetGoMemLimitFromCgroup sets GOMEMLIMIT to pct (e.g. 0.9 for 90%) of the
// container memory limit read from the cgroup filesystem.
func SetGoMemLimitFromCgroup(pct float64, log logr.Logger) error {
	if pct <= 0 || pct > 1 {
		return fmt.Errorf("percentage must be a fraction of 1, got %f", pct)
	}

	limit, err := readCgroupLimit()
	if err != nil {
		return fmt.Errorf("reading cgroup memory limit: %w", err)
	}

	if limit <= 0 || limit > maxReasonable {
		return fmt.Errorf("invalid or unlimited memory limit: %d", limit)
	}

	target := int64(float64(limit) * pct)
	log.Info("Setting GOMEMLIMIT", "limitMiB", bytesToMiB(limit), "targetMiB", bytesToMiB(target))
	debug.SetMemoryLimit(target)
	return nil
}

func readCgroupLimit() (int64, error) {
	if b, err := os.ReadFile(cgroupV2MemMax); err == nil {
		return parseCgroup(string(b))
	}

	b, err := os.ReadFile(cgroupV1MemMax)
	if err != nil {
		return 0, fmt.Errorf("unable to read cgroup memory limit: %w", err)
	}
	return parseCgroup(string(b))
}

func parseCgroup(raw string) (int64, error) {
	s := strings.TrimSpace(raw)
	if s == "max" {
		return 0, fmt.Errorf("no memory limit set (cgroup reports 'max')")
	}
	return strconv.ParseInt(s, 10, 64)
}

func bytesToMiB(b int64) int64 {
	return b / (1024 * 1024)
}
