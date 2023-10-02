package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
)

func parseArgs() (command, globPattern, password string, dryRun *bool, err error) {
	passwordFlag := flag.String("password", "", "Encryption or decryption password")
	versionFlag := flag.Bool("version", false, "Print the version information")
	dryRunFlag := flag.Bool("dry", false, "Simulate a dry run without making actual changes")

	flag.Usage = func() {
		fmt.Println("Usage: lok [command] [glob pattern] [password]")
		fmt.Println("       lok [command] [glob pattern] (with password piped in)")
		fmt.Println("       lok -password [password] [command] (with glob patterns piped in)")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version: %s\nCommit: %s\nBuild Date: %s\n", version, commit, date)
		return
	}

	positionalArgs := flag.Args()

	if len(positionalArgs) == 0 || (positionalArgs[0] != "encrypt" && positionalArgs[0] != "decrypt") {
		return "", "", "", nil, errors.New("invalid command")
	}

	command = positionalArgs[0]

	switch {
	case len(positionalArgs) == 3:
		globPattern = positionalArgs[1]
		password = positionalArgs[2]
	case len(positionalArgs) == 2 && *passwordFlag == "":
		globPattern = positionalArgs[1]
		password = readFromStdin()
	case len(positionalArgs) == 1 && *passwordFlag != "":
		password = *passwordFlag
		globPattern = readFromStdin()
	default:
		return "", "", "", nil, errors.New("invalid arguments")
	}

	return command, globPattern, password, dryRunFlag, nil
}

func readFromStdin() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}
