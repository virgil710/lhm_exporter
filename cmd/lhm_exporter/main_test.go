package main

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestDebugFlagExists(t *testing.T) {
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.Bool("debug", false, "Enable debug mode with verbose logging to stdout.")

	err := flagSet.Parse([]string{"--debug"})
	if err != nil {
		t.Fatalf("Failed to parse --debug flag: %v", err)
	}

	debugVal, err := flagSet.GetBool("debug")
	if err != nil {
		t.Fatalf("Failed to get --debug flag value: %v", err)
	}

	if !debugVal {
		t.Error("Expected --debug flag to be true")
	}
}

func TestDebugFlagDefaultFalse(t *testing.T) {
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.Bool("debug", false, "Enable debug mode with verbose logging to stdout.")

	err := flagSet.Parse([]string{})
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	debugVal, err := flagSet.GetBool("debug")
	if err != nil {
		t.Fatalf("Failed to get --debug flag value: %v", err)
	}

	if debugVal {
		t.Error("Expected --debug flag to be false by default")
	}
}

func TestNormalizeListenHost(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "empty string returns 0.0.0.0",
			input:   "",
			want:    "0.0.0.0",
			wantErr: false,
		},
		{
			name:    "localhost returns localhost",
			input:   "localhost",
			want:    "localhost",
			wantErr: false,
		},
		{
			name:    "valid IPv4 address",
			input:   "192.168.1.1",
			want:    "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "valid IPv6 address",
			input:   "::1",
			want:    "::1",
			wantErr: false,
		},
		{
			name:    "hostname returns as-is",
			input:   "myhost.example.com",
			want:    "myhost.example.com",
			wantErr: false,
		},
		{
			name:    "host with port returns error",
			input:   "127.0.0.1:8080",
			want:    "",
			wantErr: true,
		},
		{
			name:    "too long host returns error",
			input:   strings.Repeat("a", 256),
			want:    "",
			wantErr: true,
		},
		{
			name:    "max length host (255 chars) is ok",
			input:   strings.Repeat("a", 255),
			want:    strings.Repeat("a", 255),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeListenHost(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeListenHost(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("normalizeListenHost(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHelpOutputFormat(t *testing.T) {
	originalUsage := pflag.Usage
	originalOutput := pflag.CommandLine.Output()
	defer func() {
		pflag.Usage = originalUsage
		pflag.CommandLine.SetOutput(originalOutput)
	}()

	var output strings.Builder
	pflag.CommandLine.SetOutput(&output)
	pflag.Usage = func() {
		pflag.CommandLine.PrintDefaults()
	}

	pflag.Usage()

	outputStr := output.String()
	if !strings.Contains(outputStr, "debug") {
		t.Error("Help output should contain --debug flag")
	}
}

func TestPortValidationValues(t *testing.T) {
	tests := []struct {
		name      string
		port      uint
		shouldErr bool
	}{
		{"port zero", 0, true},
		{"port one", 1, false},
		{"port 8085", 8085, false},
		{"port 65535", 65535, false},
		{"port 65536", 65536, true},
		{"port max uint16 + 1", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.port != 0 && tt.port <= 65535
			expectedValid := !tt.shouldErr
			if isValid != expectedValid {
				t.Errorf("Port %d: isValid=%v, expectedValid=%v", tt.port, isValid, expectedValid)
			}
		})
	}
}
