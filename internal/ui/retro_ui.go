package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Caracteres para bordas e desenhos
const (
	// Bordas duplas
	TOP_LEFT     = "╔"
	TOP_RIGHT    = "╗"
	BOTTOM_LEFT  = "╚"
	BOTTOM_RIGHT = "╝"
	HORIZONTAL   = "═"
	VERTICAL     = "║"

	// Bordas simples
	SIMPLE_TOP_LEFT     = "┌"
	SIMPLE_TOP_RIGHT    = "┐"
	SIMPLE_BOTTOM_LEFT  = "└"
	SIMPLE_BOTTOM_RIGHT = "┘"
	SIMPLE_HORIZONTAL   = "─"
	SIMPLE_VERTICAL     = "│"

	// Sombra
	SHADOW = "░"

	// Conectores
	T_DOWN = "╦"
	T_UP   = "╩"
	CROSS  = "╬"

	COLOR_RESET   = "\033[0m"
	COLOR_BLACK   = "\033[30m"
	COLOR_RED     = "\033[31m"
	COLOR_GREEN   = "\033[32m"
	COLOR_YELLOW  = "\033[33m"
	COLOR_BLUE    = "\033[34m"
	COLOR_MAGENTA = "\033[35m"
	COLOR_CYAN    = "\033[36m"
	COLOR_WHITE   = "\033[37m"
	COLOR_GRAY    = "\033[90m"

	// Estilos
	BOLD    = "\033[1m"
	BLINK   = "\033[5m"
	REVERSE = "\033[7m"
)

var Floppy = DrawFloppyDisk()

func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// DrawFloppyDisk desenha um disquete ASCII mais detalhado
func DrawFloppyDisk() []string {
	return []string{
		" __________________________",
		"|  ______________________  |",
		"| | Dedicated to TSAI    | |",
		"| |______________________| |",
		"| |                      | |",
		"| |   [  INSERT DISK  ]  | |",
		"| |   [>READING...    ]  | |",
		"| |______________________| |",
		"|   [  ]            [##]   |",
		"|__________________________|",
	}
}

// DrawRetroHeader desenha o cabeçalho completo do sistema
func DrawRetroHeader() {
	ClearScreen()

	floppy := Floppy

	fmt.Print(COLOR_CYAN + BOLD)

	// Linha superior
	fmt.Print("╔")
	fmt.Print(strings.Repeat("═", 78))
	fmt.Print("╗\n")

	// Linha com logo e título
	fmt.Printf("║ %s  %s%s%sSERVERS UPDATER v1.0%s%s ║\n",
		floppy[0],
		COLOR_GREEN+BOLD, REVERSE,
		COLOR_RESET, COLOR_CYAN)

	// Linhas do disquete com espaçamento
	for i := 1; i < len(floppy); i++ {
		fmt.Printf("║ %s%s%s %s%s ║\n",
			COLOR_YELLOW, floppy[i], COLOR_RESET,
			strings.Repeat(" ", 56),
			COLOR_CYAN)
	}

	// Linha separadora
	fmt.Print("╠")
	fmt.Print(strings.Repeat("═", 78))
	fmt.Print("╣\n")

	// Status bar com data/hora
	now := time.Now()
	statusBar := fmt.Sprintf("System Ready - %s", now.Format("15:04:05 02/01/2006"))
	padding := 78 - len(statusBar)

	fmt.Printf("║ %s%s%s%s%s ║\n",
		COLOR_GREEN, statusBar, COLOR_RESET,
		strings.Repeat(" ", padding-1),
		COLOR_CYAN)

	// Linha inferior do cabeçalho
	fmt.Print("╚")
	fmt.Print(strings.Repeat("═", 78))
	fmt.Print("╝\n")

	fmt.Print(COLOR_RESET)
}

