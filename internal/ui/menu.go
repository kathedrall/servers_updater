package ui

import (
	"context"
	"fmt"
	"servers_updater/internal/db"
	"servers_updater/internal/worker"
	"strings"
	"time"
)

type Menu struct {
	DB *db.BoltDB
}

// NewMenu cria uma nova instância do menu principal com conexão ao banco de dados
func NewMenu(database *db.BoltDB) *Menu {
	return &Menu{
		DB: database,
	}
}

// ShowMainMenu exibe o menu principal do sistema com todas as opções disponíveis
func (m *Menu) ShowMainMenu() {
	for {
		DrawRetroHeader()
		fmt.Println()

		// Box principal do menu
		fmt.Print(COLOR_BLUE + BOLD)
		width := 60

		// Título do menu
		fmt.Print("    " + TOP_LEFT)
		titleText := " SISTEMA DE GESTAO - MENU PRINCIPAL "
		padding := (width - len(titleText)) / 2
		fmt.Print(strings.Repeat(HORIZONTAL, padding))
		fmt.Printf("%s%s%s%s", COLOR_YELLOW+BOLD, titleText, COLOR_RESET, COLOR_BLUE+BOLD)
		fmt.Print(strings.Repeat(HORIZONTAL, width-len(titleText)-padding))
		fmt.Print(TOP_RIGHT + SHADOW + SHADOW)
		fmt.Println()

		// Linha separadora
		fmt.Print("    " + VERTICAL)
		fmt.Print(strings.Repeat(" ", width-2))
		fmt.Print(VERTICAL + COLOR_BLACK + SHADOW + SHADOW + COLOR_RESET)
		fmt.Println()

		// Opções do menu
		options := []struct {
			key   int
			label string
			desc  string
		}{
			{1, "SCAN", "Infrastructure Discovery"},
			{2, "EMAIL", "SMTP Configuration"},
			{3, "LATENCY", "Network Diagnosis"},
			{4, "RUN", "Execute Updates"},
			{0, "SAIR", "Exit System"},
		}

		for _, opt := range options {
			fmt.Printf("    %s  %s%d.%s [%s%s%s] %s%s %s%s%s\n",
				COLOR_BLUE+BOLD+VERTICAL,
				COLOR_WHITE, opt.key, COLOR_RESET,
				COLOR_CYAN+BOLD, opt.label, COLOR_RESET,
				COLOR_GREEN, opt.desc,
				COLOR_BLUE+BOLD, VERTICAL+COLOR_BLACK+SHADOW+SHADOW+COLOR_RESET)
		}

		// Linha separadora
		fmt.Print("    " + COLOR_BLUE + BOLD + VERTICAL)
		fmt.Print(strings.Repeat(" ", width-2))
		fmt.Print(VERTICAL + COLOR_BLACK + SHADOW + SHADOW + COLOR_RESET)
		fmt.Println()

		// Linha inferior
		fmt.Print("    " + COLOR_BLUE + BOLD + BOTTOM_LEFT)
		fmt.Print(strings.Repeat(HORIZONTAL, width-2))
		fmt.Print(BOTTOM_RIGHT + COLOR_BLACK + SHADOW + SHADOW + COLOR_RESET)
		fmt.Println()
		fmt.Print("      " + COLOR_BLACK + strings.Repeat(SHADOW, width) + COLOR_RESET)
		fmt.Println()

		// Prompt de entrada
		fmt.Printf("\n%s[DIGITE A OPCAO]%s > ", COLOR_YELLOW+BOLD, COLOR_RESET)

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			var discard string
			fmt.Scanln(&discard)
			continue
		}

		switch choice {
		case 1:
			m.screenConfig()
		case 2:
			ShowEmailConfiguration(m.DB)
		case 3:
			m.screenNetworkTest()
		case 4:
			m.screenRunUpdate()
		case 0:
			fmt.Println("\nEjetando disquete... Até logo!")
			time.Sleep(1 * time.Second)
			return
		default:
			fmt.Println(COLOR_RED + "\n [ERRO] Opção Inválida!" + COLOR_RESET)
			time.Sleep(1 * time.Second)
		}
	}
}

