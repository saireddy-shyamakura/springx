package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/ui"
)

// ProjectConfig holds all user-configurable parameters for generating
// a Spring Boot project via Spring Initializr.
type ProjectConfig struct {
	ProjectName  string
	GroupID      string
	ArtifactID   string
	PackageName  string
	BuildTool    string
	Packaging    string
	JavaVersion  string
	Dependencies []string
}

type promptOption struct {
	ID   string
	Name string
}

// PromptForConfig walks the user through a series of interactive prompts
// and returns a fully populated ProjectConfig. All inputs are trimmed and
// validated before returning.
func PromptForConfig() (*ProjectConfig, error) {
	meta, err := metadata.Fetch()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Spring Initializr metadata: %w", err)
	}
	return PromptForConfigIO(os.Stdin, os.Stdout, meta)
}

// PromptForConfigIO performs prompting using the provided reader, writer, and metadata.
// This allows full unit testing of interactive terminal flows.
func PromptForConfigIO(r io.Reader, w io.Writer, meta *metadata.Metadata) (*ProjectConfig, error) {
	reader := bufio.NewReader(r)

	config := &ProjectConfig{}

	var err error

	// 1. Project name — required, no default.
	config.ProjectName, err = promptRequired(w, reader, "Project Name", "")
	if err != nil {
		return nil, err
	}

	// 2. Group ID — default: com.example.
	config.GroupID, err = promptRequired(w, reader, "Group ID", "com.example")
	if err != nil {
		return nil, err
	}

	// 3. Artifact ID — default: project name.
	config.ArtifactID, err = promptRequired(w, reader, "Artifact ID", config.ProjectName)
	if err != nil {
		return nil, err
	}

	// 4. Package name — default: groupId + "." + artifactId.
	defaultPackage := config.GroupID + "." + config.ArtifactID
	config.PackageName, err = promptRequired(w, reader, "Package Name", defaultPackage)
	if err != nil {
		return nil, err
	}

	// 5. Build tool — select project generation types from metadata.
	var buildOptions []promptOption
	var defaultBuildID string
	for _, val := range meta.Type.Values {
		if val.Action == "/starter.zip" {
			buildOptions = append(buildOptions, promptOption{ID: val.ID, Name: val.Name})
			if val.ID == meta.Type.Default {
				defaultBuildID = val.ID
			}
		}
	}
	if defaultBuildID == "" && len(buildOptions) > 0 {
		defaultBuildID = buildOptions[0].ID
	}
	config.BuildTool, err = promptSelectDynamic(w, reader, "Build Tool", buildOptions, defaultBuildID)
	if err != nil {
		return nil, err
	}

	// 6. Packaging — select from packaging metadata.
	var packagingOptions []promptOption
	var defaultPackagingID string
	for _, val := range meta.Packaging.Values {
		packagingOptions = append(packagingOptions, promptOption{ID: val.ID, Name: val.Name})
		if val.ID == meta.Packaging.Default {
			defaultPackagingID = val.ID
		}
	}
	if defaultPackagingID == "" && len(packagingOptions) > 0 {
		defaultPackagingID = packagingOptions[0].ID
	}
	config.Packaging, err = promptSelectDynamic(w, reader, "Packaging", packagingOptions, defaultPackagingID)
	if err != nil {
		return nil, err
	}

	// 7. Java version — select from javaVersion metadata.
	var javaOptions []promptOption
	var defaultJavaID string
	for _, val := range meta.JavaVersion.Values {
		javaOptions = append(javaOptions, promptOption{ID: val.ID, Name: val.Name})
		if val.ID == meta.JavaVersion.Default {
			defaultJavaID = val.ID
		}
	}
	if defaultJavaID == "" && len(javaOptions) > 0 {
		defaultJavaID = javaOptions[0].ID
	}
	config.JavaVersion, err = promptSelectDynamic(w, reader, "Java Version", javaOptions, defaultJavaID)
	if err != nil {
		return nil, err
	}

	// 8. Dependencies — interactive Bubble Tea TUI selection.
	deps, err := ui.RunDependencyPicker(meta)
	if err != nil {
		// Fallback for non-interactive/piped environments
		deps = []string{"web"}
	}
	config.Dependencies = deps

	// Print a summary of the configuration before proceeding.
	PrintSummary(w, config, meta)

	return config, nil
}

