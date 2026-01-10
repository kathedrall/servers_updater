package worker

import (
 "os"	
 "context"
 "errors"
 "servers_updater/internal/db"
 "servers_updater/internal/domain"
 "testing"
)

type MockSSH struct {
 Output string
 Err    error
 Closed bool
}

func (m *MockSSH) ExecuteCommand(cmd string) (string, error) {
 return m.Output, m.Err
} 

func (m *MockSSH) Close() error {
 m.Closed = true
 return nil 
}

func TestProcessingSingleServerFlow(t *testing.T) {
 ctx := context.Background()
 dbPath := "logic_unit_test.db"
 database, _ := db.InitDB(dbPath)
 defer func() { database.Close(); os.Remove(dbPath) }()

 t.Run("Scenario: System already updated", func(t *testing.T) {
 var mock domain.SSHClient = &MockSSH {
   Output: "0 upgraded, 0 newly installed, 0 to remove",
   Err: nil,
  }
 
  err := ProcessSingleServer(ctx, "localhost", database, mock)
  if err != nil {
   t.Errorf("It should not return an error for a clean system.: %v", err)
  }
 })
 
 t.Run("Scenario: Command execution error", func(t *testing.T) {
 var mock domain.SSHClient = &MockSSH {
   Output: "",
   Err: errors.New("sudo: password required"),
  }

  err := ProcessSingleServer(ctx, "localhost", database, mock)
  if err != nil {
   t.Error("It should have returned an error. The ssh command failed.")
  }
 })

 t.Run("Scenario: Connection or command failure", func(t *testing.T){
  mock := &MockSSH {
   Output: "",
   Err: errors.New("timeout ssh"),
  }
  
  err := ProcessSingleServer(ctx, "host-err", database, mock)
  if err != nil {
   t.Error("No return on failure")
  }
  if !mock.Closed {
   t.Error("The SSH client should have been closed. (defer client.Close())")
  }
 })
}
