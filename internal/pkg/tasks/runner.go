package tasks

//Runner executes external process and manages it
type Runner struct {
}

// Close terminates runnig preocess if any
func (r *Runner) Close() error {
	return nil
}

// Run starts the preocess
func (r *Runner) Run(cmd string, env []string) error {
	return nil
}

// Running returns the running status
func (r *Runner) Running() bool {
	return false
}
