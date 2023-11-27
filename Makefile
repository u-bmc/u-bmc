TOOLS_PATH=./tools

.PHONY: configure build build-native clean help

# Target for setting up the project with a user-supplied target.
configure:
	@go run $(TOOLS_PATH)/configure -platform=$(TARGET)

# Target for building the project.
build:
	@go run $(TOOLS_PATH)/build

# Target for building the project.
build-native:
	@go run $(TOOLS_PATH)/build -native

# Target for cleaning up the build directory.
clean:
	@rm -rf build

# Target for displaying help message.
help:
	@echo "Available commands:"
	@echo "  configure TARGET=<NAME>: Set up the project for a user-supplied target"
	@echo "  build: Build the project, requires setup to be run first"
	@echo "  build-native: Build the project without a container runtime, requires setup to be run first and dependencies to be installed"
	@echo "  clean: Clean up the build directory"
