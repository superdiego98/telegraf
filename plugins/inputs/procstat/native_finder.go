package procstat

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

// NativeFinder uses gopsutil to find processes
type NativeFinder struct{}

// Uid will return all pids for the given user
func (pg *NativeFinder) UID(user string) ([]PID, error) {
	var dst []PID
	procs, err := process.Processes()
	if err != nil {
		return dst, err
	}
	for _, p := range procs {
		username, err := p.Username()
		if err != nil {
			// skip, this can be caused by the pid no longer exists, or you don't have permissions to access it
			continue
		}
		if username == user {
			dst = append(dst, PID(p.Pid))
		}
	}
	return dst, nil
}

// PidFile returns the pid from the pid file given.
func (pg *NativeFinder) PidFile(path string) ([]PID, error) {
	var pids []PID
	pidString, err := os.ReadFile(path)
	if err != nil {
		return pids, fmt.Errorf("failed to read pidfile %q: %w", path, err)
	}
	pid, err := strconv.ParseInt(strings.TrimSpace(string(pidString)), 10, 32)
	if err != nil {
		return pids, err
	}
	pids = append(pids, PID(pid))
	return pids, nil
}

// FullPattern matches on the command line when the process was executed
func (pg *NativeFinder) FullPattern(pattern string) ([]PID, error) {
	var pids []PID
	regxPattern, err := regexp.Compile(pattern)
	if err != nil {
		return pids, err
	}
	procs, err := pg.FastProcessList()
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		cmd, err := p.Cmdline()
		if err != nil {
			// skip, this can be caused by the pid no longer exists, or you don't have permissions to access it
			continue
		}
		if regxPattern.MatchString(cmd) {
			pids = append(pids, PID(p.Pid))
		}
	}
	return pids, err
}

// Children matches children pids on the command line when the process was executed
func (pg *NativeFinder) Children(pid PID) ([]PID, error) {
	// Get all running processes
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("getting process %d failed: %w", pid, err)
	}

	// Get all children of the current process
	children, err := p.Children()
	if err != nil {
		return nil, fmt.Errorf("unable to get children of process %d: %w", p.Pid, err)
	}
	pids := make([]PID, 0, len(children))
	for _, child := range children {
		pids = append(pids, PID(child.Pid))
	}

	return pids, err
}

func (pg *NativeFinder) FastProcessList() ([]*process.Process, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}

	result := make([]*process.Process, 0, len(pids))
	for _, pid := range pids {
		result = append(result, &process.Process{Pid: pid})
	}
	return result, nil
}

// Pattern matches on the process name
func (pg *NativeFinder) Pattern(pattern string) ([]PID, error) {
	var pids []PID
	regxPattern, err := regexp.Compile(pattern)
	if err != nil {
		return pids, err
	}
	procs, err := pg.FastProcessList()
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		name, err := processName(p)
		if err != nil {
			// skip, this can be caused by the pid no longer exists, or you don't have permissions to access it
			continue
		}
		if regxPattern.MatchString(name) {
			pids = append(pids, PID(p.Pid))
		}
	}
	return pids, err
}
