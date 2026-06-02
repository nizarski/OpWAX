package models

// ActionKind categorizes a planned or executed operation.
type ActionKind string

const (
	ActionDisable ActionKind = "disable"
	ActionClean   ActionKind = "clean"
	ActionSecure  ActionKind = "secure"
)

// Action describes a single planned operation (dry-run output).
type Action struct {
	Module      string     `json:"module"`
	Kind        ActionKind `json:"kind"`
	Description string     `json:"description"`
	Target      string     `json:"target"`
}

// Result reports outcome of an executed operation.
type Result struct {
	Action  Action `json:"action"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// DryRunReport aggregates all planned actions.
type DryRunReport struct {
	Actions []Action `json:"actions"`
}

// ExecutionReport aggregates execution results (in-memory only).
type ExecutionReport struct {
	Results      []Result              `json:"results"`
	Preflight    *PreflightReport      `json:"preflight,omitempty"`
	ManifestDiff *ManifestDiff         `json:"manifest_diff,omitempty"`
	Verification *VerificationReport   `json:"verification,omitempty"`
	NeedsReboot  bool                  `json:"needs_reboot"`
	RebootQueued bool                  `json:"reboot_queued"`
	Cancelled    bool                  `json:"cancelled"`
	CancelledAt  string                `json:"cancelled_at,omitempty"`
	SecondPassScheduled bool           `json:"second_pass_scheduled,omitempty"`
}

// VerificationGap is a location that should have been cleared but still has data.
type VerificationGap struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Detail   string `json:"detail"`
}

// VerificationReport summarizes post-run artifact checks.
type VerificationReport struct {
	Checked int               `json:"checked"`
	Clean   int               `json:"clean"`
	Gaps    []VerificationGap `json:"gaps,omitempty"`
}

// RunContext carries resolved runtime state through modules.
type RunContext struct {
	Config       Config
	CurrentUser  string
	TargetUsers  []UserProfile
	TargetDrives []DriveInfo
	WindowsDir   string
	ProgramData  string
}

// UserProfile represents a local user profile path.
type UserProfile struct {
	Username    string
	SID         string
	ProfilePath string
	NTUserPath  string
	UsrClassPath string
	AppDataRoaming string
	AppDataLocal   string
}

// DriveInfo represents a volume to process.
type DriveInfo struct {
	Letter     string
	Root       string
	IsSystem   bool
	FileSystem string
}