// promptRequired displays a prompt with an optional default value and reads
// a line of input. Empty input is rejected and the user is re-prompted.
// The returned value is trimmed of leading and trailing whitespace.
func promptRequired(w io.Writer, reader *bufio.Reader, label, defaultVal string) (string, error) {
	for {
		if defaultVal != "" {
			fmt.Fprintf(w, "%s [%s]: ", label, defaultVal)
		} else {
			fmt.Fprintf(w, "%s: ", label)
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)

		if input == "" {
			if defaultVal != "" {
				return defaultVal, nil
			}
			fmt.Fprintln(w, "  Value cannot be empty. Please try again.")
			continue
		}

		return input, nil
	}
}

// promptSelectDynamic displays a numbered list of options and prompts the user to
// choose one, returning the ID of the chosen option. The user may enter the option number,
// the option name, or the option ID case-insensitively. If the input is empty, the default
// option ID is returned.
func promptSelectDynamic(w io.Writer, reader *bufio.Reader, label string, options []promptOption, defaultID string) (string, error) {
	var defaultName string
	for _, opt := range options {
		if opt.ID == defaultID {
			defaultName = opt.Name
			break
		}
	}
	if defaultName == "" && len(options) > 0 {
		defaultName = options[0].Name
		defaultID = options[0].ID
	}

	for {
		fmt.Fprintf(w, "%s:\n", label)
		for i, opt := range options {
			mark := ""
			if opt.ID == defaultID {
				mark = " (default)"
			}
			fmt.Fprintf(w, "  %d. %s%s\n", i+1, opt.Name, mark)
		}
		fmt.Fprintf(w, "Choose [%s]: ", defaultName)

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)

		if input == "" {
			return defaultID, nil
		}

		// Check if the input matches an option name (case-insensitive).
		for _, opt := range options {
			if strings.EqualFold(input, opt.Name) {
				return opt.ID, nil
			}
		}

		// Check if the input matches an option ID (case-insensitive).
		for _, opt := range options {
			if strings.EqualFold(input, opt.ID) {
				return opt.ID, nil
			}
		}

		// Check if the input is a valid option number.
		var n int
		if _, err := fmt.Sscanf(input, "%d", &n); err == nil {
			if n >= 1 && n <= len(options) {
				return options[n-1].ID, nil
			}
		}

		fmt.Fprintln(w, "  Invalid choice. Please try again.")
	}
}

// PrintSummary displays the project configuration in a formatted table
// before the download begins. It resolves configuration IDs back to display
// names using the provided metadata.
func PrintSummary(w io.Writer, config *ProjectConfig, meta *metadata.Metadata) {
	buildToolName := config.BuildTool
	packagingName := config.Packaging
	javaVersionName := config.JavaVersion

	if meta != nil {
		for _, val := range meta.Type.Values {
			if val.ID == config.BuildTool {
				buildToolName = val.Name
				break
			}
		}
		for _, val := range meta.Packaging.Values {
			if val.ID == config.Packaging {
				packagingName = val.Name
				break
			}
		}
		for _, val := range meta.JavaVersion.Values {
			if val.ID == config.JavaVersion {
				javaVersionName = val.Name
				break
			}
		}
	}

	depsStr := "None"
	if len(config.Dependencies) > 0 {
		depsStr = strings.Join(config.Dependencies, ", ")
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "----------------------------------")
	fmt.Fprintln(w, "Project Configuration")
	fmt.Fprintln(w, "----------------------------------")
	fmt.Fprintf(w, "%-12s : %s\n", "Project Name", config.ProjectName)
	fmt.Fprintf(w, "%-12s : %s\n", "Group ID", config.GroupID)
	fmt.Fprintf(w, "%-12s : %s\n", "Artifact ID", config.ArtifactID)
	fmt.Fprintf(w, "%-12s : %s\n", "Package Name", config.PackageName)
	fmt.Fprintf(w, "%-12s : %s\n", "Build Tool", buildToolName)
	fmt.Fprintf(w, "%-12s : %s\n", "Packaging", packagingName)
	fmt.Fprintf(w, "%-12s : %s\n", "Java Version", javaVersionName)
	fmt.Fprintf(w, "%-12s : %s\n", "Dependencies", depsStr)
	fmt.Fprintln(w, "----------------------------------")
	fmt.Fprintln(w)
}