// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"dagger.io/dagger"
)

const GoVer = "1.22"

var (
	ErrUnableToConnect = errors.New("unable to connect to client")
	ErrGetPwd          = errors.New("unable to get current work dir")
	ErrUnableToRun     = errors.New("unable to run pipeline")
	ErrNoTarget        = errors.New("no target found")
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	native := flag.Bool("native", false, "Run the Taskfiles natively, assuming the host has all the dependencies installed")
	flag.Parse()

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetPwd, err)
	}

	targetFilePath := filepath.Join(pwd, "output", "TARGET")
	if stat, err := os.Stat(targetFilePath); err != nil || !stat.Mode().IsRegular() {
		return fmt.Errorf("%w: %w", ErrNoTarget, err)
	}

	t, err := os.ReadFile(targetFilePath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNoTarget, err)
	}

	targetPath := filepath.Join(pwd, "output", string(t))

	if *native {
		return runNative(targetPath)
	}

	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnableToConnect, err)
	}
	defer client.Close()

	src := client.Host().Directory(targetPath)

	_, err = client.Container().
		From(fmt.Sprintf("golang:%s-alpine", GoVer)).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{
			"apk",
			"add",
			"--no-cache",
			"bc",
			"binutils-cross-embedded",
			"build-base",
			"coreutils",
			"cpio",
			"dtc",
			"fakeroot",
			"findutils",
			"gcc-cross-embedded",
			"git",
			"go",
			"go-task-task",
			"linux-headers",
			"linux-lts-dev",
			"lz4",
			"mtd-utils-ubi",
			"openssl-dev",
			"openssl",
			"upx",
			"u-boot-tools",
			"xz",
			"zstd",
		}).
		WithExec([]string{
			"ln", "-sf", "/bin/bash", "/bin/sh",
		}).
		WithExec([]string{
			"task", "build",
		}).
		File("/src/flash.img").
		Export(ctx, filepath.Join(targetPath, "flash.img"))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnableToRun, err)
	}

	return nil
}

func runNative(path string) error {
	cmd := exec.Command("task", "build")
	cmd.Dir = path
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %w", ErrUnableToRun, err)
	}

	return nil
}
