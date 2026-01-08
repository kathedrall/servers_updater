package worker

import (
 "context"
 "servers_updater/internal/db"
 "testing"
)

func TestPoolExecution(t *testing.T) {
 ctx := context.Background()
 hosts := []string{"host1","host2","host3"}
 dbPath := "test_pool.db"
 database, err := db.InitDB(dbPath)
 if err != nil {
  t.Fatalf("The test database could not be initialized.: %v", err)
 }
 
 defer func() {
  database.Close()
  os.Remove(dbPath)
 }()

 t.Run("Scenario: Setting concurrency limits", func(t *testing.T) {
  err := database.SavePoolLimit(5)
  if err != nil {
   t.Errorf("Error saving pool limit in BoltDB.: %v", err)
  }
 })

 t.Run("Scenario: Pool execution with test hosts", func(t *testing.T) {
  err := Run(ctx, hosts, database)
  if err != nil {
   t.Errorf("The pool returned an unexpected fatal error.: %v", err)
  } else {
   t.Log("The pool was successfully completed. Connection errors were successfully ignored.")
  }
 })
}
