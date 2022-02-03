package runcmd_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/lestrrat-go/runcmd"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var stdin bytes.Buffer
	environ := []string{"FOO=bar", "BAZ=quux"}

	dir, err := ioutil.TempDir("", "runcmd-test-*")
	if !assert.NoError(t, err, `ioutil.TempDir() should succeed`) {
		return
	}
	defer os.Remove(dir)

	ctx := runcmd.Context(context.Background()).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithStdin(&stdin).
		WithDir(dir).
		WithEnv(environ...)

	cmd, err := runcmd.Create(ctx, "ls")
	if !assert.NoError(t, err, `runcmd.Create should succeed`) {
		return
	}

	if !assert.Equal(t, cmd.Stdout, &stdout, `cmd.Stdout should match`) {
		return
	}
	if !assert.Equal(t, cmd.Stderr, &stderr, `cmd.Stderr should match`) {
		return
	}
	if !assert.Equal(t, cmd.Stdin, &stdin, `cmd.Stdin should match`) {
		return
	}
	if !assert.Equal(t, cmd.Dir, dir, `cmd.Dir should match`) {
		return
	}
	if !assert.Equal(t, cmd.Env, environ, `cmd.Env should match`) {
		return
	}
}
