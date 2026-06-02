package models

// ProgressPhase identifies the execution phase.
type ProgressPhase string

const (
	ProgressPhaseBootstrap ProgressPhase = "bootstrap"
	ProgressPhaseDisable   ProgressPhase = "disable"
	ProgressPhaseClean     ProgressPhase = "clean"
	ProgressPhaseSecure    ProgressPhase = "secure"
	ProgressPhaseFinalize  ProgressPhase = "finalize"
)

// ProgressUpdate reports orchestrator progress for GUI callbacks.
type ProgressUpdate struct {
	Phase       ProgressPhase `json:"phase"`
	Module      string        `json:"module"`
	ModuleIndex int           `json:"module_index"`
	ModuleTotal int           `json:"module_total"`
	Message     string        `json:"message"`
	Complete    bool          `json:"complete"`
}

// ProgressFunc receives progress updates during execution.
type ProgressFunc func(update ProgressUpdate)
