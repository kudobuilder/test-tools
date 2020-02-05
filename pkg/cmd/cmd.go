// Package cmd implements a simple wrapper around "os/exec".
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Builder tracks the options set for a command.
type Builder struct {
	Command     string
	Arguments   []string
	Environment []string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
}

// New creates a new command.
// Additional command options are added to the builder returned by New.
//   err := cmd.New("echo").
//   	WithArguments("Hello, World!").
//   	WithStdout(os.Stdout).
//   	Run()
func New(command string) Builder {
	return Builder{
		Command: command,
	}
}

// Run runs a command.
func (builder Builder) Run() error {
	cmd := exec.Command(builder.Command, builder.Arguments...) //nolint:gosec
	builder.prepareCmd(cmd)

	return cmd.Run()
}

// RunWithContext runs a command with a context.
func (builder Builder) RunWithContext(ctx context.Context) error {
	cmd := exec.Command(builder.Command, builder.Arguments...) //nolint:gosec
	builder.prepareCmd(cmd)

	return cmd.Run()
}

// WithArguments adds arguments to a command.
func (builder Builder) WithArguments(arguments ...string) Builder {
	builder.Arguments = arguments
	return builder
}

// WithEnvironment adds environment variables to a command.
func (builder Builder) WithEnvironment(environment map[string]string) Builder {
	for key, value := range environment {
		builder.Environment = append(
			builder.Environment,
			fmt.Sprintf("%s=%s", key, value))
	}

	return builder
}

// WithStdin sets an io.Reader to use as input to the command.
func (builder Builder) WithStdin(stdin io.Reader) Builder {
	builder.Stdin = stdin
	return builder
}

// WithStdout sets an io.Writer to retrieve the output of the command.
func (builder Builder) WithStdout(stdout io.Writer) Builder {
	builder.Stdout = stdout
	return builder
}

// WithStderr sets an io.Writer to retrieve the error output of the command.
func (builder Builder) WithStderr(stderr io.Writer) Builder {
	builder.Stderr = stderr
	return builder
}

func (builder Builder) prepareCmd(cmd *exec.Cmd) {
	cmd.Env = append(os.Environ(), builder.Environment...)

	if builder.Stdin != nil {
		cmd.Stdin = builder.Stdin
	}

	if builder.Stdout != nil {
		cmd.Stdout = builder.Stdout
	}

	if builder.Stderr != nil {
		cmd.Stderr = builder.Stderr
	}
}
