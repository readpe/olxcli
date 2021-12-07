// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/readpe/goolx"
	"github.com/spf13/cobra"
)

// versionCmd represents the version information command.
var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Display the current command line application version",
	RunE:    runVersion,
}

func init() {
	versionCmd.Flags().BoolVar(&apiVerFlag, "api", false, "display the OlxAPI version")
}

var (
	apiVerFlag bool
)

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Fprintf(os.Stdout, "OlxCLI Version: %s\n", version)
	if apiVerFlag {
		api := goolx.NewClient()
		defer api.Release()
		fmt.Fprintln(os.Stdout, "OlxAPI Version:")
		sc := bufio.NewScanner(strings.NewReader(api.Info()))
		for sc.Scan() {
			fmt.Fprintf(os.Stdout, "\t%s\n", sc.Text())
		}
	}
	return nil
}
