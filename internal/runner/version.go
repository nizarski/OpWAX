package runner

import (
	"fmt"
	"os"

	"github.com/opwax/opwax/internal/version"
)

// PrintVersion writes the application banner to stdout and exits.
func PrintVersion() {
	fmt.Println(version.Banner())
	os.Exit(0)
}
