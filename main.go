package main

import (
	"log"

	"github.com/bwebb-hx/hxutil/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const genDocs = true

func main() {
	if genDocs {
		generateDoc(cmd.RootCmd)
		return
	}
	cmd.Execute()
}

func generateDoc(cmd *cobra.Command) {
	// generate markdown
	err := doc.GenMarkdownTree(cmd, "./docs")
	if err != nil {
		log.Fatal("failed to generate docs:", err)
	}
	log.Println("generated docs!")
}
