package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	VPNConnectCmd    string `json:"vpn_connect_cmd,omitempty"`
	VPNDisconnectCmd string `json:"vpn_disconnect_cmd,omitempty"`
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".gonauta")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "credentials.enc"), nil
}

func getEncryptionKey() []byte {
	hostname, _ := os.Hostname()
	key := sha256.Sum256([]byte("gonauta-" + hostname))
	return key[:]
}

func encrypt(data []byte, key []byte) ([]byte, error) {
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

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("datos cifrados inválidos")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func SaveCredentials(username, password, vpnConnectCmd, vpnDisconnectCmd string) error {
	config := Config{
		Username:         username,
		Password:         password,
		VPNConnectCmd:    vpnConnectCmd,
		VPNDisconnectCmd: vpnDisconnectCmd,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	encrypted, err := encrypt(data, getEncryptionKey())
	if err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, encrypted, 0600)
}

func LoadCredentials() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	encrypted, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("no hay credenciales guardadas. Use 'gonauta login' primero")
		}
		return nil, err
	}

	decrypted, err := decrypt(encrypted, getEncryptionKey())
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(decrypted, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func DeleteCredentials() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	err = os.Remove(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
