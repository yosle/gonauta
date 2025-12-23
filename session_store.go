package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func getSessionPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".gonauta")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "session.json"), nil
}

func SaveSession(session *SessionData) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	sessionPath, err := getSessionPath()
	if err != nil {
		return err
	}

	return os.WriteFile(sessionPath, data, 0600)
}

func LoadSession() (*SessionData, error) {
	sessionPath, err := getSessionPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("no hay sesión activa")
		}
		return nil, err
	}

	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, err
	}

	return &sessionData, nil
}

func DeleteSession() error {
	sessionPath, err := getSessionPath()
	if err != nil {
		return err
	}

	err = os.Remove(sessionPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
