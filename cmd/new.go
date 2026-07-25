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
	Use: "new",
	RunE: func(cmd *cobra.Command, args []string) error {

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Project name: ")

		name, _ := reader.ReadString('\n')

		name = strings.TrimSpace(name)

		return initializr.Download(name)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
