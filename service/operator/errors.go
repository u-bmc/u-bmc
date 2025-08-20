package operator

import "errors"

var (
	ErrNameEmpty       = errors.New("name cannot be empty")
	ErrIPCNil          = errors.New("IPC cannot be nil, provide either ipcConn or WithIPC")
	ErrSetupMounts     = errors.New("failed to setup mounts")
	ErrAddProcess      = errors.New("failed to add process")
	ErrAddExtraService = errors.New("failed to add extra service")
	ErrPanicked        = errors.New("panicked")
)
