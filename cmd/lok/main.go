package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	crypto "github.com/let-value/lok/pkg/crypto"
	utils "github.com/let-value/lok/pkg/utils"
)

const encryptedFileExtension = ".lokd"
const encryptedPattern = "**/*.lokd"

type CryptoFunction func(globPattern string, password string, dryRun bool) error

func processFiles(globPattern, password string, dryRun bool, action CryptoFunction) error {
	return action(globPattern, password, dryRun)
}

func main() {
	command, globPattern, password, dryRun, err := parseArgs()
	if err != nil {
		fmt.Println(err)
		flag.Usage()
		return
	}

	var actionFunc CryptoFunction
	switch command {
	case "encrypt":
		actionFunc = encrypt
	case "decrypt":
		actionFunc = decrypt
	default:
		fmt.Println("Invalid command. Use 'encrypt' or 'decrypt'.")
		return
	}

	if err := processFiles(globPattern, password, *dryRun, actionFunc); err != nil {
		fmt.Println(err)
	}
}

func encrypt(globPattern string, password string, dryRun bool) error {
	var basepath, pattern = doublestar.SplitPattern(filepath.ToSlash(globPattern))

	fsys := os.DirFS(basepath)
	files, err := doublestar.Glob(fsys, pattern)
	if err != nil {
		return errors.New("error reading files: " + err.Error())
	}

	if len(files) == 0 {
		return errors.New("no files found matching the pattern: " + pattern)
	}

	root := utils.BuildTree(files)

	cipher, err := crypto.CreateCipher(password)
	if err != nil {
		return errors.New("error creating cipher: " + err.Error())
	}

	processFunc := func(node *utils.Node) {
		if node.Path == "" || node.Path == "." {
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
			fmt.Printf("error encrypting name %s: %v\n", node.Path, err)
			return
		}

		encryptedName = encryptedName + encryptedFileExtension

		if dryRun {
			fmt.Printf("%s -> %s\n", node.Path, encryptedName)
			return
		}

		newPath := filepath.Join(filepath.Dir(fullPath), encryptedName)

		if isDir {
			if err := os.Rename(fullPath, newPath); err != nil {
				fmt.Printf("error renaming directory %s: %v\n", node.Path, err)
			}

			return
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("error reading file %s: %v\n", node.Path, err)
			return
		}

		encryptedData, err := crypto.EncryptBytes(data, node.Name, cipher)
		if err != nil {
			fmt.Printf("error encrypting file %s: %v\n", node.Path, err)
			return
		}

		if err := os.WriteFile(newPath, encryptedData, 0644); err != nil {
			fmt.Printf("error writing to file %s: %v\n", newPath, err)
			return
		}

		if err := os.Remove(fullPath); err != nil {
			fmt.Printf("error removing original file %s: %v\n", node.Path, err)
		}

	}

	utils.Traverse(root, processFunc)

	return nil
}

func decrypt(globPattern string, password string, dryRun bool) error {
	var basepath, pattern = doublestar.SplitPattern(filepath.ToSlash(globPattern))

	fsys := os.DirFS(basepath)
	files, err := doublestar.Glob(fsys, encryptedPattern)
	if err != nil {
		return errors.New("error reading files: " + err.Error())
	}

	if len(files) == 0 {
		return errors.New("no files found matching the pattern: " + encryptedPattern)
	}

	root := utils.BuildTree(files)

	cipher, err := crypto.CreateCipher(password)
	if err != nil {
		return errors.New("error creating cipher: " + err.Error())
	}

	processFunc := func(node *utils.Node) {
		if node.Path == "" || node.Path == "." {
			return
		}

		fullPath := filepath.Join(basepath, node.Path)
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			fmt.Printf("error accessing %s: %v\n", fullPath, err)
			return
		}

		isDir := fileInfo.IsDir()

		if !strings.HasSuffix(node.Name, encryptedFileExtension) {
			return
		}

		decryptedName, err := crypto.DecryptString(strings.TrimSuffix(node.Name, ".lokd"), cipher)
		if err != nil {
			fmt.Printf("error decrypting name %s: %v\n", node.Path, err)
			return
		}

		is_target, err := doublestar.Match(pattern, decryptedName)
		if err != nil {
			fmt.Printf("error matching pattern %s: %v\n", pattern, err)
			return
		}

		if !is_target {
			return
		}

		if dryRun {
			fmt.Printf("%s -> %s\n", node.Path, decryptedName)
			return
		}

		newPath := filepath.Join(filepath.Dir(fullPath), decryptedName)

		if isDir {
			if err := os.Rename(fullPath, newPath); err != nil {
				fmt.Printf("error renaming directory %s: %v\n", fullPath, err)
			}
			return
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("error reading file %s: %v\n", fullPath, err)
			return
		}

		decryptedData, err := crypto.DecryptBytes(data, decryptedName, cipher)
		if err != nil {
			fmt.Printf("error decrypting file %s: %v\n", fullPath, err)
			return
		}

		if err := os.WriteFile(newPath, decryptedData, 0644); err != nil {
			fmt.Printf("error writing to file %s: %v\n", newPath, err)
			return
		}

		if err := os.Remove(fullPath); err != nil {
			fmt.Printf("error removing original file %s: %v\n", fullPath, err)
		}
	}

	utils.Traverse(root, processFunc)

	return nil
}
