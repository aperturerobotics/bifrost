//go:build !js

package spacewave_cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aperturerobotics/cli"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/s4wave/spacewave/db/s4db"
	volume_kvtx "github.com/s4wave/spacewave/db/volume/common/kvtx"
)

// debugVolumeReindexArgs regenerates a stopped volume's reference graph indexes.
type debugVolumeReindexArgs struct{}

// BuildFlags returns the flags accepted by the offline reindex command.
func (a *debugVolumeReindexArgs) BuildFlags() []cli.Flag {
	return nil
}

// Run rebuilds postings from the primitive log without altering the graph's edges.
func (a *debugVolumeReindexArgs) Run(c *cli.Context) error {
	// Require one existing volume and exclude another process's access.
	if c.NArg() != 1 {
		return errors.New("expected one volume file argument")
	}
	path := c.Args().First()
	if _, err := os.Stat(path); err != nil {
		return err
	}

	// Open the database directly because normal volume opens require rebuilt indexes.
	db, err := s4db.Open(path, s4db.Options{})
	if err != nil {
		return errors.Wrap(err, "open volume database")
	}
	defer db.Close()
	shared, err := db.Shared()
	if err != nil {
		return err
	}
	if shared {
		return errors.Errorf("%s is open in another process; stop the daemon first", path)
	}

	// Regenerate the complete posting layout while the volume stays offline.
	// Batches commit with write ordering: each records where a rerun
	// resumes, so a crash can only rewind to an earlier batch, and one flush
	// at the end makes the rebuild durable.
	count, err := volume_kvtx.RebuildVolumeRefGraph(c.Context, kvtx.NewOrderedStore(db), reindexProgress())
	if err == nil {
		err = db.Sync(c.Context)
	}
	if err != nil {
		return errors.Wrap(err, "rebuild reference indexes; rerun before starting the daemon")
	}
	writeFields(c.App.Writer, [][2]string{
		{"Volume", path},
		{"References Indexed", strconv.FormatUint(count, 10)},
	})
	return nil
}

// reindexProgress returns a progress function that reports each whole
// percent of the primitive log indexed on standard error.
func reindexProgress() func(next, horizon uint64) {
	last := -1
	return func(next, horizon uint64) {
		pct := int((next - 1) * 100 / max(horizon, 1))
		if pct == last {
			return
		}
		last = pct
		fmt.Fprintf(os.Stderr, "indexed %d of %d primitives (%d%%)\n", next-1, horizon, pct)
	}
}

// newDebugVolumeReindexCommand builds the offline reference index regenerator.
func newDebugVolumeReindexCommand() *cli.Command {
	args := &debugVolumeReindexArgs{}
	return &cli.Command{
		Name:      "volume-reindex",
		Usage:     "regenerate a stopped volume's reference graph indexes",
		ArgsUsage: "<volume-file>",
		Description: "Regenerates GC reference postings from the volume's primitive log. " +
			"Preserves blocks, live reference edges, and node reference counts. " +
			"Stop the daemon and keep a copy of the volume before running it. " +
			"If interrupted, rerun this command before starting the daemon.",
		Flags:  args.BuildFlags(),
		Action: args.Run,
	}
}
