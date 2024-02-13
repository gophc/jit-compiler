//go:build !amd64 && !arm64

package lib

// init initializes variables for the unsupported architecture
func init() {
	newArchContext = newArchContextImpl
}

// archContext is empty on an unsupported architecture.
type archContext struct{}

// newArchContextImpl implements newArchContext for amd64 architecture.
func newArchContextImpl() (ret archContext) { return }
