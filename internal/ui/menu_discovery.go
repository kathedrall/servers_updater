package ui

import (
	"bufio"
	"fmt"
	"os"
	"servers_updater/internal/domain"
	"servers_updater/internal/ssh"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
)

const LIMIT_WORKERS = 10

func (m *Menu) screenConfig() {
	for {
		DrawRetroHeader()
		dbMachines, _ := m.DB.GetAllMachines()
		sshMachines, err := ssh.LoadMachinesFromSSHConfig()

		// Box de informações
		fmt.Print(COLOR_BLUE + BOLD)
		fmt.Print("    ╔═══════════════ GESTAO DE HOSTS & DISCOVERY ════════════════╗" + COLOR_BLACK + "░░\n")
		fmt.Printf("    ║ %s[INFO]%s %d Hosts no banco de dados                    ║%s░░\n",
			COLOR_GREEN, COLOR_RESET+COLOR_BLUE+BOLD, len(dbMachines), COLOR_BLACK)

		if err == nil {
			fmt.Printf("    ║ %s[INFO]%s %d Hosts encontrados no ~/.ssh/config          ║%s░░\n",
				COLOR_GREEN, COLOR_RESET+COLOR_BLUE+BOLD, len(sshMachines), COLOR_BLACK)
		} else {
			fmt.Printf("    ║ %s[WARN]%s Erro ao ler ~/.ssh/config                    ║%s░░\n",
				COLOR_YELLOW, COLOR_RESET+COLOR_BLUE+BOLD, COLOR_BLACK)
		}

		fmt.Print("    ║                                                          ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║ " + COLOR_YELLOW + BOLD + "[ACOES DISPONIVEIS]:" + COLOR_RESET + COLOR_BLUE + BOLD + "                              ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║  " + COLOR_WHITE + "1." + COLOR_RESET + " [" + COLOR_CYAN + BOLD + "SCAN" + COLOR_RESET + "] " + COLOR_GREEN + "Iniciar Discovery (Atualizar status OS)" + COLOR_BLUE + BOLD + "     ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║  " + COLOR_WHITE + "2." + COLOR_RESET + " [" + COLOR_CYAN + BOLD + "NOVO" + COLOR_RESET + "] " + COLOR_GREEN + "Adicionar Host Manualmente" + COLOR_BLUE + BOLD + "            ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║  " + COLOR_WHITE + "0." + COLOR_RESET + " [" + COLOR_CYAN + BOLD + "SAIR" + COLOR_RESET + "] " + COLOR_GREEN + "Voltar ao Menu Principal" + COLOR_BLUE + BOLD + "              ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")

		fmt.Printf("\n%s[OPCAO]%s > ", COLOR_YELLOW+BOLD, COLOR_RESET)
		var opt int
		fmt.Scanln(&opt)

		switch opt {
		case 1:
			m.runDiscoveryRoutine()
		case 2:
			fmt.Println("Funcao de cadastro manual")
		case 0:
			return
		default:
			fmt.Println("Opcao invalida")
		}
	}
}

