package validate_test

import (
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/validate"
)

func TestProjectNameValid(t *testing.T) {
	valid := []string{"demo", "my-service", "my_service", "My.Project", "demo2"}
	for _, v := range valid {
		if !validate.ProjectNameValid(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}

	invalid := []string{"", ".", "..", "../../etc", "a/b", "a b", "a&b", "a;b", "$(x)"}
	for _, v := range invalid {
		if validate.ProjectNameValid(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestGroupIDValid(t *testing.T) {
	valid := []string{"com.example", "org.example.foo", "com.my_company.svc-1", "io.github.user"}
	for _, v := range valid {
		if !validate.GroupIDValid(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}

	invalid := []string{"", "com example", "com/ex", "com&x", "com;x", "../etc"}
	for _, v := range invalid {
		if validate.GroupIDValid(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestPackageNameValid(t *testing.T) {
	valid := []string{"com.example", "com.custom.pkg.service-demo", "_foo.Bar", "a.b.c"}
	for _, v := range valid {
		if !validate.PackageNameValid(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}

	invalid := []string{"", "1abc", "com..example", ".com", "com.", "com/ex", "com example"}
	for _, v := range invalid {
		if validate.PackageNameValid(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestJavaVersionValid(t *testing.T) {
	valid := []string{"21", "17", "11", "21.0.1", "8"}
	for _, v := range valid {
		if !validate.JavaVersionValid(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}

	invalid := []string{"", "java21", "21.0.1-rc", "v21", "21 & rm"}
	for _, v := range invalid {
		if validate.JavaVersionValid(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestShellSafeEditor(t *testing.T) {
	valid := []string{"vim", "nano", "/usr/bin/vim", "code", "C:\\Windows\\System32\\notepad.exe"}
	for _, v := range valid {
		if !validate.ShellSafe.MatchString(v) {
			t.Errorf("expected %q to be shell-safe", v)
		}
	}

	invalid := []string{"vim; nc evil 4444", "vim --wait", "nano $HOME", "vim$(id)", "`cat`", "vim|sh", "vim&"}
	for _, v := range invalid {
		if validate.ShellSafe.MatchString(v) {
			t.Errorf("expected %q to be rejected as unsafe", v)
		}
	}
}

func TestValidateProjectConfig(t *testing.T) {
	err := validate.ValidateProjectConfig(
		"demo", "com.example", "demo", "com.example.demo",
		"maven-project", "jar", "21",
	)
	if err != nil {
		t.Errorf("expected valid config to pass, got: %v", err)
	}

	err = validate.ValidateProjectConfig(
		"../../etc", "com.example", "demo", "com.example.demo",
		"maven-project", "jar", "21",
	)
	if err == nil {
		t.Error("expected path traversal project name to be rejected")
	}

	err = validate.ValidateProjectConfig(
		"demo", "com.example&x", "demo", "com.example.demo",
		"maven-project", "jar", "21",
	)
	if err == nil {
		t.Error("expected URL-injection group ID to be rejected")
	}
}
