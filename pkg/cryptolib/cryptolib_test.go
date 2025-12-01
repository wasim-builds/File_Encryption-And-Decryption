package cryptolib

import (
	"bytes"
	"os"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// Setup
	inputContent := []byte("This is a test message for encryption and decryption.")
	inputFile := "test_input.txt"
	encryptedFile := "test_input.txt.enc"
	decryptedFile := "test_input.dec.txt"
	password := "strongpassword123"

	// Create input file
	err := os.WriteFile(inputFile, inputContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	defer os.Remove(inputFile)
	defer os.Remove(encryptedFile)
	defer os.Remove(decryptedFile)

	// Encrypt
	err = EncryptFile(inputFile, encryptedFile, password, nil)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Verify encrypted file exists and is different from input
	encryptedContent, err := os.ReadFile(encryptedFile)
	if err != nil {
		t.Fatalf("Failed to read encrypted file: %v", err)
	}
	if bytes.Equal(inputContent, encryptedContent) {
		t.Fatal("Encrypted content is identical to input content (encryption failed to obfuscate)")
	}

	// Decrypt
	err = DecryptFile(encryptedFile, decryptedFile, password, nil)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify decrypted content matches original
	decryptedContent, err := os.ReadFile(decryptedFile)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}
	if !bytes.Equal(inputContent, decryptedContent) {
		t.Fatalf("Decrypted content does not match original.\nExpected: %s\nGot: %s", inputContent, decryptedContent)
	}
}

func TestDecryptWithWrongPassword(t *testing.T) {
	// Setup
	inputContent := []byte("Top secret data")
	inputFile := "wrong_pass_input.txt"
	encryptedFile := "wrong_pass_input.txt.enc"
	decryptedFile := "wrong_pass_input.dec.txt"
	password := "correctpassword"
	wrongPassword := "wrongpassword"

	err := os.WriteFile(inputFile, inputContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	defer os.Remove(inputFile)
	defer os.Remove(encryptedFile)
	defer os.Remove(decryptedFile) // Should not be created ideally, but good cleanup

	// Encrypt
	err = EncryptFile(inputFile, encryptedFile, password, nil)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Attempt Decrypt with wrong password
	err = DecryptFile(encryptedFile, decryptedFile, wrongPassword, nil)
	if err == nil {
		t.Fatal("Decryption should have failed with wrong password, but it succeeded")
	}
}
