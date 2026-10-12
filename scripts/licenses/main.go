// Command licenses reports the license text of the Go dependencies of the
// module in the working directory as JSON.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-licenses/v2/licenses"
)

// entry supplies the module deduplication and attribution generator.
type entry struct {
	// Name identifies the dependency covered by this license.
	Name string `json:"name"`
	// Version identifies the selected module revision.
	Version string `json:"version"`
	// LicenseName is the detected SPDX identifier.
	LicenseName string `json:"licenseName"`
	// LicenseText preserves the complete attribution text.
	LicenseText string `json:"licenseText"`
}

func main() {
	if err := report(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// report emits one complete report or fails without partial JSON.
func report() error {
	// Classify every dependency of the main modules.
	base, err := licenses.NewClassifier()
	if err != nil {
		return err
	}
	own, err := mainModules()
	if err != nil {
		return err
	}
	libs, err := licenses.Libraries(context.Background(), classifier{base}, false, own, os.Args[1:]...)
	if err != nil {
		return err
	}

	// Keep the actual license text, including copyright, for each library.
	entries := make([]entry, 0, len(libs))
	for _, lib := range libs {
		if lib.LicenseFile == "" || len(lib.Licenses) == 0 {
			return fmt.Errorf("no recognized license for %s", lib.Name())
		}
		data, err := os.ReadFile(lib.LicenseFile)
		if err != nil {
			return err
		}
		for _, license := range lib.Licenses {
			entries = append(entries, entry{lib.Name(), lib.Version(), license.Name, string(data)})
		}
	}
	return json.NewEncoder(os.Stdout).Encode(entries)
}

// mainModules lists the modules being built, whose own packages are not
// dependencies.
func mainModules() ([]string, error) {
	out, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return nil, fmt.Errorf("go list -m: %w", err)
	}
	return strings.Fields(string(out)), nil
}
