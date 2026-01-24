package db

import (
 "os"
 "testing"
 "servers_updater/internal/db"
 "servers_updater/internal/domain"
)

func setupTestDB(t *testing.T) (*db.BoltDB, func() {
 tmpFile, err := os.CreateTemp("", "test_email_*db")
 if err != nil {
  t.Fatalf("Could not create temp db file: %v", err)
 }
 dbPath := tmpFile.Name()
 tmpFile.Close()

 database, err := db.NewBoltDB(dbPath)
 if err != nil {
  t.Fatalf("Could not open bolt db: %v", err)
 }

 return database, func() {
  database.Close()
  os.Remove(dbPath)
 }
}

func testSMTPConfig_CRUD(t *testing.T) {
 database, teardown := setupTestDB(t)
 defer teardown()

 _, err := database.getSTMPConfig()
 if err == nil {
  t.Errorf("Expected error when getting non-existent config, got nil")
 } 
 expected := domain.SMTPConfig{
  Host: "smtp.gmail.com",
  Port: "587",
  User: "samuelbretas@gmail.com",
  Password: "A12j6n542",
 }

 err = database.SaveSMTPConfig(expected)
 if err != nil {
  t.Fatalf("Failed to save config: %v", err)
 }

 saved, err := database.GetSMTPConfig()
 if err != nil {
  t.Fatalf("Failed to get config: %v", err)
 }

 if saved.Host != expected.Host || saved.User != expected.User {
  t.Errorf("Saved config mismatch. Want %v, Got %v", expected, saved)
 }
}

func TestRecipients_CRUD(t *testing.T) {
 database, teardown := setupTestDB(t)
 defer teardown()

 var (
  email1 = "dev1@emailCompany.com"
  email2 = "dev2@emailCompany.com"
 )

 if err := database.AddRecipient(email1); err != nil {
  t.Fatalf("Error adding recipient: %v", err)
 }
 if err := database.AddRecipient(email2); err != nil {
  t.Fatalf("Error adding recipient: %v", err)
 }

 list, err := database.ListRecipients()
 if err != nil {
  t.Fatalf("Error listing recipients: %v", err)
 }

 if len(list) != 2 {
  t.Errorf("Expected 2 recipients, got %d", len(list))
 }

 if err := database.RemoveRecipient(email1); err != nil {
  t.Fatalf("Error removing recipient: %v", err)
 }

 list, _ = database.ListRecipients()
 if len(list) != 1 {
  t.Errorf("Expected 1 recipient after removal, got %d", len(list))
 }
 if list[0] != email2 {
  t.Errorf("Expetect remaining email to be %s, got %s", email2, list[0]) 
 }
}
