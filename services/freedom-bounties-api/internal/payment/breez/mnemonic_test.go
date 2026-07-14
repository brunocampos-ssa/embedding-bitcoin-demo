package breez

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tyler-smith/go-bip39"
)

func TestResolveMnemonicUsesProvided(t *testing.T) {
	provided := "legal winner thank year wave sausage worth useful legal winner thank yellow"
	m, src, _, err := resolveMnemonic(t.TempDir(), provided)
	if err != nil || m != provided || src != mnemonicProvided {
		t.Fatalf("provided: m=%q src=%q err=%v", m, src, err)
	}
}

func TestResolveMnemonicGeneratesPersistsAndReuses(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "breez") // does not exist yet
	m1, src, path, err := resolveMnemonic(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if src != mnemonicGenerated {
		t.Fatalf("expected generated, got %q", src)
	}
	if !bip39.IsMnemonicValid(m1) {
		t.Fatalf("generated mnemonic is not valid: %q", m1)
	}
	// Persisted with owner-only permissions.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mnemonic file perm = %o, want 600", perm)
	}
	// A second call reuses the persisted mnemonic rather than generating a new one.
	m2, src2, _, err := resolveMnemonic(dir, "")
	if err != nil || m2 != m1 || src2 != mnemonicLoaded {
		t.Fatalf("reuse: m2=%q src2=%q err=%v (want loaded, same words)", m2, src2, err)
	}
}

func TestResolveMnemonicRejectsInvalidProvided(t *testing.T) {
	if _, _, _, err := resolveMnemonic(t.TempDir(), "these are not twelve valid bip39 words at all"); err == nil {
		t.Fatal("expected error for an invalid provided mnemonic")
	}
}

func TestResolveMnemonicRejectsEmptyStorageDir(t *testing.T) {
	// No mnemonic provided and nowhere to persist a generated one.
	if _, _, _, err := resolveMnemonic("   ", ""); err == nil {
		t.Fatal("expected error when storage dir is empty and generation is needed")
	}
}

func TestResolveMnemonicRejectsCorruptFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, treasuryMnemonicFile), []byte("not a valid mnemonic"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := resolveMnemonic(dir, ""); err == nil {
		t.Fatal("expected error for corrupt stored mnemonic")
	}
}
