# Yummy — Your Command-Line Recipe Companion

<div align="center">
  <img src="./assets/yummy_logo.svg" alt="Yummy Logo" />

  ![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/GarroshIcecream/yummy/blob/master/LICENSE)
  [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)
  [![Go Report Card](https://goreportcard.com/badge/github.com/GarroshIcecream/yummy)](https://goreportcard.com/report/github.com/GarroshIcecream/yummy)
  ![Development Status](https://img.shields.io/badge/status-in%20development-orange.svg)
  [![CI](https://github.com/GarroshIcecream/yummy/actions/workflows/ci.yml/badge.svg)](https://github.com/GarroshIcecream/yummy/actions/workflows/ci.yml)
  [![Release Status](https://github.com/GarroshIcecream/yummy/actions/workflows/release.yml/badge.svg)](https://github.com/GarroshIcecream/yummy/actions/workflows/release.yml)
</div>

> A fast, delightful command-line application for managing recipes. Built with care and powered by Bubble Tea, Yummy brings a beautiful terminal-first experience to every home cook, developer, and recipe curator.

## ✨ Why Yummy Stands Out

- **🎨 Polished Terminal UI**: A modern, accessible TUI built with Bubble Tea that feels intuitive and responsive
- **⚡ Lightweight & Fast**: Zero bloat, instant startup, and smooth navigation across large recipe collections
- **💾 Portable Storage**: Recipes saved locally in simple, exportable formats (JSON/CSV), making backups and sharing effortless
- **🔄 Focus on Workflow**: Quick commands for adding, searching, categorizing and exporting recipes — spend less time managing and more time cooking
- **🔧 Extensible Design**: Modular packages (cmd, config, db, scrape, themes, tui, utils) make it easy to extend features or integrate with other tools

## 🚀 Core Features

- **Recipe Management**: Add, edit, and organize recipes with ingredient lists, measures, instructions, and metadata
- **Powerful Search**: Quick search and categorization to find the recipe you need
- **Export Options**: Export collections to JSON or CSV for sharing or migration
- **Clean TUI**: Navigable interface with list/detail views, editable forms, and status indicators
- **Customizable Configuration**: JSON-based configuration system for themes, key bindings, chat settings, and more
- **Developer Friendly**: Small codebase with clear package boundaries — ideal for contributors and experimentation

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/GarroshIcecream/yummy.git
cd yummy

# Build the application
go build -o yummy

# Run it
./yummy
```

### Using Go Install

```bash
go install github.com/GarroshIcecream/yummy@latest
```

### Add recipe from URL (recipe scraping)

The **Add recipe from URL** feature uses [recipe-scrapers](https://github.com/hhursev/recipe-scrapers) (Python) for best coverage of recipe sites.

- **Python 3** must be on your system. Many macOS and Linux systems already have it; if not, install from [python.org](https://www.python.org/downloads/) or your package manager (e.g. `brew install python`).
- The **recipe-scrapers** package is **auto-installed** the first time you add a recipe from a URL. You do not need to run `pip install` yourself. If your system Python is **externally managed** (PEP 668, e.g. Homebrew Python on macOS), the app will create a small venv at `~/.yummy/recipe-scrapers-venv` and use it automatically.

If the app cannot find Python, set the path in config (e.g. `~/.yummy/config.json`):

```json
"add_recipe_from_url_dialog": {
  "python_path": "/usr/bin/python3"
}
```

Use the path to the Python where you want the package installed (or leave empty to use `python3` / `python` from your PATH). If auto-install still fails (e.g. no network), install manually: `python3 -m pip install --user recipe-scrapers`, or point `python_path` to a venv that has it.

## ⚙️ Configuration

Yummy stores its configuration in `~/.yummy/config.json`. The configuration file is automatically created with default values on first run.

### Key Features

- **Theme Selection**: Choose from default, dark, light, monokai, or solarized themes
- **Chat Customization**: Configure Ollama model, temperature, viewport size, and more
- **Key Binding Customization**: Remap any key combination to your preference
- **Database Settings**: Configure auto-backup intervals and retention
- **General Settings**: Debug mode, log levels, and UI preferences

## 📁 Project Structure

```
yummy/
├── main.go                 # Entry point
├── yummy/
│   ├── cmd/                # Cobra CLI (root, export, import)
│   ├── config/             # Config loading, keybindings
│   ├── consts/             # Constants
│   ├── db/                 # GORM + SQLite (cookbook, session_log)
│   ├── log/                # Structured logging
│   ├── models/             # common (enums, TUIModel), msg (Bubble Tea messages)
│   ├── scrape/             # Recipe URL scraping (Python recipe-scrapers)
│   ├── themes/             # Theme registry, default, YAML loader
│   ├── tui/                # Bubble Tea TUI
│   │   ├── chat/           # AI chat (Ollama), executor, tools, mentions
│   │   ├── detail/         # Recipe detail view, cooking mode
│   │   ├── dialog/         # Modals (theme, session, model, add-from-URL, etc.)
│   │   ├── edit/           # Recipe editor
│   │   ├── list/           # Recipe list, filters, autocomplete
│   │   ├── main_menu/      # Main menu
│   │   └── status/         # Status bar
│   ├── utils/              # Recipe, ingredient, measures helpers
│   └── version/            # Build-time version info
├── examples/
│   ├── sample_recipes/     # Sample recipe data
│   └── themes/             # YAML theme examples
└── assets/                 # Logo, etc.
```

## 🛠️ Development

### Development Workflow

```bash
# Run all tests
go test ./...

# Run a single package test
go test ./yummy/recipe -run TestName

# Format & fix imports
gofmt -w . && goimports -w .

# Lint (recommended)
golangci-lint run
```

## 🤝 Contributing

Contributions are warmly welcomed! The project favors small, well-documented pull requests that improve UX, add tests, or refine the TUI. Please open issues for larger proposals so we can align on design.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📧 Contact

Questions, ideas, or recipes to share? Email [garroshicecream@gmail.com](mailto:garroshicecream@gmail.com)

---

<div align="center">
  <strong>Cook boldly. Ship deliciousness.</strong>
</div>
