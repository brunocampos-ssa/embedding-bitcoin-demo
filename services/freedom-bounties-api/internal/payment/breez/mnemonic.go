package breez

import (
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
//  1. an explicitly provided mnemonic (BREEZ_MNEMONIC) is used verbatim;
//  2. otherwise a previously generated mnemonic persisted under storageDir is reused;
//  3. otherwise a fresh BIP39 mnemonic is generated, persisted with 0600, and returned.
//
// It returns the mnemonic, its source, and the persistence path. Callers log the
// source and path only — never the mnemonic itself.
func resolveMnemonic(storageDir, provided string) (mnemonic string, source mnemonicSource, path string, err error) {
	if m := strings.TrimSpace(provided); m != "" {
		return m, mnemonicProvided, "", nil
	}
	path = filepath.Join(storageDir, treasuryMnemonicFile)
	if b, readErr := os.ReadFile(path); readErr == nil {
		m := strings.TrimSpace(string(b))
		if !bip39.IsMnemonicValid(m) {
			return "", "", path, fmt.Errorf("stored treasury mnemonic at %s is invalid", path)
		}
		return m, mnemonicLoaded, path, nil
	} else if !os.IsNotExist(readErr) {
		return "", "", path, fmt.Errorf("read treasury mnemonic: %w", readErr)
	}
	entropy, err := bip39.NewEntropy(128) // 128 bits -> 12 words
	if err != nil {
		return "", "", path, fmt.Errorf("generate entropy: %w", err)
	}
	m, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", "", path, fmt.Errorf("generate mnemonic: %w", err)
	}
	if err := os.MkdirAll(storageDir, 0o700); err != nil {
		return "", "", path, fmt.Errorf("create storage dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(m+"\n"), 0o600); err != nil {
		return "", "", path, fmt.Errorf("persist treasury mnemonic: %w", err)
	}
	return m, mnemonicGenerated, path, nil
}
