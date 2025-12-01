package cryptolib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	ChunkSize  = 64 * 1024 // 64KB input block
	KeySize    = 32
	NonceSize  = 12
	SaltSize   = 16
	Iterations = 100000
)

// deriveKey generates a 32-byte key from a password string using PBKDF2.
func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, Iterations, KeySize, sha256.New)
}

// incrementNonce increments the nonce (treated as a big-endian integer).
func incrementNonce(nonce []byte) {
	for i := len(nonce) - 1; i >= 0; i-- {
		nonce[i]++
		if nonce[i] != 0 {
			break
		}
	}
}

// Callback interface for progress updates
type ProgressCallback func(bytesProcessed int64, totalBytes int64)

// EncryptStream encrypts data from input reader to output writer.
// totalBytes is used for progress calculation; if unknown, pass 0 (or progress bar might look weird).
func EncryptStream(input io.Reader, output io.Writer, password string, totalBytes int64, onProgress ProgressCallback) error {
	// Generate Salt
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	// Write salt to output
	if _, err := output.Write(salt); err != nil {
		return err
	}

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// Write the base nonce to the output file
	if _, err := output.Write(nonce); err != nil {
		return err
	}

	buf := make([]byte, ChunkSize)
	currNonce := make([]byte, NonceSize)
	copy(currNonce, nonce)

	lenBuf := make([]byte, 4)
	var processedBytes int64

	for {
		n, err := input.Read(buf)
		if n > 0 {
			ciphertext := aesGCM.Seal(nil, currNonce, buf[:n], nil)
			
			binary.BigEndian.PutUint32(lenBuf, uint32(len(ciphertext)))
			if _, err := output.Write(lenBuf); err != nil {
				return err
			}

			if _, err := output.Write(ciphertext); err != nil {
				return err
			}
			
			incrementNonce(currNonce)
			processedBytes += int64(n)
			if onProgress != nil {
				onProgress(processedBytes, totalBytes)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// DecryptStream decrypts data from input reader to output writer.
// totalBytes is optional (can be 0) for progress.
func DecryptStream(input io.Reader, output io.Writer, password string, totalBytes int64, onProgress ProgressCallback) error {
	// Read Salt
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(input, salt); err != nil {
		return err
	}

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Read Base Nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(input, nonce); err != nil {
		return err
	}
	
	currNonce := make([]byte, NonceSize)
	copy(currNonce, nonce)

	lenBuf := make([]byte, 4)
	var processedBytes int64 = int64(SaltSize + NonceSize)
	
	overhead := aesGCM.Overhead()
	maxChunkLen := uint32(ChunkSize + overhead)

	for {
		if _, err := io.ReadFull(input, lenBuf); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		processedBytes += 4
		chunkLen := binary.BigEndian.Uint32(lenBuf)

		if chunkLen > maxChunkLen {
			return fmt.Errorf("chunk too large: %d > %d (corrupted file or attack)", chunkLen, maxChunkLen)
		}

		ciphertext := make([]byte, chunkLen)
		if _, err := io.ReadFull(input, ciphertext); err != nil {
			return fmt.Errorf("unexpected EOF while reading chunk")
		}
		processedBytes += int64(chunkLen)

		plaintext, err := aesGCM.Open(nil, currNonce, ciphertext, nil)
		if err != nil {
			return err
		}
		if _, err := output.Write(plaintext); err != nil {
			return err
		}
		incrementNonce(currNonce)
		if onProgress != nil {
			onProgress(processedBytes, totalBytes)
		}
	}
	return nil
}

func EncryptFile(inputPath, outputPath, password string, onProgress ProgressCallback) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	fileInfo, err := inFile.Stat()
	if err != nil {
		return err
	}
	totalSize := fileInfo.Size()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return EncryptStream(inFile, outFile, password, totalSize, onProgress)
}

func DecryptFile(inputPath, outputPath, password string, onProgress ProgressCallback) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	fileInfo, err := inFile.Stat()
	if err != nil {
		return err
	}
	totalSize := fileInfo.Size() 

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return DecryptStream(inFile, outFile, password, totalSize, onProgress)
}
