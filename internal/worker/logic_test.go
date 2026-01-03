package worker

import (
 "context"
 "errors"
 "servers_updater/internal/domain"
 "testing"
)

type MockSSH struct {
 Output string
 Err error
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
 database, _ := db.InitDB("test_logic.db")

 t.Run("Scenario: Connection or command failure", func(t *testing.T){
  mock := &MockSSH{Err: errors.New("timeout ssh")}
  
  err := ProcessSingleServer(ctx, "host-err", database, mock)
  if err ! = nil {
   t.Error("No return on failure")
  }
  if !mock.Closed {
   t.Error("The SSH client should have been closed. (defer client.Close())")
  }
  })

 t.Run("Scenario: No update packages available.", func(t *testing.T) {
  mock := &MockSSH{Output: "0 upgraded, 0 newly installed, 0 to remove"}

  err := ProcessSingleServer(ctx, "host-clean", database, mock)
  if err != nil {
   t.Errorf("It should not return an error for a clean system.: %v", err)
  }
 })
}
