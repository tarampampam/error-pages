package checkers

// LiveChecker is a liveness checker.
type LiveChecker struct{}

// NewLiveChecker creates liveness checker.
func NewLiveChecker() *LiveChecker { return &LiveChecker{} }

// Check application is alive?
func (*LiveChecker) Check() error { return nil }
