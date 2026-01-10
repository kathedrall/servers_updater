 package main

 import(
   "context"
   "log"
   "servers_updater/internal/db"
   "servers_updater/internal/worker"
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

  ctx := context.Background()
  if err := worker.Run(ctx, hosts, database); err != nil {
   log.Printf("Process completed with errors.: %v", err)
  }
 }
