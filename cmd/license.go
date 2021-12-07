// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// licenseCmd represents the license command.
var licenseCmd = &cobra.Command{
	Use:     "license",
	Aliases: []string{"l"},
	Short:   "Display the license information",
	RunE:    runLicense,
}

func runLicense(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(os.Stdout, license)
	return nil
}
