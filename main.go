// Lean – A smart tool for Managing your env files

package main

import "github.com/dominionthedev/lean/cmd/lean"

// use version from /cmd/lean/version.go
var version string

func main() {
	lean.Execute(version)
}
