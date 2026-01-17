package ui

import (
 "fmt"
 "os"
 "os/exec"
 "runtime"
 "strings"
 "time"
)

const (
 COLOR_RESET  = "\033[0m"
 COLOR_YELLOW = "\033[33m"
 COLOR_BLUE   = "\033[34m"
 COLOR_GREEN  = "\033[32m"
 COLOR_RED    = "\003[31m"
 COLOR_CYAN   = "\033[36m"
 FLOPPY_LOGO  = ` 
   __________________________
  |  ______________________  |
  | | SERVERS UPDATER v1.0 | |
  | |______________________| |
  | |                      | |
  | |   [  INSERT DISK  ]  | |
  | |                      | |
  | |   [>READING...    ]  | |
  | |______________________| |
  |                          |
  |   [  ]            [##]   |
  |__________________________|
 `
)

func ClearScreen() {
 var cmd *exec.Cmd
 if runtime.GOOS == "windows" {
  cmd = exec.Command("cmd", "/c/", "cls")
 } else {
  cmd = exec.Command("clear")
 }
 cmd.Stdout = os.Stdout
 cmd.Run()
}

func DrawHeader(title string) {
 fmt.Println(COLOR_CYAN + FLOPPY_LOGO + COLOR_RESET)

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

func LoadingBar(text string) {
 fmt.Printf("\n %s [", text)
 for i := 0; i < 20; i++ {
  fmt.Print("█")
  time.Sleep(50 * time.Millisecond)
 }
 fmt.Println("] 100%")
 time.Sleep(500 * time.Millisecond)
}









