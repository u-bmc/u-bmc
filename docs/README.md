# Documentation

## List of Contents

* [Architecture Overview](#architecture-overview)
  - [u-boot](#u-boot)
  - [Linux Kernel](#linux-kernel-and-initramfs)
  - [Operator](#operator)
* [Getting Started](#getting-started)
  - [Build](#build)
  - [Boot](#boot)
  - [Communicate](#communicate)
* [Microservices](#microservices)
  - [supervisord](#supervisord)
  - [registryd](#registryd)
  - [ipcd](#ipcd)
  - [netd](#netd)
  - [apid](#apid)
  - [hardwared](#hardwared)
  - [telemetryd](#telemetryd)
  - [kvmd](#kvmd)
  - [updated](#updated)
* [Interfaces](#interfaces)
  - [gRPC](#grpc)
  - [Redfish](#redfish)

## Architecture Overview

The architecture of the u-bmc is structured around four key components: u-boot, the Linux Kernel, Initramfs, and the Operator. Each component plays a crucial role in the boot and operation process of the BMC.

### u-boot

The first stage of the boot process is [u-boot](https://docs.u-boot.org/en/latest/index.html). It is the primary bootloader that is executed directly from the reset vector. Its primary responsibilities include:

- Activating the UBIFS (Unsorted Block Image File System) partition labeled 'boot' within the 'rootfs' MTD partition.
- On the first boot, u-boot prepares the UBIFS by creating a fastmap, which can lead to a longer boot time.
- It then loads the FIT (Flattened Image Tree) image from '/boot/fitImage' into RAM, a process crucial for booting the Linux Kernel.
- u-boot performs a secure boot of the FIT image by verifying baked-in signatures to ensure the image's integrity.
- Once the image is authenticated, u-boot copies the Linux Kernel and Device Tree Blob (DTB) to RAM and hands over execution to the Linux Kernel, providing it with the DTB.

### Linux Kernel and Initramfs

After u-boot has performed its initial setup, control is transferred to the Linux Kernel. Upon receiving control from u-boot, the Linux Kernel begins its startup sequence, which includes:

- Initializing and setting up device drivers and kernel subsystems.
- Transitioning to the initramfs, an in-memory temporary root file system included within the Kernel image.
- The init binary within initramfs then takes over, which is designed to be minimalistic to ensure a swift transition to the actual root file system.
- It optionally loads an authentication key into the kernel's session keyring if UBIFS authentication is enabled.
- The init process mounts all necessary pseudo filesystems (such as proc, sys, and dev) and the 'root' and 'data' UBIFS partitions.
- Finally, it performs the `switch_root` operation, transitioning from the initramfs to the actual root file system on the 'root' UBIFS partition.
- Handing over control to the u-bmc operator running with PID 1.

### Operator

Following the `switch_root` process, the new root file system takes over with /sbin/operator as the init process (PID 1):

- The operator binary encompasses the entire U-BMC userspace and represents the final step in the BMC's boot process.
- As PID 1, it orchestrates all subsequent operations within the BMC userspace.
- The operator handles all business logic within U-BMC, including service routines and external interfaces such as gRPC and Redfish served over HTTPS.
- It ensures that all management functions are performed, and services are running as expected, marking the end of the bootflow and the beginning of the operational state for the BMC.

## Getting Started

### Build

### Boot

### Communicate

## Microservices

### Supervisord

### Registryd

### Ipcd

### Netd

### Apid

### Hardwared

### Telemetryd

### Kvmd

### Updated

## Interfaces

### gRPC

### Redfish
