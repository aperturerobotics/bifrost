//go:build !js

package spacewave_cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aperturerobotics/cli"
	"github.com/dustin/go-humanize"
	"github.com/s4wave/spacewave/db/s4db"
	store_kvkey "github.com/s4wave/spacewave/db/store/kvkey"
	volume_kvtx "github.com/s4wave/spacewave/db/volume/common/kvtx"
)

// copyReportInterval is the least time between progress lines.
const copyReportInterval = 10 * time.Second

// newDebugVolumeCopyCommand builds the offline volume copy command.
func newDebugVolumeCopyCommand() *cli.Command {
	return &cli.Command{
		Name:      "volume-copy",
		Usage:     "copy a stopped volume file into a new file at the current store format",
		ArgsUsage: "<volume-file> <new-file>",
		Description: "Copies every key of a stopped s4db volume into a new file, converting a Cayley reference graph " +
			"to edge records on the way, so the source is never rewritten in place. The new file is written in key order " +
			"with compaction off and records the format version last; a copy that stops early lacks it and should be deleted. " +
			"The source is left unchanged. Replace the source with the new file only after checking it.",
		Action: func(c *cli.Context) error {
			if c.NArg() != 2 {
				return errors.New("expected the volume file and the new file")
			}
			return copyVolumeFile(c.Context, c.Args().Get(0), c.Args().Get(1), os.Stderr)
		},
	}
}

// copyVolumeFile copies the stopped volume at src into the new file dst,
// writing progress lines to w. It removes dst when the copy fails.
func copyVolumeFile(ctx context.Context, src, dst string, w io.Writer) (err error) {
	// Open the source, refusing a missing file, which the open would create,
	// and a file a daemon has open.
	if _, err := os.Stat(src); err != nil {
		return err
	}
	sdb, err := s4db.Open(src, s4db.Options{})
	if err != nil {
		return err
	}
	defer sdb.Close()
	shared, err := sdb.Shared()
	if err != nil {
		return err
	}
	if shared {
		return fmt.Errorf("%s is open in another process; stop the daemon first", src)
	}

	// Create the destination, which must not exist. Its keys arrive in order
	// and never move, so compaction has nothing to do.
	if _, err := os.Stat(dst); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s already exists", dst)
	}
	ddb, err := s4db.Open(dst, s4db.Options{RelocateBudget: -1})
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, ddb.Close())
		if err != nil {
			err = errors.Join(err, os.Remove(dst))
		}
	}()

	// Copy, reporting progress at most once per interval.
	start := time.Now()
	var last time.Time
	report := func(p volume_kvtx.CopyProgress) {
		if time.Since(last) < copyReportInterval {
			return
		}
		last = time.Now()
		fmt.Fprintf(w, "%s: %d keys, %s, %d graph edges\n", time.Since(start).Round(time.Second), p.Keys, humanize.IBytes(p.Bytes), p.Edges)
	}
	if err := volume_kvtx.CopyStore(ctx, sdb, ddb, store_kvkey.NewDefaultKVKey(), report); err != nil {
		return err
	}
	fmt.Fprintf(w, "copied %s to %s in %s\n", src, dst, time.Since(start).Round(time.Second))
	return nil
}
