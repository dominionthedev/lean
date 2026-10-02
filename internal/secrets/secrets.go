package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/globaldir"
	"golang.org/x/crypto/pbkdf2"
)

const secretsDir = ".lean/secrets"

type Store struct {
	Backend       string
	Recipient     string
	Identity      string
	MasterKeyEnv  string
	MasterKeyFile string
}

func NewStore(cfg *config.Config) *Store {
	return &Store{
		Backend:       cfg.Secrets.Backend,
		Recipient:     cfg.Secrets.Recipient,
		Identity:      cfg.Secrets.Identity,
		MasterKeyEnv:  cfg.Secrets.MasterKeyEnv,
		MasterKeyFile: cfg.Secrets.MasterKeyFile,
	}
}

func secretPath(profile, key string) string {
	safe := strings.ReplaceAll(key, "/", "_")
	return filepath.Join(secretsDir, profile, safe+".secret")
}

func (s *Store) Put(profile, key, value string) error {
	dir := filepath.Join(secretsDir, profile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	encrypted, err := s.encrypt([]byte(value))
	if err != nil {
		return err
	}
	return os.WriteFile(secretPath(profile, key), encrypted, 0600)
}

func (s *Store) Get(profile, key string) (string, error) {
	data, err := os.ReadFile(secretPath(profile, key))
	if err != nil {
		return "", err
	}
	plain, err := s.decrypt(data)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (s *Store) List(profile string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(secretsDir, profile))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var keys []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".secret") {
			keys = append(keys, strings.TrimSuffix(e.Name(), ".secret"))
		}
	}
	return keys, nil
}

func (s *Store) Delete(profile, key string) error {
	return os.Remove(secretPath(profile, key))
}

func (s *Store) Exists(profile, key string) bool {
	_, err := os.Stat(secretPath(profile, key))
	return err == nil
}

func Hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

// Keygen writes a random master key to ~/.lean/key with mode 0600.
func Keygen() (string, error) {
	path, err := globaldir.KeyPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("key already exists at %s — delete it first to regenerate", path)
	}
	raw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(hex.EncodeToString(raw)+"\n"), 0600); err != nil {
		return "", err
	}
	_ = os.Chmod(path, 0600)
	return path, nil
}

func KeyExists() bool {
	path, err := globaldir.KeyPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func (s *Store) encrypt(plain []byte) ([]byte, error) {
	switch s.Backend {
	case "gpg":
		return encryptGPG(plain, s.Recipient)
	case "age", "ssh":
		return encryptAge(plain, s.Recipient)
	default:
		return s.encryptLocal(plain)
	}
}

func (s *Store) decrypt(data []byte) ([]byte, error) {
	switch s.Backend {
	case "gpg":
		return decryptGPG(data)
	case "age", "ssh":
		return decryptAge(data, s.Identity)
	default:
		return s.decryptLocal(data)
	}
}

// resolveMasterKeyMaterial priority:
//  1. env (LEAN_MASTER_KEY or secrets.master_key_env)
//  2. secrets.master_key_file
//  3. ~/.lean/key
func (s *Store) resolveMasterKeyMaterial() ([]byte, error) {
	envName := s.MasterKeyEnv
	if envName == "" {
		envName = "LEAN_MASTER_KEY"
	}
	if v := os.Getenv(envName); v != "" {
		return []byte(v), nil
	}
	if s.MasterKeyFile != "" {
		data, err := os.ReadFile(expandHome(s.MasterKeyFile))
		if err != nil {
			return nil, fmt.Errorf("master_key_file: %w", err)
		}
		return []byte(strings.TrimSpace(string(data))), nil
	}
	path, err := globaldir.KeyPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no master key — run `lean secret keygen` (~/.lean/key), set %s, or secrets.master_key_file", envName)
		}
		return nil, err
	}
	return []byte(strings.TrimSpace(string(data))), nil
}

func (s *Store) derivedKey() ([]byte, error) {
	raw, err := s.resolveMasterKeyMaterial()
	if err != nil {
		return nil, err
	}
	return pbkdf2.Key(raw, []byte("lean-secrets-v1"), 100000, 32, sha256.New), nil
}

func (s *Store) encryptLocal(plain []byte) ([]byte, error) {
	key, err := s.derivedKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return append([]byte("local:"), gcm.Seal(nonce, nonce, plain, nil)...), nil
}

func (s *Store) decryptLocal(data []byte) ([]byte, error) {
	if !strings.HasPrefix(string(data), "local:") {
		return nil, fmt.Errorf("not a local-backend secret blob")
	}
	data = data[len("local:"):]
	key, err := s.derivedKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}

func encryptGPG(plain []byte, recipient string) ([]byte, error) {
	if recipient == "" {
		return nil, fmt.Errorf("secrets.recipient required for gpg backend")
	}
	cmd := exec.Command("gpg", "--batch", "--yes", "-e", "-r", recipient, "--output", "-")
	cmd.Stdin = strings.NewReader(string(plain))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gpg encrypt: %w", err)
	}
	return append([]byte("gpg:"), out...), nil
}

func decryptGPG(data []byte) ([]byte, error) {
	if !strings.HasPrefix(string(data), "gpg:") {
		return nil, fmt.Errorf("not a gpg-backend secret blob")
	}
	cmd := exec.Command("gpg", "--batch", "--yes", "-d", "--output", "-")
	cmd.Stdin = strings.NewReader(string(data[len("gpg:"):]))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gpg decrypt: %w", err)
	}
	return out, nil
}

func encryptAge(plain []byte, recipient string) ([]byte, error) {
	if recipient == "" {
		return nil, fmt.Errorf("secrets.recipient required (age recipient or path to SSH .pub)")
	}
	r := recipient
	expanded := expandHome(r)
	if data, err := os.ReadFile(expanded); err == nil {
		r = strings.TrimSpace(string(data))
	}
	cmd := exec.Command("age", "-r", r, "-o", "-")
	cmd.Stdin = strings.NewReader(string(plain))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("age encrypt: %w (is age installed?)", err)
	}
	return append([]byte("age:"), out...), nil
}

func decryptAge(data []byte, identity string) ([]byte, error) {
	if !strings.HasPrefix(string(data), "age:") {
		return nil, fmt.Errorf("not an age-backend secret blob")
	}
	data = data[len("age:"):]
	args := []string{"-d"}
	if identity == "" {
		home, _ := os.UserHomeDir()
		for _, cand := range []string{
			filepath.Join(home, ".ssh", "id_ed25519"),
			filepath.Join(home, ".ssh", "id_rsa"),
			filepath.Join(home, ".ssh", "id_ecdsa"),
		} {
			if _, err := os.Stat(cand); err == nil {
				identity = cand
				break
			}
		}
	}
	if identity != "" {
		args = append(args, "-i", expandHome(identity))
	}
	args = append(args, "-o", "-")
	cmd := exec.Command("age", args...)
	cmd.Stdin = strings.NewReader(string(data))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("age decrypt: %w", err)
	}
	return out, nil
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}
