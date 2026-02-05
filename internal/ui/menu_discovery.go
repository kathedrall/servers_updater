package ui

import (
	"bufio"
	"fmt"
	"os"
	"servers_updater/internal/domain"
	"servers_updater/internal/ssh"
	"sync"
	"time"
)

const LIMIT_WORKERS = 10

func (m *Menu) screenConfig() {
	for {
		ClearScreen()
		DrawHeader("Gestao de hosts & Discovery")
		dbMachines, _ := m.DB.GetAllMachines()
		sshMachines, err := ssh.LoadMachinesFromSSHConfig()

		fmt.Printf("\n [INFO] %d Hosts no banco de dados", len(dbMachines))
		if err == nil {
			fmt.Printf("\n [INFO] %d Hosts encontrados no ~/.ssh/config", len(sshMachines))
		} else {
			fmt.Printf("\n [WARN] Erro ao ler ~/.ssh/config: %v", err)
		}

		fmt.Println("\n\n [ACOES DISPONIVEIS]:")
		fmt.Println(" 1. [SCAN] Iniciar Discovery (Atualizar status OS)")
		fmt.Println(" 2. [NOVO] Adicionar Host Manualmente")
		fmt.Println(" 0. Voltar ao menu Principal")

		fmt.Print("\n > ")
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

	fmt.Println("==============================================================")
	fmt.Println(" TURBO DISCOVERY: VARREDURA DE REDE")
	fmt.Println("==============================================================")

	machines, err := ssh.LoadMachinesFromSSHConfig()
	if err != nil {
		fmt.Printf(COLOR_YELLOW+"\n [WARN] Erro ao ler ~/.ssh/config: %v"+COLOR_RESET+"\n", err)
		fmt.Println(" Tentando carregar do banco de dados...")

		machines, err = m.DB.GetAllMachines()
		if err != nil || len(machines) == 0 {
			fmt.Println(COLOR_RED + "\n [!] Nenhuma máquina encontrada no banco e no ~/.ssh/config." + COLOR_RESET)
			fmt.Println(" Adicione hosts primeiro via ~/.ssh/config ou importação manual.")
			m.waitEnter()
			return
		}
		fmt.Printf(COLOR_GREEN+" [OK] Carregadas %d máquinas do banco de dados"+COLOR_RESET+"\n", len(machines))
	} else {
		fmt.Printf(COLOR_GREEN+" [OK] Carregadas %d máquinas do ~/.ssh/config"+COLOR_RESET+"\n", len(machines))

		for _, machine := range machines {
			m.DB.SaveMachine(machine)
		}
	}

	limitWorkes := LIMIT_WORKERS
	if poolLimit, err := m.DB.GetPoolLimit(); err == nil && poolLimit > 0 {
		limitWorkes = poolLimit
	}

	fmt.Printf("\n Calibragem de rede (Pool) %s%d Workers simultaneos%s", COLOR_GREEN, limitWorkes, COLOR_RESET)
	fmt.Printf("\n >> Iniciando scan em %d servidores Gnu Linux... \n\n", len(machines))

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

			client, err := ssh.Connect(targetMachine)
			if err != nil {
				targetMachine.Status = "AUTH_FAIL"
				targetMachine.PrettyName = "-"
				targetMachine.IsSupported = false
			} else {
				defer client.Close()
				targetMachine.Status = "ONLINE"
				targetMachine.LastSeen = time.Now()
				if err := ssh.IdentifyOS(client, &targetMachine); err != nil {
					targetMachine.PrettyName = "Error de leitura de OS"
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
	fmt.Printf("\n[ Scan completed on %s]\n", duration)

	m.printDiscoveryGrid(updateMachines)
	fmt.Println("\n What do you want ")
	fmt.Println(" [ENTER] return to menu | [R] Reload")

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
	fmt.Println("\nInfrastructure Report")
	fmt.Println("========================================================================================================")
	fmt.Printf("%-16s | %-10s | %-30s | %-12s | %s5s | %-5s\n", "HOST", "USER", "SISTEMA (DETECTED)", "STATUS", "SUP", "JUMPER PROXY")
	fmt.Println("========================================================================================================")

	for _, mac := range machines {
		support := "[--]"
		if mac.IsSupported {
			support = "[OK]"
		} else if mac.Status == "ONLINE" && !mac.IsSupported {
			support = "[!!]"
		}
		jumper := "[--]"
		if mac.ProxyJumper != nil {
			jumper = *mac.ProxyJumper
		}

		statusVis := mac.Status
		color := COLOR_RESET

		switch mac.Status {
		case "AUTH_FAIL":
			statusVis = "FAIL"
			color = COLOR_RED
		case "ONLINE":
			statusVis = "[ONLINE]"
			color = COLOR_GREEN
		}

		userStr := ""
		if mac.User != "" {
			userStr = mac.User
		}
		fmt.Printf("%s%-16s | %-10s | %-30s | %-12s | %-5s | %-5s%5s\n", color, mac.Host, userStr, mac.PrettyName, statusVis, support, jumper, COLOR_RESET)
	}
	fmt.Println("========================================================================================================")
}

func (m *Menu) waitEnter() {
	fmt.Println("\n Press [ENTER] for continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
