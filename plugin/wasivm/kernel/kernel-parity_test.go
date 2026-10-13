package kernel

import "testing"

// parityApplets lists the BusyBox applets the parity table runs beyond those
// BusyboxApplets installs.
var parityApplets = []string{
	"free", "ln", "ps", "readlink", "setsid", "stat", "yes",
}

// TestLinuxParity checks BusyBox scripts print the same output and exit with
// the same code under the kernel as on Linux. The outputs were captured once
// from Alpine's static BusyBox 1.37 run as PID 1 of a new PID namespace, with
// the environment the test sets. Each script creates the files it reads.
func TestLinuxParity(t *testing.T) {
	tests := []struct {
		name, script, stdin, want string
		exit                      int
	}{
		{name: "echo", script: `echo hello world`, want: "hello world\n"},
		{name: "pipe", script: `echo HELLO | tr A-Z a-z`, want: "hello\n"},
		{name: "multi-pipe", script: `echo abc | cat | cat`, want: "abc\n"},
		{name: "cat-file", script: `echo testdata > /tmp/f && cat /tmp/f`, want: "testdata\n"},
		{name: "cwd", script: `cd /tmp && pwd`, want: "/tmp\n"},
		{name: "mkdir", script: `mkdir -p /tmp/d/e && ls /tmp/d`, want: "e\n"},
		{name: "append", script: `echo line1 > /tmp/a && echo line2 >> /tmp/a && cat /tmp/a`, want: "line1\nline2\n"},
		{name: "status", script: `false; echo $?`, want: "1\n"},
		{name: "exit-code", script: `exit 7`, want: "", exit: 7},
		{name: "pipe-status", script: `false | true; echo $?`, want: "0\n"},
		{name: "printf", script: `printf '%s %d\n' hello 42`, want: "hello 42\n"},
		{name: "seq", script: `seq 1 5`, want: "1\n2\n3\n4\n5\n"},
		{name: "arithmetic", script: `echo $((2 + 3 * 4))`, want: "14\n"},
		{name: "variables", script: `X=hello; echo $X world`, want: "hello world\n"},
		{name: "uname", script: `uname -s`, want: "Linux\n"},
		{name: "test", script: `if [ 1 -eq 1 ]; then echo yes; else echo no; fi`, want: "yes\n"},
		{name: "grep", script: `printf 'apple\nbanana\ncherry\n' | grep an`, want: "banana\n"},
		{name: "stdin", script: `cat`, stdin: "line1\nline2\nline3\n", want: "line1\nline2\nline3\n"},
		{name: "read", script: `read a b; echo "$b $a"`, stdin: "one two\n", want: "two one\n"},
		{name: "substitution", script: `echo $(echo nested)`, want: "nested\n"},
		{name: "rename", script: `echo data > /tmp/src && mv /tmp/src /tmp/dst && cat /tmp/dst`, want: "data\n"},
		{name: "here-doc", script: "cat <<EOF\nhello\nworld\nEOF", want: "hello\nworld\n"},
		{name: "symlink", script: `ln -s /tmp/target /tmp/link && readlink /tmp/link`, want: "/tmp/target\n"},
		{name: "paths", script: `basename /a/b/c.txt .txt; dirname /a/b/c.txt`, want: "c\n/a/b\n"},
		{name: "wc", script: `printf 'a b\nc\n' | wc -l -w -c`, want: "        2         3         6\n"},
		{name: "sort", script: `printf 'b\nc\na\n' | sort -r`, want: "c\nb\na\n"},
		{name: "umask", script: `umask 027; umask; touch /tmp/m; stat -c %a /tmp/m`, want: "0027\n640\n"},
		{name: "chmod", script: `touch /tmp/x; chmod 4750 /tmp/x; chmod o+r /tmp/x; stat -c '%a %A' /tmp/x`, want: "4754 -rwsr-xr--\n"},
		{name: "awk-system", script: `awk 'BEGIN { system("echo from-system"); "echo piped" | getline x; print x }'`, want: "from-system\npiped\n"},
		{name: "shebang", script: `printf '#!/bin/sh -e\necho "$0 $1"\n' >/tmp/s; chmod +x /tmp/s; /tmp/s arg`, want: "/tmp/s arg\n"},
		{name: "exec-noperm", script: `echo 'echo run' >/tmp/n; /tmp/n 2>/dev/null; echo $?; chmod +x /tmp/n; /tmp/n`, want: "126\nrun\n"},
		{name: "id", script: `id -u; id -g`, want: "0\n0\n"},
		{name: "yes-head", script: `yes | head -3`, want: "y\ny\ny\n"},
		{name: "pid-1", script: `echo $$`, want: "1\n"},
		{name: "ppid", script: `p=$$; sh -c '[ $PPID = '$p' ] && echo child of shell'; true`, want: "child of shell\n"},
		{name: "subshell-pid", script: `a=$$; b=$(echo $$); [ $a = $b ] && echo same`, want: "same\n"},
		{name: "kill-init", script: `kill $$; echo init survives`, want: "init survives\n"},
		{name: "kill-child", script: `sh -c 'kill $$; echo no'; echo $?`, want: "143\n"},
		{name: "sigint-status", script: `sh -c 'kill -INT $$'; echo $?`, want: "130\n"},
		{name: "trap", script: `sh -c 'trap "echo caught" TERM; kill $$; echo after'`, want: "caught\nafter\n"},
		{name: "background-wait", script: `sleep 0 & wait $!; echo $?`, want: "0\n"},
		{name: "background-status", script: `(exit 3) & wait $!; echo $?`, want: "3\n"},
		{name: "killed-background", script: `sleep 10 & kill $!; wait $!; echo $?`, want: "143\n"},
		{name: "exec", script: `exec echo replaced; echo not`, want: "replaced\n"},
		{name: "session", script: `setsid sh -c 'ps -o pid,sid | awk "NR > 1 && \$1 == \$2" | wc -l'`, want: "1\n"},
		{name: "proc-self", script: `readlink /proc/self/exe >/dev/null && echo exe; [ -r /proc/self/status ] && echo status`, want: "exe\nstatus\n"},
		{name: "ps-header", script: `ps -o pid,ppid,comm | head -1`, want: "PID   PPID  COMMAND\n"},
		{name: "free-header", script: `free | head -1`, want: "              total        used        free      shared  buff/cache   available\n"},
		{name: "dev-null", script: `echo gone > /dev/null; cat /dev/null | wc -c`, want: "0\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Install the applets the scripts use.
			t.Parallel()
			bb := busyboxWasm(t)
			files := make(map[string][]byte, len(parityApplets))
			for _, name := range parityApplets {
				files["/bin/"+name] = bb
			}

			// Run the script and compare its output with Linux.
			st := shellTest{files: files, script: tc.script, stdin: tc.stdin, exit: tc.exit}
			if got, stderr := st.run(t); got != tc.want {
				t.Errorf("stdout %q, want %q, stderr: %q", got, tc.want, stderr)
			}
		})
	}
}
