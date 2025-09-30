// SPDX-License-Identifier: BSD-3-Clause

package ipc

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go/micro"
)

// IPC Subject Constants for NATS Micro Services
// These constants define all the subjects used for inter-process communication.
// Services should use these constants rather than constructing subjects dynamically.

// State Management Service Subjects
const (
	// Host management
	SubjectHostState   = "host.state"
	SubjectHostControl = "host.control"
	SubjectHostInfo    = "host.info"
	SubjectHostList    = "host.list"

	// Chassis management
	SubjectChassisState   = "chassis.state"
	SubjectChassisControl = "chassis.control"
	SubjectChassisInfo    = "chassis.info"
	SubjectChassisList    = "chassis.list"

	// BMC management
	SubjectBMCState   = "bmc.state"
	SubjectBMCControl = "bmc.control"
	SubjectBMCInfo    = "bmc.info"
	SubjectBMCList    = "bmc.list"
)

// Inventory Management Service Subjects
const (
	// Asset management
	SubjectAssetInfo   = "asset.info"
	SubjectAssetCreate = "asset.create"
	SubjectAssetUpdate = "asset.update"
	SubjectAssetDelete = "asset.delete"
	SubjectAssetList   = "asset.list"
)

// User Management Service Subjects
const (
	// User management
	SubjectUserCreate         = "user.create"
	SubjectUserInfo           = "user.info"
	SubjectUserUpdate         = "user.update"
	SubjectUserDelete         = "user.delete"
	SubjectUserList           = "user.list"
	SubjectUserChangePassword = "user.change_password"
	SubjectUserResetPassword  = "user.reset_password"
	SubjectUserAuthenticate   = "user.authenticate"
)

// Security Management Service Subjects
const (
	// Certificate management
	SubjectCertificateInfo   = "certificate.info"
	SubjectCertificateCreate = "certificate.create"
	SubjectCertificateUpdate = "certificate.update"
	SubjectCertificateDelete = "certificate.delete"
	SubjectCertificateList   = "certificate.list"

	// Key management
	SubjectKeyInfo   = "key.info"
	SubjectKeyCreate = "key.create"
	SubjectKeyUpdate = "key.update"
	SubjectKeyDelete = "key.delete"
	SubjectKeyList   = "key.list"

	// Policy management
	SubjectPolicyInfo   = "policy.info"
	SubjectPolicyCreate = "policy.create"
	SubjectPolicyUpdate = "policy.update"
	SubjectPolicyDelete = "policy.delete"
	SubjectPolicyList   = "policy.list"
)

// Sensor Monitoring Service Subjects
const (
	// Sensor management
	SubjectSensorInfo = "sensor.info"
	SubjectSensorList = "sensor.list"
)

// Thermal Management Service Subjects
const (
	// Thermal zone management
	SubjectThermalZoneInfo = "thermal_zone.info"
	SubjectThermalZoneSet  = "thermal_zone.set"
	SubjectThermalZoneList = "thermal_zone.list"
)

// System Information Service Subjects
const (
	// System information
	SubjectSystemInfo   = "system.info"
	SubjectSystemHealth = "system.health"
)

// Power Management Service Subjects (for coordination)
const (
	// Power control
	SubjectPowerAction = "power.action"
	SubjectPowerResult = "power.result"
	SubjectPowerStatus = "power.status"
)

// LED Management Service Subjects (for coordination)
const (
	// LED control
	SubjectLEDControl = "led.control"
	SubjectLEDStatus  = "led.status"
)

// Event and Notification Subjects
const (
	// State events
	SubjectStateEvent      = "state.event"
	SubjectTransitionEvent = "transition.event"

	// System events
	SubjectSystemEvent = "system.event"
	SubjectAlertEvent  = "alert.event"
)

// Stream Subjects for JetStream Persistence
const (
	// State persistence streams
	StreamSubjectStateChanges    = "statemgr.state.>"
	StreamSubjectEvents          = "statemgr.event.>"
	StreamSubjectSystemEvents    = "system.event.>"
	StreamSubjectSecurityEvents  = "security.event.>"
	StreamSubjectInventoryEvents = "inventory.event.>"
)

// Internal IPC Subjects (for service-to-service communication)
const (
	// Power manager coordination
	InternalPowerAction = "internal.power.action"
	InternalPowerResult = "internal.power.result"

	// LED manager coordination
	InternalLEDControl = "internal.led.control"
	InternalLEDStatus  = "internal.led.status"

	// Sensor data propagation
	InternalSensorData = "internal.sensor.data"

	// Thermal management coordination
	InternalThermalControl = "internal.thermal.control"
	InternalThermalStatus  = "internal.thermal.status"
)

