// Package prompt provides interactive text prompts for springx project configuration.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/config"
	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/templates"
	"github.com/saireddy-shyamakura/springx/internal/ui"
	"github.com/saireddy-shyamakura/springx/internal/validate"
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

	cfg, err := config.Load()
	if err != nil {
		def := config.DefaultConfig()
		cfg = &def
	}

	return PromptForConfigIO(os.Stdin, os.Stdout, meta, cfg)
}

// PromptForConfigWithTemplate is identical to PromptForConfig but seeds defaults
// and pre-selected dependencies from the named template before the prompts begin.
// The user may still modify every field after the template is applied.
func PromptForConfigWithTemplate(templateName string) (*ProjectConfig, error) {
	meta, err := metadata.Fetch()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Spring Initializr metadata: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		def := config.DefaultConfig()
		cfg = &def
	}

	tmpl, err := templates.Get(templateName)
	if err != nil {
		return nil, err
	}

	preSelected := templates.ApplyTemplate(tmpl, cfg)

	fmt.Printf("✔ Template %q loaded: %s\n\n", tmpl.Name, tmpl.Description)

	return PromptForConfigIOWithPreset(os.Stdin, os.Stdout, meta, cfg, preSelected, tmpl.Name)
}

// PromptForConfigIO performs prompting using the provided reader, writer, metadata, and config settings.
// This allows full unit testing of interactive terminal flows.
func PromptForConfigIO(r io.Reader, w io.Writer, meta *metadata.Metadata, cfg *config.Config) (*ProjectConfig, error) {
	return PromptForConfigIOWithPreset(r, w, meta, cfg, nil, "")
}

