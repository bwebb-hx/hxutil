package utils

import (
	"fmt"
	"log"
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
