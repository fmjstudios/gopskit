package errors

import "fmt"

type errExecute struct {
	ExitCode int
	Err      error
}

func Execute(code int, err error) error {
	return &errExecute{
		ExitCode: code,
		Err:      err,
	}
}

// implement the Error interface
func (e *errExecute) Error() string {
	return fmt.Sprintf("exited with code: %d. error: %v", e.ExitCode, e.Err)
}

type errNotInPATH struct {
	Executable string
}

func NotInPATH(exec string) error {
	return &errNotInPATH{
		Executable: exec,
	}
}

func (e *errNotInPATH) Error() string {
	return fmt.Sprintf("executable: %s was not found in system PATH", e.Executable)
}
