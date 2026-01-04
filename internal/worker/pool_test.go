package worker

import (
 "context"
 "servers_updater/internal/db"
 "testing"
)

func TestPoolOrchestration(t *testing.T) {
 ctx := context.Background()
 hosts := []string{"srv1","srv2","srv3"}

 database, _ := db.InitDB("test_pool.db")
 t.Run("Pool execution with test hosts.", func(t *testing.T) {
  err := Run(ctx, hosts, database)
  if err != nil {
   t.Logf("Pool ended with expected connection errors.: %v", err)
  } else {
   t.Log("Pool finished successfully.")
  }
 })
} 


func TestPoolExecution(t *testing.T) {
 ctx := context.Background()
 hosts := []string{"host1","host2","host3"}
 dbPath := "test_pool.db"

 database,_ := db.InitDB(dbPath)
 if err != nil {
  t.Fatalf("The test database could not be initialized.: %v", err)
 }
 defer func() {
  database.Close()
  os.Remove(dbPath)
 }()

 err := Run(ctx,hosts,database)
 if err != nil {
  t.Log("Pool executed, but no errors returned (strange for fake hosts.)")
 }
}