// screenNetworkTest executa diagnóstico de rede ICMP para análise de latência e sugestão de pool de workers
func (m *Menu) screenNetworkTest() {
	ClearScreen()
	DrawRetroHeader()

	// Box principal de diagnóstico de rede
	fmt.Print(COLOR_BLUE + BOLD)
	fmt.Print("    ╔═══════════════ DIAGNOSTICO DE REDE (ICMP) ═══════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║                   ANALISE DE LATENCIA                   ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

	// Box de status do teste
	fmt.Print(COLOR_YELLOW + BOLD)
	fmt.Print("    ╔══════════════════ INICIANDO TESTE ════════════════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ " + COLOR_WHITE + "Sondando latência da rede..." + COLOR_RESET + COLOR_YELLOW + BOLD + "                      ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ " + COLOR_CYAN + "Enviando pacotes ICMP..." + COLOR_RESET + COLOR_YELLOW + BOLD + "                       ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

	// Loading animation
	LoadingBar("Calculando métricas de rede")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	profile, err := worker.AnalyzeNetwork(ctx, "")
	if err != nil {
		// Box de erro
		fmt.Print(COLOR_RED + BOLD)
		fmt.Print("    ╔═══════════════════ ERRO CRITICO ══════════════════════════╗" + COLOR_BLACK + "░░\n")
		fmt.Printf("    ║ Erro no teste de rede: %v                              ║%s░░\n", err, COLOR_BLACK)
		fmt.Print("    ║ Verifique a conectividade de rede                      ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")
		PauseWithMessage("Pressione [ENTER] para voltar...")
		return
	}

	// Box de resultados
	fmt.Print(COLOR_GREEN + BOLD)
	fmt.Print("    ╔═══════════════════ RESULTADOS ════════════════════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Printf("    ║ %s[LATENCIA MEDIA]%s %v                                  ║%s░░\n",
		COLOR_CYAN, COLOR_RESET+COLOR_GREEN+BOLD, profile.RTT, COLOR_BLACK)
	fmt.Printf("    ║ %s[POOL SUGERIDO]%s %d Workers simultâneos               ║%s░░\n",
		COLOR_CYAN, COLOR_RESET+COLOR_GREEN+BOLD, profile.SuggestedWorkers, COLOR_BLACK)
	fmt.Print("    ║                                                          ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ " + COLOR_YELLOW + BOLD + "Configuração baseada na análise de rede:" + COLOR_RESET + COLOR_GREEN + BOLD + "            ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ " + COLOR_WHITE + "• Pool otimizado para sua conexão" + COLOR_RESET + COLOR_GREEN + BOLD + "                 ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ " + COLOR_WHITE + "• Previne sobrecarga da rede" + COLOR_RESET + COLOR_GREEN + BOLD + "                    ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

	// Box de confirmação
	fmt.Print(COLOR_BLUE + BOLD)
	fmt.Print("    ╔═══════════════════ ACAO NECESSARIA ═══════════════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ " + COLOR_YELLOW + BOLD + "Deseja salvar esta configuração no banco?" + COLOR_RESET + COLOR_BLUE + BOLD + "          ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║                                                          ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║  " + COLOR_WHITE + "[S]" + COLOR_RESET + " " + COLOR_GREEN + "Sim, salvar configuração" + COLOR_BLUE + BOLD + "                  ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║  " + COLOR_WHITE + "[N]" + COLOR_RESET + " " + COLOR_GREEN + "Não, descartar" + COLOR_BLUE + BOLD + "                           ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")

	fmt.Printf("\n%s[OPCAO]%s > ", COLOR_YELLOW+BOLD, COLOR_RESET)

	var confirm string
	fmt.Scanln(&confirm)

	if confirm == "s" || confirm == "S" {
		if err := m.DB.SavePoolLimit(profile.SuggestedWorkers); err != nil {
			// Box de erro no salvamento
			fmt.Print(COLOR_RED + BOLD)
			fmt.Print("    ╔═══════════════════ ERRO DE GRAVACAO ══════════════════════╗" + COLOR_BLACK + "░░\n")
			fmt.Printf("    ║ Erro ao salvar: %v                                    ║%s░░\n", err, COLOR_BLACK)
			fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
			fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")
		} else {
			// Box de sucesso
			fmt.Print(COLOR_GREEN + BOLD)
			fmt.Print("    ╔═══════════════════ CONFIGURACAO SALVA ═══════════════════╗" + COLOR_BLACK + "░░\n")
			fmt.Printf("    ║ %s[SUCESSO]%s Pool configurado para %d workers!          ║%s░░\n",
				COLOR_GREEN, COLOR_RESET+COLOR_GREEN+BOLD, profile.SuggestedWorkers, COLOR_BLACK)
			fmt.Print("    ║ A configuração será aplicada nos próximos scans        ║" + COLOR_BLACK + "░░\n")
			fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
			fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")
		}
		time.Sleep(2 * time.Second)
	}

	PauseWithMessage("Pressione [ENTER] para retornar ao menu principal...")
}

// screenRunUpdate executa o processo de atualização dos servidores (funcionalidade em construção)
func (m *Menu) screenRunUpdate() {
	DrawHeader("EXECUTAR UPDATE")
	fmt.Println("\n [EM CONSTRUCAO] Aqui chamaremos worker.Run()...")
	fmt.Println("\n Pressione ENTER para voltar.")
	fmt.Scanln()
}
