package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gocrypt/pkg/cryptolib"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "encrypt":
		runEncrypt(os.Args[2:])
	case "decrypt":
		runDecrypt(os.Args[2:])
	case "serve":
		runServeCommand(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gocrypt encrypt -f <file> -p <password>")
	fmt.Println("  gocrypt decrypt -f <file> -p <password>")
	fmt.Println("  gocrypt serve [-port <port>]")
}

func printProgress(current, total int64) {
	percentage := float64(current) / float64(total) * 100
	fmt.Printf("\rProgress: %.2f%%", percentage)
	if current >= total {
		fmt.Println()
	}
}

func runEncrypt(args []string) {
	fs := flag.NewFlagSet("encrypt", flag.ExitOnError)
	filePtr := fs.String("f", "", "File to encrypt")
	passPtr := fs.String("p", "", "Password")

	fs.Parse(args)

	if *filePtr == "" || *passPtr == "" {
		fmt.Println("Error: Both -f (file) and -p (password) are required.")
		fs.PrintDefaults()
		os.Exit(1)
	}

	inputFile := *filePtr
	outputFile := inputFile + ".enc"

	fmt.Printf("Encrypting %s...\n", inputFile)
	err := cryptolib.EncryptFile(inputFile, outputFile, *passPtr, printProgress)
	if err != nil {
		fmt.Printf("\nError encrypting file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Success! Encrypted file saved as %s\n", outputFile)
}

func runDecrypt(args []string) {
	fs := flag.NewFlagSet("decrypt", flag.ExitOnError)
	filePtr := fs.String("f", "", "File to decrypt")
	passPtr := fs.String("p", "", "Password")

	fs.Parse(args)

	if *filePtr == "" || *passPtr == "" {
		fmt.Println("Error: Both -f (file) and -p (password) are required.")
		fs.PrintDefaults()
		os.Exit(1)
	}

	inputFile := *filePtr
	outputFile := strings.TrimSuffix(inputFile, ".enc")
    if outputFile == inputFile {
        outputFile = inputFile + ".dec"
    }

	fmt.Printf("Decrypting %s...\n", inputFile)
	err := cryptolib.DecryptFile(inputFile, outputFile, *passPtr, printProgress)
	if err != nil {
		fmt.Printf("\nError decrypting file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Success! Decrypted file saved as %s\n", outputFile)
}

func runServeCommand(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	portPtr := fs.String("port", "8080", "Port to listen on")
	fs.Parse(args)
	runServe(*portPtr)
}
