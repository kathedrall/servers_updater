package worker

import (
 "context"
 "servers_updater/internal/db"
 "testing"
)

func TestPoolExecution(t *testing.T) {
 ctx := context.Background()
 hosts := []string{"host1","host2","host3"}

 database,_ := db.InitDB("test_pool.db")
 defer database.Close()

 err := Run(ctx,hosts,database)
 if err != nil {
  t.Log("Pool executed, but no errors returned (strange for fake hosts.)")
 }



}
