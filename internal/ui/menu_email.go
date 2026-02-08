package ui

import (
	"bufio"
	"fmt"
	"os"
	"servers_updater/internal/db"
	"servers_updater/internal/domain"
	"strings"
	"time"
)

func ShowEmailConfiguration(database *db.BoltDB) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		DrawRetroHeader()

		// Verificar configuração SMTP atual
		smtpConfig, smtpErr := database.GetSMTPConfig()
		recipients, _ := database.ListRecipient()

		// Box de informações do sistema de email
		fmt.Print(COLOR_BLUE + BOLD)
		fmt.Print("    ╔═══════════════ SMTP MAILER SYSTEM ════════════════╗" + COLOR_BLACK + "░░\n")

		if smtpErr == nil && smtpConfig.Host != "" {
			fmt.Printf("    ║ %s[INFO]%s Servidor SMTP: %s                     ║%s░░\n",
				COLOR_GREEN, COLOR_RESET+COLOR_BLUE+BOLD, smtpConfig.Host, COLOR_BLACK)
			fmt.Printf("    ║ %s[INFO]%s Email remetente: %s                   ║%s░░\n",
				COLOR_GREEN, COLOR_RESET+COLOR_BLUE+BOLD, smtpConfig.User, COLOR_BLACK)
		} else {
			fmt.Printf("    ║ %s[WARN]%s SMTP não configurado                   ║%s░░\n",
				COLOR_YELLOW, COLOR_RESET+COLOR_BLUE+BOLD, COLOR_BLACK)
		}

		fmt.Printf("    ║ %s[INFO]%s %d destinatários cadastrados            ║%s░░\n",
			COLOR_GREEN, COLOR_RESET+COLOR_BLUE+BOLD, len(recipients), COLOR_BLACK)

		fmt.Print("    ║                                                  ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║ " + COLOR_YELLOW + BOLD + "[ACOES DISPONIVEIS]:" + COLOR_RESET + COLOR_BLUE + BOLD + "                      ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║  " + COLOR_WHITE + "1." + COLOR_RESET + " [" + COLOR_CYAN + BOLD + "SERVER" + COLOR_RESET + "] " + COLOR_GREEN + "Configurar SMTP (Gmail/Outros)" + COLOR_BLUE + BOLD + "   ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║  " + COLOR_WHITE + "2." + COLOR_RESET + " [" + COLOR_CYAN + BOLD + "LISTA" + COLOR_RESET + "] " + COLOR_GREEN + "Gerenciar Destinatários" + COLOR_BLUE + BOLD + "        ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║  " + COLOR_WHITE + "0." + COLOR_RESET + " [" + COLOR_CYAN + BOLD + "SAIR" + COLOR_RESET + "] " + COLOR_GREEN + "Voltar ao Menu Principal" + COLOR_BLUE + BOLD + "        ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")

		fmt.Printf("\n%s[OPCAO]%s > ", COLOR_YELLOW+BOLD, COLOR_RESET)

		scanner.Scan()
		switch scanner.Text() {
		case "1":
			configureSMTP(scanner, database)
		case "2":
			manageRecipients(scanner, database)
		case "0":
			return
		default:
			fmt.Printf("%s[ERRO]%s Opção inválida%s\n", COLOR_RED+BOLD, COLOR_RESET, COLOR_RESET)
			time.Sleep(1 * time.Second)
		}
	}
}

