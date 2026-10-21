//go:build !js

package devtool

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/aperturerobotics/cli"
	"github.com/pkg/errors"
	bldr_callpath "github.com/s4wave/spacewave/bldr/devtool/callpath"
)

// callpathArgs are the flags of the callpath command.
type callpathArgs struct {
	from   string
	to     string
	within string
	tests  bool
	full   bool
}

// BuildCallpathCommand builds the callpath command.
func BuildCallpathCommand() *cli.Command {
	args := &callpathArgs{}
	return &cli.Command{
		Name:      "callpath",
		Usage:     "print static call paths from matching functions to matching targets",
		ArgsUsage: "<package-pattern>...",
		Description: "Builds a class hierarchy call graph of the packages, so an interface call\n" +
			"reaches every method that could satisfy it, and prints the shortest path from\n" +
			"each function matching --from to each function matching --to. Paths pass\n" +
			"only through the loaded packages' modules unless --within is set; a target\n" +
			"may lie anywhere. Names are SSA names: pkg.Func, (*pkg.T).Method, pkg.Func$1.\n\n" +
			"Example, every World op that can take a mutex:\n" +
			"  callpath --from '\\)\\.ApplyWorldOp$' --to '^\\(\\*sync\\.(RW)?Mutex\\)\\.R?Lock$' ./core/...",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "from",
				Usage:       "regexp matching the functions paths start from",
				Required:    true,
				Destination: &args.from,
			},
			&cli.StringFlag{
				Name:        "to",
				Usage:       "regexp matching the functions paths end at",
				Required:    true,
				Destination: &args.to,
			},
			&cli.StringFlag{
				Name:        "within",
				Usage:       "regexp matching the package paths a path may pass through",
				Destination: &args.within,
			},
			&cli.BoolFlag{
				Name:        "tests",
				Usage:       "also load the packages' test files",
				Destination: &args.tests,
			},
			&cli.BoolFlag{
				Name:        "full",
				Usage:       "print full package paths instead of trimming the loaded modules",
				Destination: &args.full,
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return errors.New("usage: callpath --from <regexp> --to <regexp> <package-pattern>...")
			}
			return args.run(os.Stdout, c.Args().Slice())
		},
	}
}

// run searches the packages and writes one block per path to w.
func (a *callpathArgs) run(w io.Writer, patterns []string) error {
	// Compile the patterns.
	opts := bldr_callpath.Options{Patterns: patterns, Tests: a.tests}
	var err error
	if opts.From, err = regexp.Compile(a.from); err != nil {
		return errors.Wrap(err, "from")
	}
	if opts.To, err = regexp.Compile(a.to); err != nil {
		return errors.Wrap(err, "to")
	}
	if a.within != "" {
		if opts.Within, err = regexp.Compile(a.within); err != nil {
			return errors.Wrap(err, "within")
		}
	}

	// Search and print the paths.
	res, err := bldr_callpath.Find(opts)
	if err != nil {
		return err
	}
	trim := func(name string) string { return name }
	if !a.full {
		var pairs []string
		for _, mod := range res.Modules {
			pairs = append(pairs, mod+"/", "")
		}
		trim = strings.NewReplacer(pairs...).Replace
	}
	for _, path := range res.Paths {
		for i, name := range path {
			if i == 0 {
				fmt.Fprintln(w, trim(name))
				continue
			}
			fmt.Fprintln(w, "  -> "+trim(name))
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "%d paths from %d roots\n", len(res.Paths), res.Roots)
	return nil
}
