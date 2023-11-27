// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

type Bootloader struct {
	Type         string   `toml:"type"`
	Source       string   `toml:"source"`
	FetchCmd     string   `toml:"fetch_cmd"`
	ExtraEnable  []string `toml:"extra_enable"`
	ExtraDisable []string `toml:"extra_disable"`
	Config       string   `toml:"config"`
	CustomDT     bool     `toml:"custom_dt"`
	DT           string   `toml:"dt"`
	Concurrency  int      `toml:"concurrency"`
	Clang        bool     `toml:"clang"`
}

type Kernel struct {
	Type          string   `toml:"type"`
	Source        string   `toml:"source"`
	FetchCmd      string   `toml:"fetch_cmd"`
	ExtraEnable   []string `toml:"extra_enable"`
	ExtraDisable  []string `toml:"extra_disable"`
	ExtraModule   []string `toml:"extra_module"`
	Config        string   `toml:"config"`
	CustomDT      bool     `toml:"custom_dt"`
	DT            string   `toml:"dt"`
	Concurrency   int      `toml:"concurrency"`
	Clang         bool     `toml:"clang"`
	InitrdGoFlags string   `toml:"initrd_goflags"`
}

type Rootfs struct {
	URoot bool `toml:"uroot"`
}

type Image struct {
	FlashSize        string `toml:"flash_size"`
	BootloaderOffset string `toml:"bootloader_offset"`
	RootfsOffset     string `toml:"rootfs_offset"`
	ABScheme         bool   `toml:"ab_scheme"`
}

type Config struct {
	Platform   string     `toml:"platform"`
	Arch       string     `toml:"arch"`
	SoC        string     `toml:"soc"`
	Vendor     string     `toml:"vendor"`
	Toolchain  string     `toml:"toolchain"`
	Bootloader Bootloader `toml:"bootloader"`
	Kernel     Kernel     `toml:"kernel"`
	Rootfs     Rootfs     `toml:"rootfs"`
	Image      Image      `toml:"image"`
}

func main() {
	platform := flag.String("platform", "", "Platform target of form 'vendor/name'")
	flag.Parse()

	if *platform == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	config := Config{
		Bootloader: Bootloader{
			Concurrency: runtime.NumCPU(),
		},
		Kernel: Kernel{
			Concurrency: runtime.NumCPU(),
		},
	}

	if _, err := toml.DecodeFile(filepath.Join("configs", *platform, "config.toml"), &config); err != nil {
		log.Fatalln(err)
	}

	for _, kind := range []string{"bootloader", "kernel", "rootfs", "image"} {
		if err := decodeTemplate(*platform, kind, config); err != nil {
			log.Fatalln(err)
		}
	}

	if err := copyFile(filepath.Join("output", *platform, "Taskfile.yml"), filepath.Join("tools", "configure", "templates", "Taskfile.yml.tmpl")); err != nil {
		log.Fatalln(err)
	}

	if err := os.WriteFile(filepath.Join("output", "TARGET"), []byte(config.Platform), os.ModePerm); err != nil {
		log.Fatalln(err)
	}
}

func decodeTemplate(platform, kind string, config Config) error { //nolint:cyclop
	if err := filepath.WalkDir(filepath.Join("tools", "configure", "templates", kind), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return err
		}

		result := filepath.Join("output", platform, kind, strings.TrimSuffix(d.Name(), ".tmpl"))

		if err := os.MkdirAll(filepath.Dir(result), os.ModePerm); err != nil {
			return err
		}

		f, err := os.Create(result)
		if err != nil {
			return err
		}
		defer f.Close()

		return tmpl.Execute(f, config)
	}); err != nil {
		return err
	}

	switch kind {
	case "bootloader":
		if err := copyFile(filepath.Join("output", platform, kind, "defconfig"), filepath.Join("configs", config.Bootloader.Config)); err != nil {
			return err
		}
	case "kernel":
		if err := copyFile(filepath.Join("output", platform, kind, "u-bmc.dts"), filepath.Join("configs", config.Kernel.DT)); err != nil {
			return err
		}

		if err := copyFile(filepath.Join("output", platform, kind, "dts-makefile"), filepath.Join("configs", filepath.Dir(config.Kernel.DT), "dts-makefile")); err != nil {
			return err
		}

		if err := copyFile(filepath.Join("output", platform, kind, "defconfig"), filepath.Join("configs", config.Kernel.Config)); err != nil {
			return err
		}
	case "rootfs":
		for _, f := range []string{"main.go", "go.mod", "go.sum"} {
			if err := copyFile(filepath.Join("output", platform, kind, f), filepath.Join("configs", config.Platform, f)); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(dst, src string) error {
	sp, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	sf, err := os.Open(sp)
	if err != nil {
		return err
	}
	defer sf.Close()

	dp, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dp), os.ModePerm); err != nil {
		return err
	}

	df, err := os.Create(dp)
	if err != nil {
		return err
	}
	defer df.Close()

	if _, err := io.Copy(df, sf); err != nil {
		return err
	}

	return nil
}

// TODO: Might be needed for u-boot to merge configs
// func copyAndAppendFile(dst string, src ...string) error {
// 	var sc []byte

// 	for _, s := range src {
// 		sp, err := filepath.Abs(s)
// 		if err != nil {
// 			return err
// 		}

// 		b, err := os.ReadFile(sp)
// 		if err != nil {
// 			return err
// 		}

// 		sc = append(sc, []byte("\n")...)
// 		sc = append(sc, b...)
// 	}

// 	dp, err := filepath.Abs(dst)
// 	if err != nil {
// 		return err
// 	}

// if err := os.MkdirAll(filepath.Dir(dp), os.ModePerm); err != nil {
// 	return err
// }

// 	return os.WriteFile(dp, sc, os.ModePerm)
// }
