package ui

import (
	"context"
	"fmt"
	"servers_updater/internal/db"
	"servers_updater/internal/worker"
	"time"
)

type Menu struct {
	DB *db.BoltDB
}

func NewMenu(database *db.BoltDB) *Menu {
	return &Menu{
		DB: database,
	}
}

func (m *Menu) ShowMainMenu() {
	for {
		DrawHeader("SISTEMA DE GESTAO - MENU PRINCIPAL")
		fmt.Println(COLOR_YELLOW + "║                                           ║")
		fmt.Println("║  1. [CONFIG] Definir Credenciais Globais (SSH)           ║")
		fmt.Println("║  2. [REDE]   Calcular Workers (Throughput Test)          ║")
		fmt.Println("║  3. [RUN]    Executar Atualização em Lote                ║")
		fmt.Println("║  4. [SAIR]   Encerrar Sessão                             ║")
		fmt.Println("║                                                          ║" + COLOR_RESET)
		fmt.Println(COLOR_YELLOW + "╚══════════════════════════════════════════════════════════╝" + COLOR_RESET)

		fmt.Print("\n [DIGITE A OPCAO] > ")
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
			m.screenNetworkTest()
		case 3:
			m.screenRunUpdate()
		case 4:
			fmt.Println("\nEjetando disquete... Até logo!")
			time.Sleep(1 * time.Second)
			return
		default:
			fmt.Println(COLOR_RED + "\n [ERRO] Opção Inválida!" + COLOR_RESET)
			time.Sleep(1 * time.Second)
		}
	}
}

func (m *Menu) screenNetworkTest() {
	DrawHeader("DIAGNOSTICO DE REDE (ICMP)")
	fmt.Println("\n " + COLOR_YELLOW + "[AGUARDE] Sondando latência da rede..." + COLOR_RESET)
	
	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Print(" Enviando pacotes... ")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	profile, err := worker.AnalyzeNetwork(ctx, "") 
	if err != nil {
	  fmt.Printf(COLOR_RED+"\n\n [FALHA] Erro no teste de rede: %v\n"+COLOR_RESET, err)
	  fmt.Println(" Pressione ENTER para voltar...")
	  fmt.Scanln()
	return
	}

	LoadingBar("Calculando") 
	fmt.Printf("\n >> LATENCIA MEDIA: %v", profile.RTT)
	fmt.Printf("\n >> SUGESTAO POOL: " + COLOR_GREEN + "%d WORKERS" + COLOR_RESET, profile.SuggestedWorkers)

	fmt.Print("\n\n Deseja salvar esta configuração no banco? (s/n): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm == "s" || confirm == "S" {
	 
	if err := m.DB.SavePoolLimit(profile.SuggestedWorkers); err != nil {
	   fmt.Printf(COLOR_RED + "Erro ao salvar: %v" + COLOR_RESET, err)
	 } else {
	   fmt.Println(COLOR_GREEN + "\n [SUCESSO] Configuração gravada!" + COLOR_RESET)
	 }
	  time.Sleep(1500 * time.Millisecond)
       }
}

func (m *Menu) screenConfig() {
 DrawHeader("CONFIGURACOES GLOBAIS")
 fmt.Println("\n [EM CONSTRUCAO] Aqui pediremos User/Key SSH...")
 fmt.Println("\n Pressione ENTER para voltar.")
 fmt.Scanln()
}

func (m *Menu) screenRunUpdate() {
 DrawHeader("EXECUTAR UPDATE")
 fmt.Println("\n [EM CONSTRUCAO] Aqui chamaremos worker.Run()...")
 fmt.Println("\n Pressione ENTER para voltar.")
 fmt.Scanln()
}

