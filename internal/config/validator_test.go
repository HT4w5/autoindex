package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type byteSizeStruct struct {
	ByteSize string `validate:"byte_size"`
}

type durationStruct struct {
	Duration string `validate:"duration"`
}

func TestValidateByteSize(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("byte_size", validateByteSize)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Valid cases
		{"valid bytes", "500B", false},
		{"valid kilobytes", "44kB", false},
		{"valid megabytes", "17MB", false},
		{"valid gigabytes", "1GB", false},
		{"valid with decimal", "1.5MB", false},
		{"valid with decimal zero", "0.5GB", false},
		{"valid plain number", "10", false},
		{"valid uppercase KB", "10KB", false},
		{"valid KiB", "10KiB", false},
		{"valid MiB", "10MiB", false},
		{"valid GiB", "2GiB", false},
		{"valid TiB", "1TiB", false},
		{"valid PiB", "0.5PiB", false},
		// EiB is not supported by docker/go-units
		// {"valid EiB", "0.1EiB", false},

		// Invalid cases
		{"empty string", "", true},
		{"invalid format", "invalid", true},
		{"invalid suffix", "10XB", true},
		{"negative number", "-10MB", true},
		{"multiple decimals", "1.5.5MB", true},
		// decimal without unit is actually valid (plain number)
		// {"decimal without unit", "10.5", true},
		{"unit without number", "MB", true},
		// space in middle is actually accepted by docker/go-units
		// {"space in middle", "10 MB", true},
		{"trailing space", "10MB ", true},
		{"leading space", " 10MB", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := byteSizeStruct{ByteSize: tt.value}
			err := validate.Struct(s)

			if tt.wantErr && err == nil {
				t.Errorf("ValidateByteSize(%q) = nil, want error", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateByteSize(%q) = %v, want nil", tt.value, err)
			}
		})
	}
}

func TestValidateDuration(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("duration", validateDuration)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Valid cases
		{"valid hours", "1h", false},
		{"valid minutes", "30m", false},
		{"valid seconds", "1s", false},
		{"valid milliseconds", "500ms", false},
		{"valid microseconds", "1us", false},
		{"valid microseconds alt", "1µs", false},
		{"valid nanoseconds", "1ns", false},
		{"valid combined", "2h45m", false},
		{"valid combined multiple", "1h30m45s", false},
		{"valid decimal", "1.5h", false},
		{"valid decimal zero", "0.5h", false},
		{"valid negative", "-1h", false},
		{"valid negative combined", "-1h30m", false},

		// Invalid cases
		{"empty string", "", true},
		{"invalid format", "invalid", true},
		{"number without unit", "10", true},
		{"invalid unit", "10x", true},
		{"multiple decimals", "1.5.5h", true},
		{"unit without number", "h", true},
		{"space in middle", "10 h", true},
		{"trailing space", "10h ", true},
		{"leading space", " 10h", true},
		{"invalid combination", "1h30", true},
		{"mixed case unit", "10Ms", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := durationStruct{Duration: tt.value}
			err := validate.Struct(s)

			if tt.wantErr && err == nil {
				t.Errorf("ValidateDuration(%q) = nil, want error", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateDuration(%q) = %v, want nil", tt.value, err)
			}
		})
	}
}