func (m *Menu) runDiscoveryRoutine() {
	ClearScreen()
	DrawRetroHeader()

	// Box principal do Discovery
	fmt.Print(COLOR_BLUE + BOLD)
	fmt.Print("    ╔═══════════════ TURBO DISCOVERY: VARREDURA ═════════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║                    DE REDE & SERVIDORES                  ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

	machines, err := ssh.LoadMachinesFromSSHConfig()
	if err != nil {
		fmt.Print(COLOR_YELLOW + BOLD)
		fmt.Print("    ╔══════════════════ AVISO IMPORTANTE ═══════════════════════╗" + COLOR_BLACK + "░░\n")
		fmt.Printf("    ║ %s[WARN]%s Erro ao ler ~/.ssh/config: %v"+COLOR_YELLOW+BOLD, COLOR_RED, COLOR_RESET+COLOR_YELLOW+BOLD)
		fmt.Print("          ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║ Tentando carregar do banco de dados...                  ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

		machines, err = m.DB.GetAllMachines()
		if err != nil || len(machines) == 0 {
			fmt.Print(COLOR_RED + BOLD)
			fmt.Print("    ╔══════════════════ ERRO CRITICO ═══════════════════════════╗" + COLOR_BLACK + "░░\n")
			fmt.Print("    ║ Nenhuma máquina encontrada no banco e no ~/.ssh/config   ║" + COLOR_BLACK + "░░\n")
			fmt.Print("    ║ Adicione hosts primeiro via ~/.ssh/config ou manual     ║" + COLOR_BLACK + "░░\n")
			fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
			fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")
			m.waitEnter()
			return
		}

		fmt.Print(COLOR_GREEN + BOLD)
		fmt.Print("    ╔═══════════════ MAQUINAS CARREGADAS ═══════════════════════╗" + COLOR_BLACK + "░░\n")
		fmt.Printf("    ║ %s[OK]%s Carregadas %d máquinas do banco de dados       ║%s░░\n",
			COLOR_GREEN, COLOR_RESET+COLOR_GREEN+BOLD, len(machines), COLOR_BLACK)
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")
	} else {
		fmt.Print(COLOR_GREEN + BOLD)
		fmt.Print("    ╔═══════════════ MAQUINAS CARREGADAS ═══════════════════════╗" + COLOR_BLACK + "░░\n")
		fmt.Printf("    ║ %s[OK]%s Carregadas %d máquinas do ~/.ssh/config        ║%s░░\n",
			COLOR_GREEN, COLOR_RESET+COLOR_GREEN+BOLD, len(machines), COLOR_BLACK)
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

		for _, machine := range machines {
			m.DB.SaveMachine(machine)
		}
	}

	limitWorkes := LIMIT_WORKERS
	if poolLimit, err := m.DB.GetPoolLimit(); err == nil && poolLimit > 0 {
		limitWorkes = poolLimit
	}

	// Box de configurações do scan
	fmt.Print(COLOR_CYAN + BOLD)
	fmt.Print("    ╔════════════════ CONFIGURACOES SCAN ═══════════════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Printf("    ║ Calibragem de rede (Pool): %s%d Workers simultaneos%s    ║%s░░\n",
		COLOR_GREEN, limitWorkes, COLOR_RESET+COLOR_CYAN+BOLD, COLOR_BLACK)
	fmt.Printf("    ║ Iniciando scan em %s%d servidores%s GNU Linux...          ║%s░░\n",
		COLOR_GREEN, len(machines), COLOR_RESET+COLOR_CYAN+BOLD, COLOR_BLACK)
	fmt.Print("    ║ " + COLOR_YELLOW + "OTIMIZACAO INTELIGENTE:" + COLOR_RESET + COLOR_CYAN + BOLD + "                          ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ • Credenciais salvas no banco (sem re-digitação)       ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ • Info do OS cached (evita re-discovery)               ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ • Senha solicitada apenas se conexão falhar            ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, limitWorkes)
	resultsChan := make(chan domain.Machine, len(machines))

	processedCount := 0
	var mu sync.Mutex
	start := time.Now()

	for _, mach := range machines {
		wg.Add(1)
		go func(targetMachine domain.Machine) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Primeiro, tentar carregar informações da máquina do banco (credenciais e OS)
			dbMachine, err := m.DB.GetMachineByHost(targetMachine.Host)
			if err == nil {
				// Usar credenciais salvas se disponíveis
				if targetMachine.Password == "" && dbMachine.Password != "" {
					targetMachine.Password = dbMachine.Password
				}
				// Usar informações do OS salvas para evitar re-discovery
				if dbMachine.OSName != "" {
					targetMachine.OSName = dbMachine.OSName
					targetMachine.OSVersion = dbMachine.OSVersion
					targetMachine.PrettyName = dbMachine.PrettyName
					targetMachine.IsSupported = dbMachine.IsSupported
				}
			}

			// Se não tem credenciais, vamos tentar conectar e ver o que acontece

			client, err := ssh.Connect(targetMachine)
			if err != nil {
				// Se falhou e não tem senha ainda, solicitar (independente de ter KeyPath)
				if targetMachine.Password == "" {
					mu.Lock()
					password := m.requestPasswordForMachine(&targetMachine)
					if password != "" {
						targetMachine.Password = password
						// Tentar conectar novamente com a senha
						client, err = ssh.Connect(targetMachine)
						if err == nil {
							// Salvar imediatamente no banco - FORÇA SEMPRE SALVAR
							if saveErr := m.DB.SaveMachine(targetMachine); saveErr != nil {
								// Silenciar erro para não poluir interface
							}
						}
					}
					mu.Unlock()
				}

				if err != nil {
					targetMachine.Status = "AUTH_FAIL"
					targetMachine.PrettyName = "-"
					targetMachine.IsSupported = false
				} else {
					// Conectou com sucesso
					defer client.Close()
					targetMachine.Status = "ONLINE"
					targetMachine.LastSeen = time.Now()

					// Identificar OS apenas se não temos informações salvas
					if targetMachine.OSName == "" {
						if err := ssh.IdentifyOS(client, &targetMachine); err != nil {
							targetMachine.PrettyName = "Error de leitura de OS"
						}
					}
				}
			} else {
				defer client.Close()
				targetMachine.Status = "ONLINE"
				targetMachine.LastSeen = time.Now()

				// Identificar OS apenas se não temos informações salvas
				if targetMachine.OSName == "" {
					if err := ssh.IdentifyOS(client, &targetMachine); err != nil {
						targetMachine.PrettyName = "Error de leitura de OS"
					}
				}
			}

			resultsChan <- targetMachine
			mu.Lock()
			processedCount++
			fmt.Printf("\r [Progress] %d/%d verifys....", processedCount, len(machines))
			mu.Unlock()
		}(mach)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
		close(semaphore)
	}()

	var updateMachines []domain.Machine
	for updated := range resultsChan {
		updateMachines = append(updateMachines, updated)
		m.DB.SaveMachine(updated)
	}

	duration := time.Since(start)
	ClearScreen()
	DrawRetroHeader()

	// Box de resultado final
	fmt.Print(COLOR_GREEN + BOLD)
	fmt.Print("    ╔═════════════════ SCAN COMPLETED ══════════════════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Printf("    ║ %s[SUCESSO]%s Scan finalizado em: %s                      ║%s░░\n",
		COLOR_GREEN, COLOR_RESET+COLOR_GREEN+BOLD, duration, COLOR_BLACK)
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")

	m.printDiscoveryGrid(updateMachines)

	// Box de ações disponíveis
	fmt.Print(COLOR_BLUE + BOLD)
	fmt.Print("    ╔══════════════════ ACOES DISPONIVEL ═══════════════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║ " + COLOR_YELLOW + BOLD + "[OPCOES]:" + COLOR_RESET + COLOR_BLUE + BOLD + "                                               ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║  " + COLOR_WHITE + "[ENTER]" + COLOR_RESET + " " + COLOR_GREEN + "Voltar ao Menu Principal" + COLOR_BLUE + BOLD + "                ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ║  " + COLOR_WHITE + "[R]" + COLOR_RESET + "     " + COLOR_GREEN + "Executar Novo Scan" + COLOR_BLUE + BOLD + "                     ║" + COLOR_BLACK + "░░\n")
	fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")

	fmt.Printf("\n%s[OPCAO]%s > ", COLOR_YELLOW+BOLD, COLOR_RESET)

	var input string
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input = scanner.Text()
	}

	if input == "r" || input == "R" {
		m.runDiscoveryRoutine()
	}
}

func (m *Menu) printDiscoveryGrid(machines []domain.Machine) {
	if len(machines) == 0 {
		fmt.Print(COLOR_YELLOW + BOLD)
		fmt.Print("    ╔═══════════════════ AVISO ══════════════════════════════════╗" + COLOR_BLACK + "░░\n")
		fmt.Print("    ║ Nenhuma máquina para exibir no relatório                ║" + COLOR_BLACK + "░░\n")
		fmt.Print("    ╚══════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
		fmt.Print("      " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n")
		return
	}

	// Contadores para estatísticas
	online, supported, failed := 0, 0, 0
	for _, machine := range machines {
		switch machine.Status {
		case "ONLINE":
			online++
			if machine.IsSupported {
				supported++
			}
		case "AUTH_FAIL":
			failed++
		}
	}

	// Box principal do relatório com grid interno
	fmt.Print(COLOR_BLUE + BOLD)
	fmt.Print("╔══════════════════════ RELATORIO INFRAESTRUTURA ═══════════════════════════════════════════╗" + COLOR_BLACK + "░░\n")
	fmt.Print("║                              ESTADO DOS SERVIDORES                                         ║" + COLOR_BLACK + "░░\n")
	fmt.Print("╠════════════════════════════════════════════════════════════════════════════════════════════╣" + COLOR_BLACK + "░░\n")

	// Box de estatísticas dentro do quadro
	fmt.Printf("║ %s[INFO]%s Total: %d | Online: %d | Falhas: %d | Suportadas: %d %s║%s░░\n",
		COLOR_CYAN, COLOR_RESET+COLOR_BLUE+BOLD, len(machines), online, failed, supported,
		fmt.Sprintf("%-*s", 35, ""), COLOR_BLACK)
	fmt.Print("╠════════════════════════════════════════════════════════════════════════════════════════════╣" + COLOR_BLACK + "░░\n")

	// Cabeçalho da tabela dentro do quadro
	fmt.Printf("║ %s%-18s | %-10s | %-25s | %-13s | %-8s | %-10s%s ║%s░░\n",
		COLOR_WHITE+BOLD, "HOST", "USER", "SISTEMA (DETECTED)", "STATUS", "SUPPORT", "JUMPER", COLOR_RESET+COLOR_BLUE+BOLD, COLOR_BLACK)
	fmt.Print("║═══════════════════════════════════════════════════════════════════════════════════════════║" + COLOR_BLACK + "░░\n")

	// Listagem das máquinas dentro do quadro
	for _, mac := range machines {
		support := COLOR_RED + "[--]" + COLOR_RESET
		if mac.IsSupported {
			support = COLOR_GREEN + "[OK]" + COLOR_RESET
		} else if mac.Status == "ONLINE" && !mac.IsSupported {
			support = COLOR_YELLOW + "[!!]" + COLOR_RESET
		}

		jumper := COLOR_GRAY + "[--]" + COLOR_RESET
		if mac.ProxyJumper != nil {
			jumper = COLOR_CYAN + *mac.ProxyJumper + COLOR_RESET
		}

		statusVis := mac.Status
		color := COLOR_RESET

		switch mac.Status {
		case "AUTH_FAIL":
			statusVis = COLOR_RED + "FAIL" + COLOR_RESET
			color = COLOR_RED
		case "ONLINE":
			statusVis = COLOR_GREEN + "[ONLINE]" + COLOR_RESET
			color = COLOR_GREEN
		}

		userStr := COLOR_GRAY + "---" + COLOR_RESET
		if mac.User != "" {
			userStr = COLOR_CYAN + mac.User + COLOR_RESET
		}

		fmt.Printf("║ %-18s | %-10s | %-25s | %-13s | %-8s | %-10s %s║%s░░\n",
			color+mac.Host+COLOR_RESET, userStr, mac.PrettyName, statusVis, support, jumper,
			COLOR_BLUE+BOLD, COLOR_BLACK)
	}

	fmt.Print("╠════════════════════════════════════════════════════════════════════════════════════════════╣" + COLOR_BLACK + "░░\n")

	// Box de legenda dentro do quadro
	fmt.Printf("║ %s[OK]%s Sistema suportado | %s[!!]%s Online não suportado | %s[--]%s Offline/Falha      ║%s░░\n",
		COLOR_GREEN, COLOR_RESET+COLOR_BLUE+BOLD, COLOR_YELLOW, COLOR_RESET+COLOR_BLUE+BOLD, COLOR_RED, COLOR_RESET+COLOR_BLUE+BOLD, COLOR_BLACK)

	fmt.Print("╚════════════════════════════════════════════════════════════════════════════════════════════╝" + COLOR_BLACK + "░░\n")
	fmt.Print("  " + COLOR_BLACK + "░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░" + COLOR_RESET + "\n\n")
}

func (m *Menu) waitEnter() {
	fmt.Println("\n Press [ENTER] for continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// requestPasswordForMachine solicita a senha para uma máquina específica
func (m *Menu) requestPasswordForMachine(machine *domain.Machine) string {
	fmt.Printf("\n%s[AUTENTICACAO NECESSARIA]%s Máquina: %s%s%s\n",
		COLOR_YELLOW+BOLD, COLOR_RESET, COLOR_CYAN+BOLD, machine.Host, COLOR_RESET)

	fmt.Printf("%s[INFO]%s Esta máquina não possui chave SSH configurada.\n",
		COLOR_BLUE, COLOR_RESET)
	fmt.Printf("%s[INFO]%s User: %s%s%s\n",
		COLOR_BLUE, COLOR_RESET, COLOR_CYAN, machine.User, COLOR_RESET)
	fmt.Printf("%s[BANCO]%s A senha será salva para uso futuro (não precisará digitar novamente)\n",
		COLOR_GREEN, COLOR_RESET)

	fmt.Printf("\n%sSenha para %s@%s%s: ",
		COLOR_GREEN+BOLD, machine.User, machine.Host, COLOR_RESET)

	// Lê a senha de forma segura (oculta)
	password, err := readPassword()
	if err != nil {
		fmt.Printf("%s[ERRO]%s Erro ao ler senha: %v\n", COLOR_RED+BOLD, COLOR_RESET, err)
		return ""
	}

	return password
}

// readPassword lê uma senha de forma segura (sem ecoar os caracteres)
func readPassword() (string, error) {
	fmt.Print("") // Força o flush do buffer
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println() // Nova linha após a entrada da senha
	return string(password), nil
}
