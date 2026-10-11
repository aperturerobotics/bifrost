package sobject_world_engine

import (
	"time"

	"github.com/sirupsen/logrus"
)

// slowCommitThreshold is the duration of a World write, from its writer lock
// wait to the end of its Commit, after which Commit logs its stage times. It
// is the duration after which a World call is reported as slow. Tests set it.
var slowCommitThreshold = 10 * time.Second

// commitStage names a stage of a World write.
type commitStage int

const (
	// stageLockWait waits for the writer lock.
	stageLockWait commitStage = iota
	// stageForkReplay replays the operation set and forks the candidate.
	stageForkReplay
	// stageCallback is the time the caller holds the writer between opening
	// the write and calling Commit.
	stageCallback
	// stageWorldStateCommit closes the operation buffer.
	stageWorldStateCommit
	// stageBlockCommit commits the candidate World blocks.
	stageBlockCommit
	// stagePublicationAdmission prepares the operation and waits until the
	// SharedObject admits it.
	stagePublicationAdmission
	// stageReceiptWait waits for the publication to be durable.
	stageReceiptWait
	// stageAdvance replays the operation set and installs the World.
	stageAdvance
	// commitStageCount is the number of stages.
	commitStageCount
)

// commitStageNames are the log field names of the stages.
var commitStageNames = [commitStageCount]string{
	stageLockWait:             "lock-wait",
	stageForkReplay:           "fork-replay",
	stageCallback:             "callback",
	stageWorldStateCommit:     "world-state-commit",
	stageBlockCommit:          "block-commit",
	stagePublicationAdmission: "publication-admission",
	stageReceiptWait:          "receipt-wait",
	stageAdvance:              "advance",
}

// commitClock times the stages of one World write with a clock read at each
// stage boundary. A nil commitClock records nothing.
type commitClock struct {
	// started is when the write began waiting for the writer lock.
	started time.Time
	// mark is the end of the last closed stage.
	mark time.Time
	// open is the earliest stage not yet closed.
	open commitStage
	// stages are the durations of the closed stages.
	stages [commitStageCount]time.Duration
}

// newCommitClock starts the clock at the writer lock wait.
func newCommitClock() *commitClock {
	now := time.Now()
	return &commitClock{started: now, mark: now}
}

// end closes stage at the current time. Earlier open stages that never ran
// close with no duration, and closing a closed stage does nothing.
func (c *commitClock) end(stage commitStage) {
	// Ignore a stage that is already closed.
	if c == nil || stage < c.open {
		return
	}

	// Attribute the time since the last boundary to the stage.
	now := time.Now()
	c.stages[stage] = now.Sub(c.mark)
	c.mark = now
	c.open = stage + 1
}

// warnIfSlow logs the stage times when the write took longer than
// slowCommitThreshold.
func (c *commitClock) warnIfSlow(le *logrus.Entry) {
	// Stay quiet about a write that finished in time.
	if c == nil || le == nil {
		return
	}
	elapsed := time.Since(c.started)
	if elapsed <= slowCommitThreshold {
		return
	}

	// Name the duration of every stage.
	fields := logrus.Fields{"elapsed": elapsed}
	for stage, name := range commitStageNames {
		fields[name] = c.stages[stage]
	}
	le.WithFields(fields).Warn("slow World commit")
}
