//go:build !tinygo

package space_world

import (
	"context"
	"strings"

	bldr_manifest "github.com/s4wave/spacewave/bldr/manifest"
	bldr_manifest_builder "github.com/s4wave/spacewave/bldr/manifest/builder"
	"github.com/s4wave/spacewave/bldr/manifest/builder/resultworld"
	bldr_manifest_world "github.com/s4wave/spacewave/bldr/manifest/world"
	forge_dashboard "github.com/s4wave/spacewave/core/forge/dashboard"
	"github.com/s4wave/spacewave/db/block/byteslice"
	"github.com/s4wave/spacewave/db/blocktype"
	git_block "github.com/s4wave/spacewave/db/git/block"
	git_world "github.com/s4wave/spacewave/db/git/world"
	kvtx_block "github.com/s4wave/spacewave/db/kvtx/block"
	forge_cluster "github.com/s4wave/spacewave/forge/cluster"
	forge_execution "github.com/s4wave/spacewave/forge/execution"
	forge_job "github.com/s4wave/spacewave/forge/job"
	forge_pass "github.com/s4wave/spacewave/forge/pass"
	forge_task "github.com/s4wave/spacewave/forge/task"
	forge_worker "github.com/s4wave/spacewave/forge/worker"
	"github.com/s4wave/spacewave/identity"
	identity_world "github.com/s4wave/spacewave/identity/world"
	s4wave_appconnector "github.com/s4wave/spacewave/sdk/appconnector"
	s4wave_apt "github.com/s4wave/spacewave/sdk/apt"
	s4wave_canvas "github.com/s4wave/spacewave/sdk/canvas"
	s4wave_chat "github.com/s4wave/spacewave/sdk/chat"
	s4wave_device "github.com/s4wave/spacewave/sdk/device"
	s4wave_org "github.com/s4wave/spacewave/sdk/org"
	s4wave_secret "github.com/s4wave/spacewave/sdk/secret"
	s4wave_sshhost "github.com/s4wave/spacewave/sdk/sshhost"
	s4wave_terminal "github.com/s4wave/spacewave/sdk/terminal"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	s4wave_wizard "github.com/s4wave/spacewave/sdk/world/wizard"
)

// applicationBlockTypes supplies the same persisted roots used by application
// resources. Plugin-specific roots continue through their registered controller.
var applicationBlockTypes = map[string]blocktype.BlockType{
	identity_world.KeypairTypeID:             blocktype.New[identity.Keypair](identity_world.KeypairTypeID),
	bldr_manifest_world.ManifestTypeID:       blocktype.New[bldr_manifest.Manifest](bldr_manifest_world.ManifestTypeID),
	bldr_manifest_world.ManifestBundleTypeID: blocktype.New[bldr_manifest.ManifestBundle](bldr_manifest_world.ManifestBundleTypeID),
	resultworld.ManifestBuildResultTypeID:    blocktype.New[bldr_manifest_builder.BuilderResult](resultworld.ManifestBuildResultTypeID),
	forge_dashboard.ForgeDashboardTypeID:     blocktype.New[forge_dashboard.ForgeDashboard](forge_dashboard.ForgeDashboardTypeID),
	forge_cluster.ClusterTypeID:              blocktype.New[forge_cluster.Cluster](forge_cluster.ClusterTypeID),
	forge_execution.ExecutionTypeID:          blocktype.New[forge_execution.Execution](forge_execution.ExecutionTypeID),
	forge_job.JobTypeID:                      blocktype.New[forge_job.Job](forge_job.JobTypeID),
	forge_pass.PassTypeID:                    blocktype.New[forge_pass.Pass](forge_pass.PassTypeID),
	forge_task.TaskTypeID:                    blocktype.New[forge_task.Task](forge_task.TaskTypeID),
	forge_worker.WorkerTypeID:                blocktype.New[forge_worker.Worker](forge_worker.WorkerTypeID),
	s4wave_apt.AptRepositoryTypeID:           blocktype.New[s4wave_apt.AptRepository](s4wave_apt.AptRepositoryTypeID),
	s4wave_apt.AptPackageTypeID:              blocktype.New[s4wave_apt.AptPackage](s4wave_apt.AptPackageTypeID),
	s4wave_apt.AptBuildSpecTypeID:            blocktype.New[s4wave_apt.AptBuildSpec](s4wave_apt.AptBuildSpecTypeID),
	"canvas":                                 blocktype.New[s4wave_canvas.CanvasStorage]("canvas"),
	"kv/store":                               blocktype.New[kvtx_block.KeyValueStore]("kv/store"),
	git_world.GitRepoTypeID:                  blocktype.New[git_block.Repo](git_world.GitRepoTypeID),
	git_world.GitWorktreeTypeID:              blocktype.New[git_world.Worktree](git_world.GitWorktreeTypeID),
	s4wave_chat.ChatChannelTypeID:            blocktype.New[s4wave_chat.ChatChannel](s4wave_chat.ChatChannelTypeID),
	s4wave_chat.ChatMessageTypeID:            blocktype.New[s4wave_chat.ChatMessage](s4wave_chat.ChatMessageTypeID),
	s4wave_chat.ChatMessagePageTypeID:        blocktype.New[s4wave_chat.ChatMessagePage](s4wave_chat.ChatMessagePageTypeID),
	s4wave_chat.ChatThreadTypeID:             blocktype.New[s4wave_chat.ChatThread](s4wave_chat.ChatThreadTypeID),
	s4wave_device.DeviceTypeID:               blocktype.New[s4wave_device.Device](s4wave_device.DeviceTypeID),
	s4wave_device.ComputersDashboardTypeID:   blocktype.New[s4wave_device.ComputersDashboard](s4wave_device.ComputersDashboardTypeID),
	s4wave_org.OrganizationTypeID:            blocktype.New[s4wave_org.OrgState](s4wave_org.OrganizationTypeID),
	s4wave_appconnector.AppConnectorTypeID:   blocktype.New[s4wave_appconnector.AppConnector](s4wave_appconnector.AppConnectorTypeID),
	s4wave_appconnector.AppSnapshotTypeID:    blocktype.New[s4wave_appconnector.AppSnapshot](s4wave_appconnector.AppSnapshotTypeID),
	s4wave_secret.SecretTypeID:               blocktype.New[s4wave_secret.Secret](s4wave_secret.SecretTypeID),
	s4wave_sshhost.SshHostTypeID:             blocktype.New[s4wave_sshhost.SshHost](s4wave_sshhost.SshHostTypeID),
	s4wave_terminal.TerminalTypeID:           blocktype.New[s4wave_terminal.Terminal](s4wave_terminal.TerminalTypeID),
	s4wave_vm.VmV86TypeID:                    blocktype.New[s4wave_vm.VmV86](s4wave_vm.VmV86TypeID),
}

// lookupApplicationBlockType resolves application object roots for block traversal.
func lookupApplicationBlockType(_ context.Context, typeID string) (blocktype.BlockType, error) {
	// SDK app bindings are JSON leaves. Collections and immutable executables
	// have their own typed World objects, so decoding never needs plugin code.
	if strings.HasPrefix(typeID, "sync/app/") {
		return blocktype.New[byteslice.ByteSlice](typeID), nil
	}
	if strings.HasPrefix(typeID, "wizard/") {
		return blocktype.New[s4wave_wizard.WizardState](typeID), nil
	}
	return applicationBlockTypes[typeID], nil
}
