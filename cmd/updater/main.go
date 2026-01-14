 package main

 import(
   "context"
   "log"
   "servers_updater/internal/db"
   "servers_updater/internal/ui"
)

func main() {
 database, err := db.New("updater.db")
 if err != nil {
  log.Fatal("Failed to start DB: %v, err")	
 }
 defer database.Close()

 appMenu := ui.NewMenu(database)
 appMenu.ShowMainMenu()
}