// Queue Groups for Load Balancing
const (
	// Service queue groups
	QueueGroupStateManager   = "statemgr"
	QueueGroupInventory      = "inventorymgr"
	QueueGroupUserManager    = "usermgr"
	QueueGroupSecurity       = "securitymgr"
	QueueGroupSensorMonitor  = "sensormon"
	QueueGroupThermalManager = "thermalmgr"
	QueueGroupPowerManager   = "powermgr"
	QueueGroupLEDManager     = "ledmgr"
)

// Default Timeouts (in milliseconds)
const (
	DefaultRequestTimeout  = 30000 // 30 seconds
	DefaultCommandTimeout  = 60000 // 60 seconds
	DefaultStreamTimeout   = 5000  // 5 seconds
	DefaultResponseTimeout = 10000 // 10 seconds
)

// Error Response Subjects
const (
	// Standard error responses
	SubjectErrorResponse   = "error.response"
	SubjectTimeoutResponse = "timeout.response"
	SubjectInvalidRequest  = "invalid.request"
	SubjectUnauthorized    = "unauthorized.request"
	SubjectNotFound        = "not.found"
	SubjectInternalError   = "internal.error"
)

// IPC Error Constants
var (
	// Request/Response errors
	ErrMissingRequiredField = NewIPCError("MISSING_REQUIRED_FIELD", "missing required field")
	ErrMarshalingFailed     = NewIPCError("MARSHALING_FAILED", "marshaling failed")
	ErrUnmarshalingFailed   = NewIPCError("UNMARSHALING_FAILED", "unmarshaling failed")
	ErrResponseTimeout      = NewIPCError("RESPONSE_TIMEOUT", "response timeout")

	// Component errors
	ErrComponentNotFound     = NewIPCError("COMPONENT_NOT_FOUND", "component not found")
	ErrInvalidTrigger        = NewIPCError("INVALID_TRIGGER", "invalid trigger")
	ErrStateTransitionFailed = NewIPCError("STATE_TRANSITION_FAILED", "state transition failed")

	// Service errors
	ErrInternalError = NewIPCError("INTERNAL_ERROR", "internal error")
)

// IPCError represents a structured IPC error
type IPCError struct {
	Code    string
	Message string
}

func (e *IPCError) Error() string {
	return e.Message
}

// NewIPCError creates a new IPC error
func NewIPCError(code, message string) *IPCError {
	return &IPCError{
		Code:    code,
		Message: message,
	}
}

// ParseSubject splits a subject into group and endpoint components for NATS micro registration.
// For subjects like "host.state", it returns group="host" and endpoint="state".
// Returns an error if the subject doesn't contain exactly one dot or if components are empty.
func ParseSubject(subject string) (group, endpoint string, err error) {
	if subject == "" {
		return "", "", NewIPCError("INVALID_SUBJECT", "subject cannot be empty")
	}

	parts := strings.Split(subject, ".")
	if len(parts) != 2 {
		return "", "", NewIPCError("INVALID_SUBJECT", fmt.Sprintf("subject %s must contain exactly one dot", subject))
	}

	group = strings.TrimSpace(parts[0])
	endpoint = strings.TrimSpace(parts[1])

	if group == "" {
		return "", "", NewIPCError("INVALID_SUBJECT", "group component cannot be empty")
	}

	if endpoint == "" {
		return "", "", NewIPCError("INVALID_SUBJECT", "endpoint component cannot be empty")
	}

	return group, endpoint, nil
}

// RegisterEndpointWithParsedSubject is a helper function that parses an IPC subject
// and returns the group and endpoint names for use with NATS micro registration.
// This ensures services use IPC constants consistently and follow the group.endpoint pattern.
//
// Example usage:
//
//	group, endpoint, err := ipc.RegisterEndpointWithParsedSubject(ipc.SubjectHostState)
//	if err != nil {
//	    return err
//	}
//	hostGroup := service.AddGroup(group)
//	return hostGroup.AddEndpoint(endpoint, handler)
func RegisterEndpointWithParsedSubject(subject string) (group, endpoint string, err error) {
	return ParseSubject(subject)
}

// RegisterEndpointWithGroupCache registers an endpoint by parsing the IPC subject and managing group creation.
// This helper reduces boilerplate by automatically creating and caching groups as needed.
//
// Example usage:
//
//	groups := make(map[string]micro.Group)
//	err := ipc.RegisterEndpointWithGroupCache(service, ipc.SubjectHostState, handler, groups)
func RegisterEndpointWithGroupCache(service micro.Service, subject string, handler micro.Handler, groups map[string]micro.Group) error {
	groupName, endpointName, err := ParseSubject(subject)
	if err != nil {
		return fmt.Errorf("failed to parse subject %s: %w", subject, err)
	}

	// Get or create group
	group, exists := groups[groupName]
	if !exists {
		group = service.AddGroup(groupName)
		groups[groupName] = group
	}

	// Register endpoint
	if err := group.AddEndpoint(endpointName, handler); err != nil {
		return fmt.Errorf("failed to register endpoint %s in group %s: %w", endpointName, groupName, err)
	}

	return nil
}
