package opids_test

import (
	"testing"

	s4wave_git "github.com/s4wave/spacewave/core/git"
	"github.com/s4wave/spacewave/core/git/opids"
)

// TestCreateRepoWizardMatchesGenerated keeps the dependency-free identifier
// equal to the type ID generated from git.proto.
func TestCreateRepoWizardMatchesGenerated(t *testing.T) {
	if opids.CreateRepoWizard != s4wave_git.CreateGitRepoWizardOpTypeID {
		t.Fatalf("CreateRepoWizard = %q, want %q", opids.CreateRepoWizard, s4wave_git.CreateGitRepoWizardOpTypeID)
	}
}
