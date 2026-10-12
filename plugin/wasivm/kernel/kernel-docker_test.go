//go:build docker_syscall_tests

package kernel

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

// TestDockerParity checks BusyBox commands print the same output and exit
// with the same code under the kernel as in an Alpine container. Each command
// creates the files it reads, so the two filesystems need not match.
func TestDockerParity(t *testing.T) {
	requireDocker(t)
	tests := []struct {
		name  string
		stdin string
		args  []string
	}{
		{name: "echo", args: shArgs("echo hello world")},
		{name: "pipe", args: shArgs("echo HELLO | tr A-Z a-z")},
		{name: "multi-pipe", args: shArgs("echo abc | cat | cat")},
		{name: "cat-file", args: shArgs("echo testdata > /tmp/f && cat /tmp/f")},
		{name: "cwd", args: shArgs("cd /tmp && pwd")},
		{name: "mkdir", args: shArgs("mkdir -p /tmp/docker_test_dir && ls /tmp/docker_test_dir")},
		{name: "append", args: shArgs("echo line1 > /tmp/a && echo line2 >> /tmp/a && cat /tmp/a")},
		{name: "status", args: shArgs("false; echo $?")},
		{name: "true", args: []string{"/bin/busybox", "true"}},
		{name: "false", args: []string{"/bin/busybox", "false"}},
		{name: "printf", args: shArgs(`printf '%s %d\n' hello 42`)},
		{name: "seq", args: shArgs("seq 1 5")},
		{name: "arithmetic", args: shArgs("echo $((2 + 3 * 4))")},
		{name: "variables", args: shArgs("X=hello; echo $X world")},
		{name: "uname", args: shArgs("uname -s")},
		{name: "test", args: shArgs("if [ 1 -eq 1 ]; then echo yes; else echo no; fi")},
		{name: "grep", args: shArgs(`echo -e "apple\nbanana\ncherry" | grep an`)},
		{name: "stdin", stdin: "line1\nline2\nline3\n", args: []string{"cat"}},
		{name: "substitution", args: shArgs("echo $(echo nested)")},
		{name: "rename", args: shArgs("echo data > /tmp/src && mv /tmp/src /tmp/dst && cat /tmp/dst")},
		{name: "here-doc", args: shArgs("cat <<EOF\nhello\nworld\nEOF")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Run the command in both.
			t.Parallel()
			got, gotCode := runKernelCommand(t, tc.stdin, tc.args...)
			want, wantCode := runDockerCommand(t, tc.stdin, tc.args...)

			// Compare the output and the exit code.
			if normalizeOutput(got) != normalizeOutput(want) || gotCode != wantCode {
				t.Errorf("kernel: %q, exit %d; docker: %q, exit %d", got, gotCode, want, wantCode)
			}
		})
	}
}

// shArgs returns the arguments that run script with sh -c.
func shArgs(script string) []string {
	return []string{"sh", "-c", script}
}

// requireDocker skips the test when Docker is not running.
func requireDocker(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "docker", "info").Run(); err != nil {
		t.Skip("docker not available: ", err)
	}
}

// runKernelCommand runs a BusyBox command line under the kernel and returns
// its standard output and exit code.
func runKernelCommand(t *testing.T, stdin string, args ...string) (string, int) {
	// Install BusyBox with the applets the commands use.
	t.Helper()
	bb := busyboxWasm(t)
	files := BusyboxApplets(bb)
	for _, name := range []string{"ln", "readlink", "touch", "chmod", "date", "sleep", "basename", "dirname", "printf", "[", "seq", "id"} {
		files["/bin/"+name] = bb
	}

	// Start a kernel.
	var stdout bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	k, err := New(ctx, testRuntime(t, ctx), &Config{
		Stdout:     &stdout,
		Stderr:     &bytes.Buffer{},
		Stdin:      strings.NewReader(stdin),
		TerminalFD: -1,
		Files:      files,
		Env:        map[string]string{"PATH": "/bin:/usr/bin", "HOME": "/root", "TERM": "xterm"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer k.Close(ctx)

	// Run the command.
	err = k.RunProcess(ctx, bb, args...)
	var exitErr *ProcessExitError
	switch {
	case errors.As(err, &exitErr):
		return stdout.String(), exitErr.Code
	case err != nil:
		t.Fatal(err)
	}
	return stdout.String(), 0
}

// runDockerCommand runs a command line in an Alpine container and returns its
// standard output and exit code.
func runDockerCommand(t *testing.T, stdin string, args ...string) (string, int) {
	// Run the container.
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", append([]string{"run", "--rm", "-i", "alpine:latest"}, args...)...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stdin = strings.NewReader(stdin)
	err := cmd.Run()

	// Report its exit code.
	var exitErr *exec.ExitError
	switch {
	case errors.As(err, &exitErr):
		return stdout.String(), exitErr.ExitCode()
	case err != nil:
		t.Fatal(err)
	}
	return stdout.String(), 0
}

// normalizeOutput drops trailing whitespace from each line and trailing
// newlines from the end.
func normalizeOutput(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t\r")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}
