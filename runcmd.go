// Package runcmd provides utilities that can be used to construct
// calls to external commands similar to how commands are used in
// shell scripts.

package runcmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type identEnv struct{}
type identDir struct{}
type identStdout struct{}
type identStderr struct{}
type identStdin struct{}
type identExtraFiles struct{}

type Ctx interface {
	context.Context
	// WithStdout specifies where the stdout of the command is redireted to
	WithStdout(io.Writer) Ctx

	// WithStderr specifies where the stderr of the command is redireted to
	WithStderr(io.Writer) Ctx

	// WithStdin specifies where the stdin of the command is received from
	WithStdin(io.Reader) Ctx

	// WithDir specifies the working directory for the command
	WithDir(string) Ctx

	// WithExtraFiles specifies the files that should be inherited by the command
	WithExtraFiles(...*os.File) Ctx

	// WithEnv specifies the list of environment variables that should be
	// enabled in the command
	WithEnv(...string) Ctx
}

type rcCtx struct {
	context.Context
}

// Context creates a new runcmd.Ctx object
func Context(ctx context.Context) Ctx {
	return &rcCtx{ctx}
}

func (ctx *rcCtx) WithExtraFiles(files ...*os.File) Ctx {
	ctx.Context = context.WithValue(ctx.Context, identExtraFiles{}, files)
	return ctx
}

func (ctx *rcCtx) WithEnv(environ ...string) Ctx {
	ctx.Context = context.WithValue(ctx.Context, identEnv{}, environ)
	return ctx
}

func (ctx *rcCtx) WithDir(s string) Ctx {
	ctx.Context = context.WithValue(ctx.Context, identDir{}, s)
	return ctx
}

func (ctx *rcCtx) WithStdout(v io.Writer) Ctx {
	ctx.Context = context.WithValue(ctx.Context, identStdout{}, v)
	return ctx
}

func (ctx *rcCtx) WithStderr(v io.Writer) Ctx {
	ctx.Context = context.WithValue(ctx.Context, identStderr{}, v)
	return ctx
}

func (ctx *rcCtx) WithStdin(v io.Reader) Ctx {
	ctx.Context = context.WithValue(ctx.Context, identStdin{}, v)
	return ctx
}

func getWriter(ctx context.Context, dst *io.Writer, key interface{}, name string) error {
	tmp := ctx.Value(key)
	if tmp == nil {
		return nil
	}

	v, ok := tmp.(io.Writer)
	if !ok {
		return fmt.Errorf(`expected io.Writer for %s, got %T`, name, tmp)
	}
	*dst = v
	return nil
}

func getReader(ctx context.Context, dst *io.Reader, key interface{}, name string) error {
	tmp := ctx.Value(key)
	if tmp == nil {
		return nil
	}

	v, ok := tmp.(io.Reader)
	if !ok {
		return fmt.Errorf(`expected io.Reader for %s, got %T`, name, tmp)
	}
	*dst = v
	return nil
}

func getString(ctx context.Context, dst *string, key interface{}, name string) error {
	tmp := ctx.Value(key)
	if tmp == nil {
		return nil
	}
	v, ok := tmp.(string)
	if !ok {
		return fmt.Errorf(`expected string for %s, got %T`, name, tmp)
	}
	*dst = v
	return nil
}

func getStringSlice(ctx context.Context, dst *[]string, key interface{}, name string) error {
	tmp := ctx.Value(key)
	if tmp == nil {
		return nil
	}
	v, ok := tmp.([]string)
	if !ok {
		return fmt.Errorf(`expected string for %s, got %T`, name, tmp)
	}
	*dst = v
	return nil
}

// Run is a simple wrapper around (exec.Command).Run(). The intent is to
// use almost like a command executed in a shell script.
//
// By default output is sent to os.Stdout asnd os.Stderr, and similarly
// the command's input is set to os.Stdin. Other fields in "os/exec".Cmd
// are not set.
//
// If you would like to configure the command further, you will have to
// pass runcmd.Ctx object as the first argument. To do this, create a
// runcmd.Ctx object and use the various `WithXXX()` methods with it.
func Run(ctx context.Context, path string, args ...string) error {
	cmd, err := Create(ctx, path, args...)
	if err != nil {
		return fmt.Errorf(`failed to create *exec.Cmd: %w`, err)
	}

	return cmd.Run()
}

func Create(ctx context.Context, path string, args ...string) (*exec.Cmd, error) {
	var stdin io.Reader = os.Stdin
	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr
	var dir string
	var environ []string
	if err := getWriter(ctx, &stdout, identStdout{}, "Stdout"); err != nil {
		return nil, fmt.Errorf(`failed to assign Stdout: %w`, err)
	}
	if err := getWriter(ctx, &stderr, identStderr{}, "Stderr"); err != nil {
		return nil, fmt.Errorf(`failed to assign Stderr: %w`, err)
	}
	if err := getReader(ctx, &stdin, identStdin{}, "Stdin"); err != nil {
		return nil, fmt.Errorf(`failed to assign Stdin: %w`, err)
	}
	if err := getString(ctx, &dir, identDir{}, "Dir"); err != nil {
		return nil, fmt.Errorf(`failed to assign Dir: %w`, err)
	}
	if err := getStringSlice(ctx, &environ, identEnv{}, "Env"); err != nil {
		return nil, fmt.Errorf(`failed to assign Env: %w`, err)
	}

	cmd := exec.CommandContext(ctx, path, args...)
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	if stdin != nil {
		cmd.Stdin = stdin
	}
	if dir != "" {
		cmd.Dir = dir
	}
	if environ != nil {
		cmd.Env = environ
	}

	return cmd, nil
}
