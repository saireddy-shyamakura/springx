package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/initializr"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Spring Boot project",
	RunE: func(cmd *cobra.Command, args []string) error {

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Project name: ")

		name, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		name = strings.TrimSpace(name)

		zipFile, err := initializr.Download(name)
		if err != nil {
			return err
		}

		fmt.Printf("Downloaded project to %s\n", zipFile)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}