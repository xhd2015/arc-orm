package main

import (
	"fmt"
	"os"
	"strings"
)

const help = `
Usage: ormx <command>

Commands:
  gen     generate models

`

func main() {
	err := handle(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires command, run `ormx help`")
	}

	switch args[0] {
	case "help":
		fmt.Println(strings.TrimPrefix(help, "\n"))
		return nil
	case "gen":
		return gen(args[1:])
	}

	return fmt.Errorf("unknown command, run `ormx help`")
}

func gen(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires table name, run `ormx gen <table_name>`")
	}
	// TODO:

	return nil
}
