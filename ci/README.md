# Dagger CI Tool

This tool wraps any command into a Go container to guarantee a clean environment.
Any artifacts stored at 'output/' will be exported from the container and the tool is meant to be invoked from the repository root.
Example usage standalone:

```console
go run ci/main.go -cmd "go build -o output/"
```

Example usage with Dagger CLI:

```console
dagger run go run ci/main.go -cmd "go build -o output/"
```
