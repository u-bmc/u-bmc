// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

var (
	ErrTooFewArgs      = errors.New("too few arguments")
	ErrUnableToConnect = errors.New("unable to connect to client")
	ErrGetPwd          = errors.New("unable to get current work dir")
	ErrMkdir           = errors.New("unable to create directory")
	ErrUnableToRun     = errors.New("unable to run pipeline")
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	cmd := flag.String("cmd", "", "The command to be executed inside the golang container environment")
	flag.Parse()
	if len(*cmd) == 0 {
		return ErrTooFewArgs
	}

	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnableToConnect, err)
	}
	defer client.Close()

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetPwd, err)
	}

	src := client.Host().Directory(pwd)

	if err := os.MkdirAll(filepath.Join(pwd, "output"), os.ModePerm); err != nil {
		return fmt.Errorf("%w: %w", ErrMkdir, err)
	}

	if ok, err := client.
		Container().
		From("golang:latest").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"mkdir", "-p", "/src/output"}).
		WithExec(strings.Split(*cmd, " ")).
		Directory("/src/output").
		Export(ctx, filepath.Join(pwd, "output")); !ok || err != nil {
		return fmt.Errorf("%w: %w", ErrUnableToRun, err)
	}

	return nil
}
