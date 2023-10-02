package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	crypto "github.com/let-value/lok/pkg/crypto"
	utils "github.com/let-value/lok/pkg/shared"
)

func encrypt(globPattern string, password string, dryRun bool) {
	// Find files matching the glob pattern
	var basepath, pattern = doublestar.SplitPattern(filepath.ToSlash(globPattern))

	fsys := os.DirFS(basepath)
	files, err := doublestar.Glob(fsys, pattern)
	if err != nil {
		fmt.Println("Error reading files:", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No files found matching the pattern:", globPattern)
		return
	}

	// Build a tree of files
	root := utils.BuildTree(files, false)

	cipher, err := crypto.CreateCipher(password)
	if err != nil {
		fmt.Println("Error creating cipher:", err)
		return
	}

	//Traverse tree and encrypt files
	processFunc := func(node *utils.Node) {
		if node.Path == "" || node.Path == "." {
			// Skip root node
			return
		}

		isDir := false

		fullPath := filepath.Join(basepath, node.Path)
		fileInfo, err := os.Stat(fullPath)
		if err == nil {
			isDir = fileInfo.IsDir()
		}

		encryptedName, err := crypto.EncryptString(node.Name, cipher)
		if err != nil {
			fmt.Printf("Error encrypting name %s: %v\n", node.Path, err)
			return
		}

		encryptedName = encryptedName + ".lokd"

		if dryRun {
			// For dry run, just display the name before and after encryption

			fmt.Printf("%s -> %s\n", node.Path, encryptedName)
			return
		}

		newPath := filepath.Join(filepath.Dir(node.Path), encryptedName)

		if isDir {
			// Encrypt directory name and rename the directory
			if err := os.Rename(node.Path, newPath); err != nil {
				fmt.Printf("Error renaming directory %s: %v\n", node.Path, err)
			}

			return
		}

		// Encrypt file name and contents
		data, err := os.ReadFile(node.Path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", node.Path, err)
			return
		}

		// Encrypt the data
		nonce := crypto.DeriveNonce(data, cipher)
		encryptedData := cipher.Seal(nil, nonce, data, nil)

		// Write encrypted data back to the file
		if err := os.WriteFile(newPath, encryptedData, 0644); err != nil {
			fmt.Printf("Error writing to file %s: %v\n", newPath, err)
			return
		}

		// Remove the original file
		if err := os.Remove(node.Path); err != nil {
			fmt.Printf("Error removing original file %s: %v\n", node.Path, err)
		}

	}

	utils.Traverse(root, processFunc)

	return
}

func decrypt(globPattern string, password string, dryRun bool) {
	return
}

func main() {
	// Define command-line flags
	passwordFlag := flag.String("password", "", "Encryption or decryption password")
	dryRun := flag.Bool("dry", false, "Simulate a dry run without making actual changes")
	flag.Usage = func() {
		fmt.Println("Usage: lok [command] [glob pattern] [password]")
		fmt.Println("       lok [command] [glob pattern] (with password piped in)")
		fmt.Println("       lok -password [password] [command] (with glob patterns piped in)")
		flag.PrintDefaults()
	}

	// Parse the flags
	flag.Parse()

	// Filter out flags to get positional arguments
	var positionalArgs []string
	for _, arg := range flag.Args() {
		if !strings.HasPrefix(arg, "-") {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	command := positionalArgs[0]
	if command != "encrypt" && command != "decrypt" {
		fmt.Println("Invalid command. Use 'encrypt' or 'decrypt'.")
		return
	}

	var globPattern, password string

	// Handle positional arguments and piped input
	switch {
	case len(positionalArgs) == 3:
		globPattern = positionalArgs[1]
		password = positionalArgs[2]
	case len(positionalArgs) == 2 && *passwordFlag == "":
		globPattern = positionalArgs[1]
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			password = scanner.Text()
		}
	case len(positionalArgs) == 1 && *passwordFlag != "":
		password = *passwordFlag
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			globPattern = scanner.Text()
		}
	default:
		fmt.Println("Invalid arguments.")
		flag.Usage()
		return
	}

	switch command {
	case "encrypt":
		encrypt(globPattern, password, *dryRun)
	case "decrypt":
		decrypt(globPattern, password, *dryRun)
	}
}
