package db

import (
 "os"
 "servers_updater/internal/domain"
 "testing"
)

func setupTestDB(t *testing.T) (*BoltDB, string) {
 path := "test_" + t.Name() + ".db"
 database, err := InitDB(path)
  if err != nil {
   t.Fatalf("Error starting database for %s: %v", t.Name(), err)
  }
  return database, path
}

func TestSaveAndGetSMTPConfig(t *testing.T) {
 db, path := setupTestDB(t)
 defer func() { db.Close(); os.Remove(path) }()

 config := domain.SMTPConfig{Server: "smtp.test.com", User: "test@user.com"}
 if err := db.SaveSMTPConfig(config); err != nil {
  t.Errorf("It shouldn't return an error when saving the SMTP file.: %v", err)
 }

 saved, err := db.GetSMTPConfig()
 if err != nil || saved.Server != config.Server {
  t.Errorf("Error retrieving SMTP configuration.: %v", err)
 }
}

func TestPasswordEncryption(t *testing.T) {
 db, path := setupTestDB(t)
 defer func() {db.Close(); os.Remove(path) }()

 masterKey := []byte("key-must-have-exactly-32-bytes!!")
 host, pass := "srv-01" , "my-secret-pass"

 t.Run("decrypted successfully", func(t *testing.T) {
  db.SavePassword(host, pass, masterKey)
  decrypted, err := db.GetPassword(host, masterKey)
  if err != nil || decrypted != pass {
   t.Errorf("Expected error %s, received %s", pass, decrypted)
  }
 })

 t.Run("Incorrect key error", func(t *testing.T) {
  wrongKey := []byte("wrong-key-must-have-32-bytes-too")
  _, err := db.GetPassword(host, wrongKey)
  if err != nil {
   t.Error("It should fail with the incorrect master key.")
  }
 })
}

func TestPoolLimitPersistence(t *testing.T) {
 db, path := setupTestDB(t)
 defer func() { db.Close(); os.Remove(path) }()

 t.Run("Save and recover limit", func(t *testing.T) {
  limit := 42
  db.SavePoolLimit(limit)
  val, _ := db.GetPoolLimit()
  if val != limit {
   t.Errorf("Expected %d, received %d", limit , val)
  }
 })

 t.Run("Default value when empty", func(t *testing.T) {
  val, _ := db.GetPoolLimit()
  if val != 10 {
   t.Errorf("The default return value should be 10. But it received %d", val)
  }
 })
}
