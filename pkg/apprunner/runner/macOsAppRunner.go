package runner

import "errors"

type macOsAppRunner struct{}

func GetMacOsAppRunner() macOsAppRunner {
	f := macOsAppRunner{}

	return f
}

func (lr *macOsAppRunner) Run(command string) error {
	// TODO:

	return errors.New("MacOs is not suport")
}
