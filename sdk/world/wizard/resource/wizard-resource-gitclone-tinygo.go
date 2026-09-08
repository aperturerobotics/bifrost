//go:build tinygo

package wizard_resource

import (
	"context"

	wizard "github.com/s4wave/spacewave/sdk/world/wizard"
)

func (r *WizardResource) runGitClone(ctx context.Context, req *wizard.StartGitCloneRequest) {
	if err := ctx.Err(); err != nil {
		r.setGitCloneProgress(&wizard.GitCloneProgress{
			State:     wizard.GitCloneProgressState_GIT_CLONE_PROGRESS_STATE_FAILED,
			Message:   "Clone canceled.",
			Error:     err.Error(),
			ObjectKey: req.GetObjectKey(),
		})
		return
	}
	r.setGitCloneProgress(&wizard.GitCloneProgress{
		State:     wizard.GitCloneProgressState_GIT_CLONE_PROGRESS_STATE_FAILED,
		Message:   "Git clone is not available in this browser build.",
		ObjectKey: req.GetObjectKey(),
	})
}
