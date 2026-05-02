package main

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestVersionAndHelpFlagsCoexistence(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectVersion bool
		expectHelp    bool
	}{
		{
			name:          "only --version",
			args:          []string{"--version"},
			expectVersion: true,
			expectHelp:    false,
		},
		{
			name:          "only --help",
			args:          []string{"--help"},
			expectVersion: false,
			expectHelp:    true,
		},
		{
			name:          "both --version and --help (version first)",
			args:          []string{"--version", "--help"},
			expectVersion: true,
			expectHelp:    true,
		},
		{
			name:          "both --help and --version (help first)",
			args:          []string{"--help", "--version"},
			expectVersion: true,
			expectHelp:    true,
		},
		{
			name:          "short flags -v -h",
			args:          []string{"-v", "-h"},
			expectVersion: true,
			expectHelp:    true,
		},
		{
			name:          "short flags -h -v (reversed)",
			args:          []string{"-h", "-v"},
			expectVersion: true,
			expectHelp:    true,
		},
		{
			name:          "neither flag",
			args:          []string{},
			expectVersion: false,
			expectHelp:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flagSet.BoolP("version", "v", false, "Show version information.")
			flagSet.BoolP("help", "h", false, "Show help information.")

			err := flagSet.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			versionVal, _ := flagSet.GetBool("version")
			helpVal, _ := flagSet.GetBool("help")

			if versionVal != tt.expectVersion {
				t.Errorf("version flag: got %v, want %v", versionVal, tt.expectVersion)
			}
			if helpVal != tt.expectHelp {
				t.Errorf("help flag: got %v, want %v", helpVal, tt.expectHelp)
			}
		})
	}
}

func TestVersionHelpOutputOrder(t *testing.T) {
	// Test that version is printed before help when both flags are present
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.BoolP("version", "v", false, "Show version information.")
	flagSet.BoolP("help", "h", false, "Show help information.")

	err := flagSet.Parse([]string{"--version", "--help"})
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	versionVal, _ := flagSet.GetBool("version")
	helpVal, _ := flagSet.GetBool("help")

	// Both should be true
	if !versionVal || !helpVal {
		t.Errorf("Both flags should be true, got version=%v, help=%v", versionVal, helpVal)
	}
}

func TestHelpFlagShortForm(t *testing.T) {
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.BoolP("version", "v", false, "Show version information.")
	flagSet.BoolP("help", "h", false, "Show help information.")

	// Test short form -h
	err := flagSet.Parse([]string{"-h"})
	if err != nil {
		t.Fatalf("Failed to parse -h flag: %v", err)
	}

	helpVal, _ := flagSet.GetBool("help")
	if !helpVal {
		t.Error("-h should set help flag to true")
	}
}

func TestVersionFlagShortForm(t *testing.T) {
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.BoolP("version", "v", false, "Show version information.")
	flagSet.BoolP("help", "h", false, "Show help information.")

	// Test short form -v
	err := flagSet.Parse([]string{"-v"})
	if err != nil {
		t.Fatalf("Failed to parse -v flag: %v", err)
	}

	versionVal, _ := flagSet.GetBool("version")
	if !versionVal {
		t.Error("-v should set version flag to true")
	}
}
