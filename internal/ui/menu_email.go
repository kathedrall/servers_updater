package ui

import (
 "bufio"
 "fmt"
 "os"
 "strings"
 "servers_updater/internal/db"
 "servers_updater/internal/domain"
 "time"
)

func ShowEmailConfiguration(database *db.BoltDB) {
 scanner := bufio.NewScanner(os.Stdin)

 for {
  ClearScreen()
  DrawHeader("SMTP MAILER SYSTEM")
  drawEmailMenuOptions()
 

  if !scanner.Scan() {
   break
  }
		
  switch scanner.Text() {
  case "1":
   configureSMTP(scanner, database)
  case "2":
   manageRecipients(scanner, database)
  case "0":
   return
  default:
   fmt.Println(COLOR_RED + "Invalid Option" + COLOR_RESET)
   time.Sleep(1 * time.Second)
  }
 }
}

func drawEmailMenuOptions() {
 fmt.Println(COLOR_YELLOW)
 fmt.Println("╔══════════════════════════════════════════════╗")
 fmt.Println("║     PAINEL DE CONTROLE DE EMAIL              ║")
 fmt.Println("╠══════════════════════════════════════════════╣")
 fmt.Println("║                                              ║")
 fmt.Println("║  1. [SERVER] Configurar SMTP (Gmail/Outros)  ║")
 fmt.Println("║  2. [LISTA]  Gerenciar Destinatários         ║")
 fmt.Println("║  0. [VOLTAR] Retornar ao Menu Principal      ║")
 fmt.Println("║                                              ║")
 fmt.Println("╚══════════════════════════════════════════════╝")
 fmt.Print(COLOR_RESET + "\n [OPCAO] > ")
}




func configureSMTP(scanner *bufio.Scanner, database *db.BoltDB) {
 ClearScreen()
 DrawHeader("SMTP CONFIG")
 
 fmt.Print(COLOR_YELLOW + "\n-- [ EDIT DATA ] ---" + COLOR_RESET)
 fmt.Print("Host SMTP (ex: smtp.gmail.com): ")
  scanner.Scan()
 host := scanner.Text()
 fmt.Print("SMTP PORT (ex: 587): ")
  scanner.Scan()
  port := scanner.Text()
 fmt.Print("EMAIL SENDER: ")
  scanner.Scan()
  user := scanner.Text()
 fmt.Print("App Password: ")
  scanner.Scan()
  pass := scanner.Text()

 config := domain.SMTPConfig{
  Host:     host,
  Port:     port,
  User:     user,
  Password: pass, 
 }
 
 if err := database.SaveSMTPConfig(config); err != nil {
  fmt.Printf(COLOR_RED + "\nError while saving: %v\n" + COLOR_RESET, err)
 } else {
  fmt.Println(COLOR_GREEN + "\nConfiguration Saved!" + COLOR_RESET)
 }
  PressEnterToContinue()
}

func manageRecipients(scanner *bufio.Scanner, database *db.BoltDB) {
 for {
  emails, _ := database.ListRecipient()
  ClearScreen()
  DrawHeader("MANAGE LIST")
  fmt.Println(COLOR_YELLOW + "\n--- [ RECIPIENTS  ] ---" + COLOR_RESET)
   if len(emails) == 0 {
    fmt.Println("   (No email registered)")
   } else {
    for i, email := range emails {
     fmt.Printf("   %d. %s\n", i+1, email)
    }
   }

 fmt.Println("\n[A] Add  [R] Remove  [V] Back")
 fmt.Print("Options: ")
  scanner.Scan()
  choice := strings.ToUpper(scanner.Text())
  if choice == "V" { 
   return 
  }
  if choice == "A" {
   fmt.Print("New Emaill: ")
   scanner.Scan()
   email := scanner.Text()
   if strings.Contains(email, "@") {
    database.AddRecipient(email) 
   }
  } else if choice == "R" {
   fmt.Print("Remove Email: ")
   scanner.Scan()
   database.RemoveRecipient(scanner.Text())
  }
 }
}

func PressEnterToContinue() {
 fmt.Println("\nPress [Enter] for continue...")
 bufio.NewReader(os.Stdin).ReadBytes('\n')
}

