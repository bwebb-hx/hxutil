package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	ColorWarn    = color.New(color.FgHiYellow)
	ColorError   = color.New(color.FgHiRed)
	ColorFatal   = color.New(color.FgBlack, color.BgHiRed)
	ColorSuccess = color.New(color.FgHiGreen)
	ColorInfo    = color.New(color.FgHiMagenta)
	ColorLowkey  = color.New(color.FgHiBlack)
)

func GetInput(prompt string) string {
	var input string
	fmt.Print(prompt, ": ")
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Fatal("failed to read input:", err)
	}
	return input
}

// returns True if Yes, False otherwise. All prompts will have "[Y/n]" added to it.
func YesOrNo(prompt string) bool {
	if prompt != "" {
		prompt += " "
	}
	prompt += "[Y/n]"
	userInput := strings.ToLower(GetInput(prompt))

	if userInput == "y" || userInput == "yes" {
		return true
	}
	return false
}

func EnterToContinue() {
	fmt.Println("\nPress Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func Warn(header, desc string) {
	ColorWarn.Println("\n**Warning!", header)
	ColorWarn.Println("  " + desc)
}

func Error(header, desc string) {
	ColorError.Println("\n!!ERROR:", header)
	ColorError.Println("  " + desc)
}

func Info(header, desc string) {
	ColorInfo.Println("\nⓘ Info:", header)
	ColorInfo.Println("  " + desc)
}

func Fatal(header, desc string) {
	ColorFatal.Println("\nx FATAL:", header)
	ColorFatal.Println("  " + desc)
	os.Exit(1)
}

func Hint(text string) {
	ColorLowkey.Println("\n" + text)
}
