package cmd

import (
	"fmt"
	"os"

	"github.com/saireddy-shyamakura/springx/internal/extract"
	"github.com/saireddy-shyamakura/springx/internal/initializr"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Spring Boot project",
	Long: `Create a new Spring Boot project via Spring Initializr.

Optionally pass --template to start from a pre-configured preset (run
'springx template list' to see all available templates). You may still
modify any field after the template defaults are applied.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName, _ := cmd.Flags().GetString("template")

		var (
			config *prompt.ProjectConfig
			err    error
		)

		if templateName != "" {
			config, err = prompt.PromptForConfigWithTemplate(templateName)
		} else {
			config, err = prompt.PromptForConfig()
		}
		if err != nil {
			return err
		}

		zipFile, err := initializr.Download(config)
		if err != nil {
			return err
		}

		fmt.Printf("Downloaded project to %s\n", zipFile)

		// Extract the downloaded ZIP archive into a directory named after the
		// project. This will create the directory if it does not already exist.
		if err := extract.Unzip(zipFile, config.ProjectName); err != nil {
			return fmt.Errorf("failed to extract project: %w", err)
		}

		// Remove the ZIP archive now that extraction is complete.
		if err := os.Remove(zipFile); err != nil {
			return fmt.Errorf("failed to remove zip file: %w", err)
		}

		fmt.Println("✔ Project created successfully!")

		return nil
	},
}

func init() {
	newCmd.Flags().StringP("template", "t", "", "Bootstrap from a built-in project template (e.g. rest-api, jpa, kafka)")
	rootCmd.AddCommand(newCmd)
}
