package main

import(
 "log"
 "servers_updater/internal/db"
 "servers_updater/internal/ui"
)

func main() {
 database, err := db.InitDB("updater.db")
 if err != nil {
  log.Fatalf("Failed to start DB: %v", err)	
 }
 defer database.Close()

 appMenu := ui.NewMenu(database)
 appMenu.ShowMainMenu()
}