// DrawBox desenha uma caixa com bordas duplas e sombra
func DrawBox(title string, width, height int, withShadow bool) {
	fmt.Print(COLOR_BLUE + BOLD)

	// Linha superior
	fmt.Print(TOP_LEFT)
	if title != "" {
		titleLen := len(title)
		padding := (width - titleLen - 4) / 2
		fmt.Print(strings.Repeat(HORIZONTAL, padding))
		fmt.Printf(" %s%s%s%s ", COLOR_YELLOW+BOLD, title, COLOR_RESET, COLOR_BLUE+BOLD)
		fmt.Print(strings.Repeat(HORIZONTAL, width-titleLen-padding-4))
	} else {
		fmt.Print(strings.Repeat(HORIZONTAL, width-2))
	}
	fmt.Print(TOP_RIGHT)
	if withShadow {
		fmt.Print(COLOR_BLACK + SHADOW + SHADOW)
	}
	fmt.Println()

	// Linhas do meio
	for i := 0; i < height-2; i++ {
		fmt.Print(VERTICAL)
		fmt.Print(strings.Repeat(" ", width-2))
		fmt.Print(VERTICAL)
		if withShadow {
			fmt.Print(COLOR_BLACK + SHADOW + SHADOW)
		}
		fmt.Println()
	}

	// Linha inferior
	fmt.Print(BOTTOM_LEFT)
	fmt.Print(strings.Repeat(HORIZONTAL, width-2))
	fmt.Print(BOTTOM_RIGHT)
	if withShadow {
		fmt.Print(COLOR_BLACK + SHADOW + SHADOW)
		fmt.Println()
		fmt.Print("  " + strings.Repeat(SHADOW, width))
	}
	fmt.Println()

	fmt.Print(COLOR_RESET)
}

// DrawMenuOption desenha uma opção de menu estilizada
func DrawMenuOption(number int, label, description string, isSelected bool) {
	var color string
	var marker string

	if isSelected {
		color = COLOR_YELLOW + BOLD + REVERSE
		marker = "►"
	} else {
		color = COLOR_GREEN
		marker = " "
	}

	fmt.Printf("  %s%s %d. [%s%s%s%s] %s%s\n",
		color, marker, number,
		COLOR_CYAN+BOLD, label, COLOR_RESET,
		color, description, COLOR_RESET)
}

// DrawSeparator desenha um separador decorativo
func DrawSeparator(width int, char string) {
	fmt.Printf("%s%s%s\n",
		COLOR_BLUE,
		strings.Repeat(char, width),
		COLOR_RESET)
}

// DrawStatusBar desenha uma barra de status na parte inferior
func DrawStatusBar(message string) {
	width := 80

	fmt.Printf("\n%s%s", COLOR_BLUE, REVERSE)
	fmt.Printf(" %s", message)
	fmt.Print(strings.Repeat(" ", width-len(message)-1))
	fmt.Printf("%s\n", COLOR_RESET)
}

// PauseWithMessage exibe mensagem e aguarda input
func PauseWithMessage(message string) {
	fmt.Printf("\n%s%s%s\n", COLOR_YELLOW, message, COLOR_RESET)
	fmt.Print("Press [ENTER] to continue...")
	fmt.Scanln()
}

// DrawHeader desenha o cabeçalho com arte ASCII do disquete e título formatado
func DrawHeader(title string) {
	floppy := Floppy
	fmt.Print(COLOR_CYAN)
	for _, line := range floppy {
		fmt.Println(line)
	}
	fmt.Print(COLOR_RESET)

	width := 60
	separator := strings.Repeat("=", width)

	padding := (width - len(title) - 2) / 2
	if padding < 0 {
		padding = 0
	}

	titleLine := fmt.Sprintf("||%s %s %s||", strings.Repeat(" ", padding), title, strings.Repeat(" ", padding))

	fmt.Println(COLOR_YELLOW)
	fmt.Printf("╔%s╗\n", separator)
	fmt.Println(titleLine)
	fmt.Printf("╠%s╣\n", separator)
	fmt.Print(COLOR_RESET)
}

// LoadingBar exibe uma barra de progresso animada com texto descritivo
func LoadingBar(text string) {
	fmt.Printf("\n %s [", text)
	for i := 0; i < 20; i++ {
		fmt.Print("█")
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Println("] 100%")
	time.Sleep(500 * time.Millisecond)
}
