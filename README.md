runcmd
======

Go utility to execute external commands easier. Intended to be used for
Go code that wants to execute commands in a shell-like style

```go
var buf bytes.Buffer

// Capture both stdout and stderr
ctx := runcmd.Context(context.Background()).
  WithStdout(&buf).
  WithStderr(&buf)

if err := runcmd.Run(ctx, "docker", "ps"); err != nil {
  ...
}
// You can keep reusing the same ctx to get the same
// option set
if err := runcmd.Run(ctx, "kubectl", "apply", "-f", "foo.yaml"); err != nil {
  ...
}
```