// PromptForConfigIOWithPreset is identical to PromptForConfigIO but also accepts a list of
// pre-selected dependency IDs that will be pre-checked in the Bubble Tea picker.
// Pass nil (or an empty slice) for no pre-selections.
// templateName is the human-readable template label shown in the picker status bar; pass "" for none.
func PromptForConfigIOWithPreset(r io.Reader, w io.Writer, meta *metadata.Metadata, cfg *config.Config, preSelected []string, templateName string) (*ProjectConfig, error) {
	if cfg == nil {
		def := config.DefaultConfig()
		cfg = &def
	}

	reader := bufio.NewReader(r)
	projectCfg := &ProjectConfig{}

	var err error

	// 1. Project name — required, no default.
	projectCfg.ProjectName, err = promptRequired(w, reader, "Project Name", "")
	if err != nil {
		return nil, err
	}
	if !validate.ProjectNameValid(projectCfg.ProjectName) {
		return nil, validateProjectNameError(projectCfg.ProjectName)
	}

	// 2. Group ID — default from config or com.example.
	defaultGroup := cfg.GroupID
	if defaultGroup == "" {
		defaultGroup = "com.example"
	}
	projectCfg.GroupID, err = promptRequired(w, reader, "Group ID", defaultGroup)
	if err != nil {
		return nil, err
	}
	if !validate.GroupIDValid(projectCfg.GroupID) {
		return nil, validateGroupIDError(projectCfg.GroupID)
	}

	// 3. Artifact ID — default: artifactPrefix + project name.
	defaultArtifact := projectCfg.ProjectName
	if cfg.ArtifactPrefix != "" {
		defaultArtifact = cfg.ArtifactPrefix + projectCfg.ProjectName
	}
	projectCfg.ArtifactID, err = promptRequired(w, reader, "Artifact ID", defaultArtifact)
	if err != nil {
		return nil, err
	}
	if !validate.ArtifactIDValid(projectCfg.ArtifactID) {
		return nil, validateArtifactIDError(projectCfg.ArtifactID)
	}

	// 4. Package name — default: packagePrefix (or groupId) + "." + artifactId.
	defaultPkgPrefix := cfg.PackagePrefix
	if defaultPkgPrefix == "" {
		defaultPkgPrefix = projectCfg.GroupID
	}
	defaultPackage := defaultPkgPrefix + "." + projectCfg.ArtifactID
	projectCfg.PackageName, err = promptRequired(w, reader, "Package Name", defaultPackage)
	if err != nil {
		return nil, err
	}
	if !validate.PackageNameValid(projectCfg.PackageName) {
		return nil, validatePackageNameError(projectCfg.PackageName)
	}

	// 5. Build tool — select project generation types from metadata.
	var buildOptions []promptOption
	var defaultBuildID string
	for _, val := range meta.Type.Values {
		if val.Action == "/starter.zip" {
			buildOptions = append(buildOptions, promptOption{ID: val.ID, Name: val.Name})
			if cfg.BuildTool != "" && (strings.EqualFold(val.ID, cfg.BuildTool) || strings.EqualFold(val.Name, cfg.BuildTool)) {
				defaultBuildID = val.ID
			} else if defaultBuildID == "" && val.ID == meta.Type.Default {
				defaultBuildID = val.ID
			}
		}
	}
	if defaultBuildID == "" && len(buildOptions) > 0 {
		defaultBuildID = buildOptions[0].ID
	}
	projectCfg.BuildTool, err = promptSelectDynamic(w, reader, "Build Tool", buildOptions, defaultBuildID)
	if err != nil {
		return nil, err
	}

	// 6. Packaging — select from packaging metadata.
	var packagingOptions []promptOption
	var defaultPackagingID string
	for _, val := range meta.Packaging.Values {
		packagingOptions = append(packagingOptions, promptOption{ID: val.ID, Name: val.Name})
		if cfg.Packaging != "" && (strings.EqualFold(val.ID, cfg.Packaging) || strings.EqualFold(val.Name, cfg.Packaging)) {
			defaultPackagingID = val.ID
		} else if defaultPackagingID == "" && val.ID == meta.Packaging.Default {
			defaultPackagingID = val.ID
		}
	}
	if defaultPackagingID == "" && len(packagingOptions) > 0 {
		defaultPackagingID = packagingOptions[0].ID
	}
	projectCfg.Packaging, err = promptSelectDynamic(w, reader, "Packaging", packagingOptions, defaultPackagingID)
	if err != nil {
		return nil, err
	}

	// 7. Java version — select from javaVersion metadata.
	var javaOptions []promptOption
	var defaultJavaID string
	for _, val := range meta.JavaVersion.Values {
		javaOptions = append(javaOptions, promptOption{ID: val.ID, Name: val.Name})
		if cfg.JavaVersion != "" && (strings.EqualFold(val.ID, cfg.JavaVersion) || strings.EqualFold(val.Name, cfg.JavaVersion)) {
			defaultJavaID = val.ID
		} else if defaultJavaID == "" && val.ID == meta.JavaVersion.Default {
			defaultJavaID = val.ID
		}
	}
	if defaultJavaID == "" && len(javaOptions) > 0 {
		defaultJavaID = javaOptions[0].ID
	}
	projectCfg.JavaVersion, err = promptSelectDynamic(w, reader, "Java Version", javaOptions, defaultJavaID)
	if err != nil {
		return nil, err
	}

	// 8. Dependencies — interactive Bubble Tea TUI selection.
	// Pass context already collected (boot version, java version) so the
	// picker status bar is fully populated from the first frame.
	bootVersion := ""
	if meta != nil {
		bootVersion = meta.BootVersion.Default
	}
	pickerOpts := ui.PickerOptions{
		Metadata:    meta,
		PreSelected: preSelected,
		BootVersion: bootVersion,
		JavaVersion: projectCfg.JavaVersion,
		Template:    templateName,
	}
	deps, err := ui.RunDependencyPickerWithOptions(pickerOpts)
	if err != nil {
		// Fallback for non-interactive/piped environments: honor template pre-selections
		// or fall back to the minimal "web" starter.
		if len(preSelected) > 0 {
			deps = preSelected
		} else {
			deps = []string{"web"}
		}
	}
	projectCfg.Dependencies = deps

	// Final gate: reject any configuration that would produce an unsafe
	// URL, filesystem path, or shell command downstream.
	if err := validate.ValidateProjectConfig(
		projectCfg.ProjectName,
		projectCfg.GroupID,
		projectCfg.ArtifactID,
		projectCfg.PackageName,
		projectCfg.BuildTool,
		projectCfg.Packaging,
		projectCfg.JavaVersion,
	); err != nil {
		return nil, err
	}

	// Print a summary of the configuration before proceeding.
	PrintSummary(w, projectCfg, meta)

	return projectCfg, nil
}

