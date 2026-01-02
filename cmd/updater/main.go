 package main

 import(
   "context"
   "log"
   "servers_updater/internal/config"
   "servers_updater/internal/db"
   "servers_update/internal/worker"
)

func main() {

  database, err := db.New("updater.db")
  if err != nil {
    log.Fatal("Failed to start DB: %v, err")	
  }
  defer database.Close()

  
  hosts, err := config.ParseSSHConfig()
  if err != nil {
   log.Fatal("Error reading ssh config file.: %v, err")
  }

  ctx := conext.Background()
  if err := worker.Run(ctx, hosts, database) != nil {
   log.Printf("Process completed with errors.: %v, err")
  }