func configureSMTP(scanner *bufio.Scanner, database *db.BoltDB) {
	ClearScreen()
	DrawRetroHeader()

	// Box principal de configuração SMTP
	fmt.Print(COLOR_BLUE + BOLD)
	fmt.Print("    ╔═══════════════ CONFIGURACAO SERVIDOR SMTP ═══════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║                    CONFIGURACAO DE EMAIL                ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

	// Box de entrada de dados
	fmt.Print(COLOR_YELLOW + BOLD)
	fmt.Print("    ╔═══════════════════ DADOS SMTP ════════════════════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ " + COLOR_WHITE + "Preencha os dados do servidor SMTP:" + COLOR_RESET + COLOR_YELLOW + BOLD + "              ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

	fmt.Printf("%sHost SMTP%s (ex: smtp.gmail.com): ", COLOR_CYAN+BOLD, COLOR_RESET)
	scanner.Scan()
	host := scanner.Text()

	fmt.Printf("%sPorta SMTP%s (ex: 587): ", COLOR_CYAN+BOLD, COLOR_RESET)
	scanner.Scan()
	port := scanner.Text()

	fmt.Printf("%sEmail Remetente%s: ", COLOR_CYAN+BOLD, COLOR_RESET)
	scanner.Scan()
	user := scanner.Text()

	fmt.Printf("%sSenha do App%s: ", COLOR_CYAN+BOLD, COLOR_RESET)
	scanner.Scan()
	pass := scanner.Text()

	config := domain.SMTPConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
	}

	// Box de resultado
	if err := database.SaveSMTPConfig(config); err != nil {
		fmt.Print(COLOR_RED + BOLD)
		fmt.Print("    ╔═══════════════════ ERRO CRITICO ══════════════════════════╗" + COLOR_BLACK + "░░\n")
		fmt.Printf("    ║ Erro ao salvar configuração: %v                        ║%s░░\n", err, COLOR_BLACK)
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")
	} else {
		fmt.Print(COLOR_GREEN + BOLD)
		fmt.Print("    ╔═══════════════════ SUCESSO ═══════════════════════════════╗" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║ Configuração SMTP salva com sucesso!                   ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")
	}

	PauseWithMessage("Pressione [ENTER] para continuar...")
}

func manageRecipients(scanner *bufio.Scanner, database *db.BoltDB) {
	for {
		emails, _ := database.ListRecipient()
		ClearScreen()
		DrawRetroHeader()

		// Box principal de gerenciamento de destinatários
		fmt.Print(COLOR_BLUE + BOLD)
		fmt.Print("    ╔═══════════════ GERENCIAR DESTINATARIOS ═══════════════════╗" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║                  LISTA DE EMAILS                        ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ╠══════════════════════════════════════════════════════════╣" + COLOR_BLACK + "░░\n")

		if len(emails) == 0 {
			fmt.Print("    ║ " + COLOR_YELLOW + "Nenhum email cadastrado" + COLOR_RESET + COLOR_BLUE + BOLD + "                          ║" + COLOR_BLACK + "░░\n")
		} else {
			fmt.Printf("    ║ %s[INFO]%s %d emails cadastrados:%s                        ║%s░░\n",
				COLOR_GREEN, COLOR_RESET+COLOR_BLUE+BOLD, len(emails), COLOR_RESET+COLOR_BLUE+BOLD, COLOR_BLACK)
			fmt.Print("    ║                                                          ║" + COLOR_BLACK + "░░\n")
			for i, email := range emails {
				fmt.Printf("    ║  %s%d.%s %s%s                                         ║%s░░\n",
					COLOR_WHITE+BOLD, i+1, COLOR_RESET+COLOR_BLUE+BOLD, COLOR_CYAN, email,
					fmt.Sprintf("%-*s", 50-len(email)-len(fmt.Sprintf("%d. ", i+1)), ""), COLOR_BLACK)
			}
		}

		fmt.Print("    ║                                                          ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║ " + COLOR_YELLOW + BOLD + "[ACOES DISPONIVEIS]:" + COLOR_RESET + COLOR_BLUE + BOLD + "                              ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║  " + COLOR_WHITE + "[A]" + COLOR_RESET + " [" + COLOR_CYAN + BOLD + "ADD" + COLOR_RESET + "] " + COLOR_GREEN + "Adicionar Email" + COLOR_BLUE + BOLD + "                    ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║  " + COLOR_WHITE + "[R]" + COLOR_RESET + " [" + COLOR_CYAN + BOLD + "REM" + COLOR_RESET + "] " + COLOR_GREEN + "Remover Email" + COLOR_BLUE + BOLD + "                     ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║  " + COLOR_WHITE + "[V]" + COLOR_RESET + " [" + COLOR_CYAN + BOLD + "BACK" + COLOR_RESET + "] " + COLOR_GREEN + "Voltar" + COLOR_BLUE + BOLD + "                          ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")

		fmt.Printf("\n%s[OPCAO]%s > ", COLOR_YELLOW+BOLD, COLOR_RESET)
		scanner.Scan()
		choice := strings.ToUpper(scanner.Text())
		if choice == "V" {
			return
		}
		if choice == "A" {
			fmt.Printf("\n%sNovo Email%s: ", COLOR_CYAN+BOLD, COLOR_RESET)
			scanner.Scan()
			email := scanner.Text()
			if strings.Contains(email, "@") {
				database.AddRecipient(email)
				fmt.Printf("%s[SUCESSO]%s Email %s adicionado!\n", COLOR_GREEN+BOLD, COLOR_RESET, email)
			} else {
				fmt.Printf("%s[ERRO]%s Email inválido!\n", COLOR_RED+BOLD, COLOR_RESET)
			}
			time.Sleep(1500 * time.Millisecond)
		} else if choice == "R" {
			fmt.Printf("\n%sRemover Email%s: ", COLOR_CYAN+BOLD, COLOR_RESET)
			scanner.Scan()
			email := scanner.Text()
			database.RemoveRecipient(email)
			fmt.Printf("%s[SUCESSO]%s Email %s removido!\n", COLOR_GREEN+BOLD, COLOR_RESET, email)
			time.Sleep(1500 * time.Millisecond)
		}
	}
}