// validateProjectNameError builds a human-readable validation error.
func validateProjectNameError(v string) error {
	return fmt.Errorf("invalid project name %q: use only letters, digits, '.', '_' or '-'; must not be '.' or '..'", v)
}

func validateGroupIDError(v string) error {
	return fmt.Errorf("invalid group ID %q: use only letters, digits, '.', '_', '-' or ':'", v)
}

func validateArtifactIDError(v string) error {
	return fmt.Errorf("invalid artifact ID %q: use only letters, digits, '.', '_', '-' or ':'", v)
}

func validatePackageNameError(v string) error {
	return fmt.Errorf("invalid package name %q: use a valid Java package name (dot-separated identifiers)", v)
}

// promptRequired displays a prompt with an optional default value and reads
// a line of input. Empty input is rejected and the user is re-prompted.
// The returned value is trimmed of leading and trailing whitespace.
func promptRequired(w io.Writer, reader *bufio.Reader, label, defaultVal string) (string, error) {
	for {
		if defaultVal != "" {
			fmt.Fprintf(w, "%s [%s]: ", label, defaultVal) //nolint:errcheck
		} else {
			fmt.Fprintf(w, "%s: ", label) //nolint:errcheck
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
			fmt.Fprintln(w, "  Value cannot be empty. Please try again.") //nolint:errcheck
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
		fmt.Fprintf(w, "%s:\n", label) //nolint:errcheck
		for i, opt := range options {
			mark := ""
			if opt.ID == defaultID {
				mark = " (default)"
			}
			fmt.Fprintf(w, "  %d. %s%s\n", i+1, opt.Name, mark) //nolint:errcheck
		}
		fmt.Fprintf(w, "Choose [%s]: ", defaultName) //nolint:errcheck

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

		fmt.Fprintln(w, "  Invalid choice. Please try again.") //nolint:errcheck
	}
}

// PrintSummary displays the project configuration in a formatted table
// before the download begins. It resolves configuration IDs back to display
// names using the provided metadata.
func PrintSummary(w io.Writer, projectCfg *ProjectConfig, meta *metadata.Metadata) {
	buildToolName := projectCfg.BuildTool
	packagingName := projectCfg.Packaging
	javaVersionName := projectCfg.JavaVersion

	if meta != nil {
		for _, val := range meta.Type.Values {
			if val.ID == projectCfg.BuildTool {
				buildToolName = val.Name
				break
			}
		}
		for _, val := range meta.Packaging.Values {
			if val.ID == projectCfg.Packaging {
				packagingName = val.Name
				break
			}
		}
		for _, val := range meta.JavaVersion.Values {
			if val.ID == projectCfg.JavaVersion {
				javaVersionName = val.Name
				break
			}
		}
	}

	depsStr := "None"
	if len(projectCfg.Dependencies) > 0 {
		depsStr = strings.Join(projectCfg.Dependencies, ", ")
	}

	fmt.Fprintln(w)                                                        //nolint:errcheck
	fmt.Fprintln(w, "----------------------------------")                  //nolint:errcheck
	fmt.Fprintln(w, "Project Configuration")                               //nolint:errcheck
	fmt.Fprintln(w, "----------------------------------")                  //nolint:errcheck
	fmt.Fprintf(w, "%-12s : %s\n", "Project Name", projectCfg.ProjectName) //nolint:errcheck
	fmt.Fprintf(w, "%-12s : %s\n", "Group ID", projectCfg.GroupID)         //nolint:errcheck
	fmt.Fprintf(w, "%-12s : %s\n", "Artifact ID", projectCfg.ArtifactID)   //nolint:errcheck
	fmt.Fprintf(w, "%-12s : %s\n", "Package Name", projectCfg.PackageName) //nolint:errcheck
	fmt.Fprintf(w, "%-12s : %s\n", "Build Tool", buildToolName)            //nolint:errcheck
	fmt.Fprintf(w, "%-12s : %s\n", "Packaging", packagingName)             //nolint:errcheck
	fmt.Fprintf(w, "%-12s : %s\n", "Java Version", javaVersionName)        //nolint:errcheck
	fmt.Fprintf(w, "%-12s : %s\n", "Dependencies", depsStr)                //nolint:errcheck
	fmt.Fprintln(w, "----------------------------------")                  //nolint:errcheck
	fmt.Fprintln(w)                                                        //nolint:errcheck
}
