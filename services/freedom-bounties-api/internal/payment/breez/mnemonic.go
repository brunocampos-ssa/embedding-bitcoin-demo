package breez

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

// treasuryMnemonicFile is where a generated treasury mnemonic is persisted,
// relative to the Breez storage directory.
const treasuryMnemonicFile = "treasury.mnemonic"

// mnemonicSource records where the treasury mnemonic came from, for logging.
type mnemonicSource string

const (
	mnemonicProvided  mnemonicSource = "provided"
	mnemonicLoaded    mnemonicSource = "loaded"
	mnemonicGenerated mnemonicSource = "generated"
)

// resolveMnemonic determines the treasury mnemonic to hand to the Breez SDK.
// Precedence:
//  1. an explicitly provided mnemonic (BREEZ_MNEMONIC), trimmed and validated;
//  2. otherwise a previously generated mnemonic persisted under storageDir is reused;
//  3. otherwise a fresh BIP39 mnemonic is generated, persisted with 0600, and returned.
//
// It returns the mnemonic, its source, and the persistence path. Callers log the
// source and path only — never the mnemonic itself.
func resolveMnemonic(storageDir, provided string) (mnemonic string, source mnemonicSource, path string, err error) {
	if m := strings.TrimSpace(provided); m != "" {
		if !bip39.IsMnemonicValid(m) {
			return "", "", "", fmt.Errorf("BREEZ_MNEMONIC is not a valid BIP39 mnemonic")
		}
		return m, mnemonicProvided, "", nil
	}
	dir := strings.TrimSpace(storageDir)
	if dir == "" {
		return "", "", "", fmt.Errorf("cannot persist a generated treasury mnemonic: storage directory is empty")
	}
	path = filepath.Join(dir, treasuryMnemonicFile)
	if m, ok, rerr := readValidMnemonic(path); rerr != nil {
		return "", "", path, rerr
	} else if ok {
		return m, mnemonicLoaded, path, nil
	}
	entropy, err := bip39.NewEntropy(128) // 128 bits -> 12 words
	if err != nil {
		return "", "", path, fmt.Errorf("generate entropy: %w", err)
	}
	m, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", "", path, fmt.Errorf("generate mnemonic: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", path, fmt.Errorf("create storage dir: %w", err)
	}
	// Exclusive create: whoever wins O_EXCL owns the wallet. A concurrent
	// initializer that loses the race re-reads the winner's file instead of
	// clobbering it (which would strand funds in the first wallet).
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			if existing, ok, rerr := readValidMnemonic(path); rerr != nil {
				return "", "", path, rerr
			} else if ok {
				return existing, mnemonicLoaded, path, nil
			}
			return "", "", path, fmt.Errorf("treasury mnemonic at %s is empty or invalid", path)
		}
		return "", "", path, fmt.Errorf("create treasury mnemonic file: %w", err)
	}
	if _, werr := f.WriteString(m + "\n"); werr != nil {
		f.Close()
		return "", "", path, fmt.Errorf("persist treasury mnemonic: %w", werr)
	}
	if cerr := f.Close(); cerr != nil {
		return "", "", path, fmt.Errorf("persist treasury mnemonic: %w", cerr)
	}
	return m, mnemonicGenerated, path, nil
}

// readValidMnemonic reads a persisted mnemonic. ok=false with a nil error means
// the file does not exist; a present-but-invalid file returns an error.
func readValidMnemonic(path string) (mnemonic string, ok bool, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("read treasury mnemonic: %w", err)
	}
	m := strings.TrimSpace(string(b))
	if !bip39.IsMnemonicValid(m) {
		return "", false, fmt.Errorf("stored treasury mnemonic at %s is invalid", path)
	}
	return m, true, nil
}
