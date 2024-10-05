package proc

import (
	"fmt"
)

// assure interfaces are implemented at compile-time
var _ error = new(ExecuteError)
var _ error = new(NotInPathError)

// ExecuteError is an error during the execution of a subprocess
type ExecuteError struct {
	ExitCode int
	Err      error
}

func (e ExecuteError) Error() string {
	return fmt.Sprintf("exited with code: %d. error: %v", e.ExitCode, e.Err)
}

// NotInPathError means we couldn't find the executable somebody was trying to execute
// within a subprocess, in the system's PATH
type NotInPathError struct {
	Executable string
}

func (e NotInPathError) Error() string {
	return fmt.Sprintf("executable: %s was not found in system PATH", e.Executable)
}
