// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/readpe/olxcli/cmd"

	_ "embed"
)

// version will be overwritten by build flag.
var version = "v0.0.1"

//go:embed LICENSE
var license string

func main() {
	err := cmd.Execute(version, license)
	if err != nil {
		fmt.Println(err)
	}
}
