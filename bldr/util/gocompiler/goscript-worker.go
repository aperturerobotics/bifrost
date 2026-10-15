package gocompiler

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/aperturerobotics/fastjson"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// goScriptWorkerEnv marks a process that ExecGoScriptCompile started to run
// one GoScript compile.
const goScriptWorkerEnv = "BLDR_GOSCRIPT_COMPILE_WORKER"

// init turns a process started with goScriptWorkerEnv into the compile
// worker before main runs. The worker reads its request as JSON from stdin,
// logs to stderr, writes a failure to stdout and exits.
func init() {
	if os.Getenv(goScriptWorkerEnv) != "" {
		os.Exit(serveGoScriptWorker())
	}
}

// runGoScriptWorker runs one GoScript compile in a child process of the
// current executable and returns its failure.
func runGoScriptWorker(ctx context.Context, le *logrus.Entry, opts GoScriptCompileOptions) error {
	// Start the current executable as a worker with the options' environment.
	exe, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "find the executable for the goscript worker")
	}
	cmd := exec.CommandContext(ctx, exe)
	cmd.Env = append(os.Environ(), opts.Env...)
	cmd.Env = append(cmd.Env, goScriptWorkerEnv+"=1")

	// Send the request, forward the worker's logs, and keep its failure.
	cmd.Stdin = bytes.NewReader(marshalGoScriptWorkerRequest(opts, le.Logger.GetLevel()))
	cmd.Stderr = le.Logger.Out
	var failure bytes.Buffer
	cmd.Stdout = &failure

	// Report the worker's own failure before its exit status.
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(failure.String()); msg != "" {
			return errors.New(msg)
		}
		return errors.Wrap(err, "run the goscript worker")
	}
	return nil
}

// serveGoScriptWorker compiles the request read from stdin and returns the
// process exit code.
func serveGoScriptWorker() int {
	// Keep the marker from the compiler's own child processes.
	_ = os.Unsetenv(goScriptWorkerEnv)

	// Compile the request, reporting a failure on stdout for the parent.
	if err := compileGoScriptRequest(os.Stdin); err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		return 1
	}
	return 0
}

// compileGoScriptRequest reads a worker request from r and compiles it,
// logging to stderr.
func compileGoScriptRequest(r io.Reader) error {
	// Read and decode the request.
	data, err := io.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "read the goscript worker request")
	}
	opts, level, err := parseGoScriptWorkerRequest(data)
	if err != nil {
		return err
	}

	// Compile at the parent's log level.
	log := logrus.New()
	log.SetLevel(level)
	return compileGoScript(context.Background(), logrus.NewEntry(log), opts)
}

// marshalGoScriptWorkerRequest encodes the compile options and log level for
// the worker.
func marshalGoScriptWorkerRequest(opts GoScriptCompileOptions, level logrus.Level) []byte {
	// Encode the log level and the scalar options.
	var a fastjson.Arena
	o := a.NewObject()
	o.Set("level", a.NewString(level.String()))
	o.Set("workDir", a.NewString(opts.WorkDir))
	o.Set("outputPath", a.NewString(opts.OutputPath))
	o.Set("cacheRoot", a.NewString(opts.CacheRoot))
	o.Set("protobufTypeScriptBinding", a.NewFalse())
	if opts.ProtobufTypeScriptBinding {
		o.Set("protobufTypeScriptBinding", a.NewTrue())
	}

	// Encode the list options.
	o.Set("packages", newStringArray(&a, opts.Packages))
	o.Set("buildFlags", newStringArray(&a, opts.BuildFlags))
	o.Set("overrideDirs", newStringArray(&a, opts.OverrideDirs))
	o.Set("bindingRoots", newStringArray(&a, opts.BindingRoots))
	o.Set("deferredFunctions", newStringArray(&a, opts.DeferredFunctions))
	return o.MarshalTo(nil)
}

// parseGoScriptWorkerRequest decodes a request from
// marshalGoScriptWorkerRequest. The worker's environment already holds the
// options' Env.
func parseGoScriptWorkerRequest(data []byte) (GoScriptCompileOptions, logrus.Level, error) {
	// Parse the request and its log level.
	var p fastjson.Parser
	v, err := p.ParseBytes(data)
	if err != nil {
		return GoScriptCompileOptions{}, 0, errors.Wrap(err, "parse the goscript worker request")
	}
	level, err := logrus.ParseLevel(string(v.GetStringBytes("level")))
	if err != nil {
		return GoScriptCompileOptions{}, 0, err
	}

	// Decode the options.
	return GoScriptCompileOptions{
		WorkDir:                   string(v.GetStringBytes("workDir")),
		OutputPath:                string(v.GetStringBytes("outputPath")),
		CacheRoot:                 string(v.GetStringBytes("cacheRoot")),
		Packages:                  getStringArray(v, "packages"),
		BuildFlags:                getStringArray(v, "buildFlags"),
		OverrideDirs:              getStringArray(v, "overrideDirs"),
		BindingRoots:              getStringArray(v, "bindingRoots"),
		DeferredFunctions:         getStringArray(v, "deferredFunctions"),
		ProtobufTypeScriptBinding: v.GetBool("protobufTypeScriptBinding"),
	}, level, nil
}

// newStringArray returns a JSON array of strs allocated in a.
func newStringArray(a *fastjson.Arena, strs []string) *fastjson.Value {
	arr := a.NewArray()
	for i, s := range strs {
		arr.SetArrayItem(i, a.NewString(s))
	}
	return arr
}

// getStringArray returns the strings of the array at key in v.
func getStringArray(v *fastjson.Value, key string) []string {
	// Leave an absent or empty array nil.
	items := v.GetArray(key)
	if len(items) == 0 {
		return nil
	}

	// Copy each string out of the parser's buffer.
	strs := make([]string, len(items))
	for i, item := range items {
		strs[i] = string(item.GetStringBytes())
	}
	return strs
}
