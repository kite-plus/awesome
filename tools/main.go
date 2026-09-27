// Command tools keeps the list in the folder above it.
//
// check reads the entries, index.json and the READMEs without touching the
// network. sync reads each entry's latest release, checks it with Kite, and
// writes index.json and the READMEs from what it read. readme writes the
// READMEs again from index.json.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const usage = `usage: go -C tools run . <command>

  check                  check the entries, index.json and the READMEs, offline
  sync [flags] [entry]   read the latest releases, check them with Kite, and
                         write index.json and the READMEs
  readme                 write the READMEs from index.json

An entry is named by its file, as plugins/search.yaml. sync -h lists its flags.`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	root, err := findRoot()
	if err == nil {
		switch os.Args[1] {
		case "check":
			err = runCheck(root, os.Stdout)
		case "sync":
			err = runSync(root, os.Args[2:])
		case "readme":
			err = runReadme(root)
		default:
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// findRoot finds the list from anywhere inside it, since go -C runs the
// command in tools/.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, categoriesFile)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("not inside the list: there is no " + categoriesFile + " here or above")
		}
		dir = parent
	}
}
