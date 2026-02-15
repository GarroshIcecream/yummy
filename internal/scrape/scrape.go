package scrape

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed scripts/fetch_recipe_json.py
var fetchRecipeScript []byte

const (
	moduleNotFound    = "No module named 'recipe_scrapers'"
	externallyManaged = "externally-managed-environment"
	venvDir           = "recipe-scrapers-venv"
)

func resolvePython(cfgPath string) (string, error) {
	if cfgPath != "" {
		return cfgPath, nil
	}
	if v := venvPython(); v != "" {
		return v, nil
	}
	for _, name := range []string{"python3", "python"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no Python found in PATH (try setting python_path in config)")
}

func venvPython() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, rel := range []string{filepath.Join("bin", "python"), filepath.Join("Scripts", "python.exe")} {
		p := filepath.Join(home, ".yummy", venvDir, rel)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func pipInstall(bin string) error {
	try := func(user bool) error {
		args := []string{"-m", "pip", "install"}
		if user {
			args = append(args, "--user")
		}
		args = append(args, "recipe-scrapers")
		cmd := exec.Command(bin, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			msg := strings.TrimSpace(stderr.String())
			if msg == "" {
				msg = err.Error()
			}
			return fmt.Errorf("%s", msg)
		}
		return nil
	}
	firstErr := try(true)
	if firstErr == nil {
		return nil
	}
	if try(false) == nil {
		return nil
	}
	return firstErr
}

func createVenv(systemPython string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".yummy", venvDir)
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return "", fmt.Errorf("create .yummy: %w", err)
	}
	if out, err := exec.Command(systemPython, "-m", "venv", dir).CombinedOutput(); err != nil {
		return "", fmt.Errorf("create venv: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	pip := filepath.Join(dir, "bin", "pip")
	py := filepath.Join(dir, "bin", "python")
	if _, err := os.Stat(py); err != nil && os.IsNotExist(err) {
		pip = filepath.Join(dir, "Scripts", "pip.exe")
		py = filepath.Join(dir, "Scripts", "python.exe")
	}
	cmd := exec.Command(pip, "install", "recipe-scrapers")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pip install in venv: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return py, nil
}

func runScript(bin, url string) ([]byte, error) {
	tmp, err := os.CreateTemp("", "fetch_recipe_json-*.py")
	if err != nil {
		return nil, fmt.Errorf("temp script: %w", err)
	}
	path := tmp.Name()
	defer func() { _ = os.Remove(path) }()
	if _, err := tmp.Write(fetchRecipeScript); err != nil {
		_ = tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	out, err := exec.Command(bin, path, url).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("%w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return nil, err
	}
	return out, nil
}

// ScrapeURL returns a recipe.Scraper for the given URL. pythonPath is optional (e.g. "python3" or "" for default).
// Auto-installs recipe-scrapers or creates ~/.yummy/recipe-scrapers-venv on PEP 668 systems.
func ScrapeURL(url string, pythonPath string) (Scraper, error) {
	raw, err := ScrapeURLRaw(url, pythonPath)
	if err != nil {
		return nil, err
	}
	var j RecipeScrapersJSON
	if err := json.Unmarshal([]byte(raw), &j); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	return &adapter{j: j}, nil
}

// ScrapeURLRaw returns the recipe JSON string. Used by the gopy bindings.
func ScrapeURLRaw(url string, pythonPath string) (string, error) {
	bin, err := resolvePython(pythonPath)
	if err != nil {
		return "", err
	}
	out, err := runScript(bin, url)
	if err == nil {
		return string(out), nil
	}
	if !strings.Contains(err.Error(), moduleNotFound) {
		return "", err
	}
	runErr := err
	installErr := pipInstall(bin)
	if installErr == nil {
		out, err = runScript(bin, url)
		if err == nil {
			return string(out), nil
		}
		return "", err
	}
	if pythonPath == "" {
		managed := installErr.Error()
		if strings.Contains(managed, externallyManaged) || strings.Contains(managed, "externally managed") {
			venvPy, createErr := createVenv(bin)
			if createErr != nil {
				return "", fmt.Errorf("%v (venv: %v)", runErr, createErr)
			}
			out, err = runScript(venvPy, url)
			if err == nil {
				return string(out), nil
			}
			return "", err
		}
	}
	return "", fmt.Errorf("%v (install: %v)", runErr, installErr)
}
