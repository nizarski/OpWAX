package version

const (
	Name    = "OpWAX"
	Version = "1.3.0"
	Author  = "nizarski"
)

// Banner returns a single-line identity string for CLI and logs.
func Banner() string {
	return Name + " " + Version + " - built by " + Author
}
