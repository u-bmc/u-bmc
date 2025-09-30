// SPDX-License-Identifier: BSD-3-Clause

package ipc

import (
	"testing"
)

func TestParseSubject(t *testing.T) {
	tests := []struct {
		name             string
		subject          string
		expectedGroup    string
		expectedEndpoint string
		expectError      bool
	}{
		{
			name:             "valid host subject",
			subject:          "host.state",
			expectedGroup:    "host",
			expectedEndpoint: "state",
			expectError:      false,
		},
		{
			name:             "valid chassis subject",
			subject:          "chassis.control",
			expectedGroup:    "chassis",
			expectedEndpoint: "control",
			expectError:      false,
		},
		{
			name:             "valid user subject",
			subject:          "user.create",
			expectedGroup:    "user",
			expectedEndpoint: "create",
			expectError:      false,
		},
		{
			name:             "valid thermal zone subject",
			subject:          "thermal_zone.info",
			expectedGroup:    "thermal_zone",
			expectedEndpoint: "info",
			expectError:      false,
		},
		{
			name:        "empty subject",
			subject:     "",
			expectError: true,
		},
		{
			name:        "no dot",
			subject:     "hoststate",
			expectError: true,
		},
		{
			name:        "multiple dots",
			subject:     "host.state.extra",
			expectError: true,
		},
		{
			name:        "empty group",
			subject:     ".state",
			expectError: true,
		},
		{
			name:        "empty endpoint",
			subject:     "host.",
			expectError: true,
		},
		{
			name:        "whitespace group",
			subject:     "  .state",
			expectError: true,
		},
		{
			name:        "whitespace endpoint",
			subject:     "host.  ",
			expectError: true,
		},
		{
			name:             "subject with underscores",
			subject:          "user.change_password",
			expectedGroup:    "user",
			expectedEndpoint: "change_password",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, endpoint, err := ParseSubject(tt.subject)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseSubject(%q) expected error, got none", tt.subject)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSubject(%q) unexpected error: %v", tt.subject, err)
				return
			}

			if group != tt.expectedGroup {
				t.Errorf("ParseSubject(%q) group = %q, want %q", tt.subject, group, tt.expectedGroup)
			}

			if endpoint != tt.expectedEndpoint {
				t.Errorf("ParseSubject(%q) endpoint = %q, want %q", tt.subject, endpoint, tt.expectedEndpoint)
			}
		})
	}
}

func TestRegisterEndpointWithParsedSubject(t *testing.T) {
	tests := []struct {
		name             string
		subject          string
		expectedGroup    string
		expectedEndpoint string
		expectError      bool
	}{
		{
			name:             "valid subject",
			subject:          SubjectHostState,
			expectedGroup:    "host",
			expectedEndpoint: "state",
			expectError:      false,
		},
		{
			name:             "valid user subject",
			subject:          SubjectUserCreate,
			expectedGroup:    "user",
			expectedEndpoint: "create",
			expectError:      false,
		},
		{
			name:        "invalid subject",
			subject:     "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, endpoint, err := RegisterEndpointWithParsedSubject(tt.subject)

			if tt.expectError {
				if err == nil {
					t.Errorf("RegisterEndpointWithParsedSubject(%q) expected error, got none", tt.subject)
				}
				return
			}

			if err != nil {
				t.Errorf("RegisterEndpointWithParsedSubject(%q) unexpected error: %v", tt.subject, err)
				return
			}

			if group != tt.expectedGroup {
				t.Errorf("RegisterEndpointWithParsedSubject(%q) group = %q, want %q", tt.subject, group, tt.expectedGroup)
			}

			if endpoint != tt.expectedEndpoint {
				t.Errorf("RegisterEndpointWithParsedSubject(%q) endpoint = %q, want %q", tt.subject, endpoint, tt.expectedEndpoint)
			}
		})
	}
}

func TestIPCSubjectConstants(t *testing.T) {
	// Test that all IPC subject constants can be parsed correctly
	subjects := []string{
		SubjectHostState,
		SubjectHostControl,
		SubjectHostInfo,
		SubjectHostList,
		SubjectChassisState,
		SubjectChassisControl,
		SubjectChassisInfo,
		SubjectChassisList,
		SubjectBMCState,
		SubjectBMCControl,
		SubjectBMCInfo,
		SubjectBMCList,
		SubjectAssetInfo,
		SubjectAssetCreate,
		SubjectAssetUpdate,
		SubjectAssetDelete,
		SubjectAssetList,
		SubjectUserCreate,
		SubjectUserInfo,
		SubjectUserUpdate,
		SubjectUserDelete,
		SubjectUserList,
		SubjectUserChangePassword,
		SubjectUserResetPassword,
		SubjectUserAuthenticate,
		SubjectSensorInfo,
		SubjectSensorList,
		SubjectThermalZoneInfo,
		SubjectThermalZoneSet,
		SubjectThermalZoneList,
		SubjectSystemInfo,
		SubjectSystemHealth,
		SubjectPowerAction,
		SubjectPowerResult,
		SubjectPowerStatus,
		SubjectLEDControl,
		SubjectLEDStatus,
	}

	for _, subject := range subjects {
		t.Run(subject, func(t *testing.T) {
			group, endpoint, err := ParseSubject(subject)
			if err != nil {
				t.Errorf("ParseSubject(%q) failed: %v", subject, err)
				return
			}

			if group == "" {
				t.Errorf("ParseSubject(%q) returned empty group", subject)
			}

			if endpoint == "" {
				t.Errorf("ParseSubject(%q) returned empty endpoint", subject)
			}

			// Verify that reconstructing the subject works
			reconstructed := group + "." + endpoint
			if reconstructed != subject {
				t.Errorf("ParseSubject(%q) -> %q + %q = %q, want %q",
					subject, group, endpoint, reconstructed, subject)
			}
		})
	}
}
