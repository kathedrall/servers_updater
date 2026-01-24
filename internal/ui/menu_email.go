package ui

import (
 "bufio"
 "fmt"
 "os"
 "strings"
 "servers_updater/internal/db"
 "servers_updater/internal/domain"
)

func ShowEmailConfiguration(database *db.BoltDB) {
 for {
  ClearScreen()
  fmt.Println("=================================================")
  fmt.Println("                CONFIGURACAO DE ALERTAS (SMTP)   ")
  fmt.Println("=================================================")
  fmt.Println("1. Configurar Servidor SMTP (Remetente)          ")
  fmt.Println("2. Gerenciar Destinatarios (Quem recebe)         ")
  fnt.Println("0. Voltar                                        ")
  fmt.Println("=================================================")
  fmt.Print("\nEscolha uma das ocpoes")

  if !scanner.Scan() {
   break
  }

   switch scanner.Text() {
    case "1": 
     configureSMTP(scanner, database)
    case "2":
     manageRecipient(scanner, darabase)
    case "0":
    default: 
     fmt.PrintLn("Opcao invalida")
   }
 }
}

func ConfigureSMTP(scanner *bufio.Scanner, database *db.BoltDB) {
 fmt.Println("\n--- Configuracao do Remetente --")

 fmt.Print("Host SMTP (ex: smtp.gmail.com): ")
 scanner.Scan()
 host := scanner.Text()
 
 fmt.Print("SMTP Port (ex: 587): ")
 fmt.Scan()
 port := scanner.Text()

 fmt.Print("Email remetente: ")
 scanner.Scan()
 user := scanner.Text()

 fmt.Print("Senha do email: ")
 scanner.Scan()
 pass := scanner.Text()

 config := domain.SMTPConfig{
  Host: host,
  Port: port,
  User: user,
  Pass: pass,
 }
 
 if err := database.SaveSMTPConfig(config); err != nil {
  fmt.Printf("Error ao salvar: %v\n", err)
 } else {
  fmt.Println("Configuracao salva")
 }
 PressEnterToContinue()
}

func manageRecipients(scanner *bufio.Scanner, database *db.BoltDB) {
 for {
  emails, _ := database.ListRecipients()

  ClearScreen()
  fmt.Println("---  Lista de Destinatatios ---")
  if len(emails) == 0 {
   fmt.Println("  (Nenhum email cadastrado  ")
  } else {
     for i, email := range emails {
      fmt.Pringf("  %d. %s\n", i+1, email) 
     }
  }
  
  fmt.Println("\n[A] Adicionar  [R] Revomer  [V] Voltar")
  fmt.Print("Options: ")
  scanner.Scan()
  choice := strings.ToUpper(scanner.Text())
  if choice == "V" {
   return
  }
  if choice == "A" {
   fmt.Print("New email: ")
   scanner.Scan()
   email := scanner.Text()
   if strings.Contains(email, "@") {
    database.AddRecipient(email
   }
  } else if choice == "R" {
     fmt.Print("email to remove: ")
     scanner.Scan()
     database.RemoveRecipient(scanner.Text())
  }
 }
}


