package db

import (
	"os"
	"servers_updater/internal/domain"
	"strings"
	"testing"
)

func setupTestDB(t *testing.T) (*BoltDB, string) {
	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	path := "test_" + safeName + ".db"
	os.Remove(path)
	database, err := InitDB(path)
	if err != nil {
		t.Fatalf("Error starting database for %s: %v", t.Name(), err)
	}
	return database, path
}

func TestSaveAndGetSMTPConfig(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	config := domain.SMTPConfig{
		Host: "smtp.test.com",
		User: "test@user.com",
	}
	if err := db.SaveSMTPConfig(config); err != nil {
		t.Errorf("It shouldn't return an error when saving the SMTP file.: %v", err)
	}

	saved, err := db.GetSMTPConfig()
	if err != nil || saved.Host != config.Host {
		t.Errorf("Error retrieving SMTP configuration.: %v", err)
	}
}

func TestSSHConfigPersistence(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	t.Run("Save and Retrieve", func(t *testing.T) {
		expectedUser := "root"
		expectedKey := "/home/user/.ssh/id_rsa_test"

		err := db.SaveSSHConfig(expectedUser, expectedKey)
		if err != nil {
			t.Fatalf("Failed to save SSH config: %v", err)
		}

		user, key, err := db.GetSSHConfig()
		if err != nil {
			t.Fatalf("Failed to retrieve SSH config: %v", err)
		}

		if user != expectedUser {
			t.Errorf("Expected user %s, got %s", expectedUser, user)
		}
		if key != expectedKey {
			t.Errorf("Expected key path %s, got %s", expectedKey, key)
		}
	})

	t.Run("Error when not found", func(t *testing.T) {
		dbClean, pathClean := setupTestDB(t)
		defer func() { dbClean.Close(); os.Remove(pathClean) }()

		_, _, err := dbClean.GetSSHConfig()
		if err == nil {
			t.Error("Expected error when SSH config is missing, got nil")
		}
	})
}

func TestPasswordEncryption(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	masterKey := []byte("key-must-have-exactly-32-bytes!!")
	host, pass := "srv-01", "my-secret-pass"

	t.Run("decrypted successfully", func(t *testing.T) {
		db.SavePassword(host, pass, masterKey)
		decrypted, err := db.GetPassword(host, masterKey)
		if err != nil || decrypted != pass {
			t.Errorf("Expected pass %s, received %s (err: %v)", pass, decrypted, err)
		}
	})

	t.Run("Incorrect key error", func(t *testing.T) {
		wrongKey := []byte("wrong-key-must-have-32-bytes-too")
		_, err := db.GetPassword(host, wrongKey)

		if err == nil {
			t.Error("It should have failed with the incorrect master key, but it succeeded.")
		}
	})
}

func TestPoolLimitPersistence(t *testing.T) {
	t.Run("Save and recover limit", func(t *testing.T) {
		db, path := setupTestDB(t)
		defer func() { db.Close(); os.Remove(path) }()

		limit := 42
		db.SavePoolLimit(limit)
		val, _ := db.GetPoolLimit()
		if val != limit {
			t.Errorf("Expected %d, received %d", limit, val)
		}
	})

	t.Run("Default value when empty", func(t *testing.T) {
		db, path := setupTestDB(t)
		defer func() { db.Close(); os.Remove(path) }()

		val, _ := db.GetPoolLimit()
		if val != 10 {
			t.Errorf("The default return value should be 10. But it received %d", val)
		}
	})
}

func TestMachineCRUD(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	userPtr := "admin"
	portPtr := 22
	machine := domain.Machine{
		Host:       "192.168.1.50",
		User:       userPtr,
		Port:       portPtr,
		PrettyName: "Ubuntu 22.04 LTS",
		Status:     "ONLINE",
	}

	if err := db.SaveMachine(machine); err != nil {
		t.Fatalf("Error save machine: %v", err)
	}

	machines, err := db.GetAllMachines()
	if err != nil {
		t.Fatalf("Error reading the machines: %v", err)
	}

	if len(machines) != 1 {
		t.Errorf("waiting for 1 machine, bug return %d:", len(machines))
	}

	if machines[0].Host != machine.Host {
		t.Errorf("Corrupted data. I was expecting host %s, but received %s", machine.Host, machines[0].Host)
	}
}
