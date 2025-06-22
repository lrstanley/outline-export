// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"fmt"

	"github.com/lrstanley/clix"
)

var (
	version = "master"
	commit  = "latest"
	date    = "-"

	cli = &clix.CLI[Flags]{
		Links: clix.GithubLinks("github.com/lrstanley/YOUR_PROJECT_NAME", "master", "https://liam.sh"),
		VersionInfo: &clix.VersionInfo[Flags]{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
)

type Flags struct {
	ExampleFlag bool `long:"example-flag" env:"EXAMPLE_FLAG" description:"example flag"`
}

func main() {
	cli.Parse(clix.OptDisableLogging | clix.OptDisableGlobalLogger)

	fmt.Println("Hello, world!")
}
