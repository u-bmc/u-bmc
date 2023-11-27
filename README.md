# u-bmc

## Build requirements

Either build natively on your host or using a predefined container environment.
In order to build in the container environment just three dependencies are required: 'go 1.21 or newer, docker or podman or any other container runtime with buildkit enabled and make'.

For the native variant all build tools are required that are otherwise available in the container. As we use Alpine Linux, the needed packages are right now are:
bc, binutils-cross-embedded, build-base, coreutils, cpio, dtc, fakeroot, findutils, gcc-cross-embedded, git, go, go-task-task, linux-headers, linux-lts-dev, lz4, mtd-utils-ubi, openssl-dev, openssl, upx, u-boot-tools, xz and zstd.

## Building the image

Once the dependencies are met, following two commands will produce a working flash image:

### Containerized

```console
make configure TARGET=$vendor/$platform (e.g. TARGET=asrock/paul)
make build
```

### Natively

```console
make configure TARGET=$vendor/$platform (e.g. TARGET=asrock/paul)
make build-native
```

The resulting flash image will be placed at "output/$vendor/$platform/flash.img
