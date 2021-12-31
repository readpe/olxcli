// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/readpe/olxcli/cmd/busfaults"
	"github.com/readpe/olxcli/cmd/noclear"
	"github.com/spf13/cobra"
)

var rootCmd = cobra.Command{
	Use:   "olxcli",
	Short: "olxcli is an unofficial command line interface for ASPEN's Oneliner.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

// Command package global variables.
var (
	version string
	license string
)

func init() {
	// Add to subcommands to root command.
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(licenseCmd)
	rootCmd.AddCommand(busfaults.BFCmd)
	rootCmd.AddCommand(noclear.NCCmd)
}

// Execute the root command.
func Execute(v, l string) error {
	version = v
	license = l
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	return rootCmd.Execute()
}
