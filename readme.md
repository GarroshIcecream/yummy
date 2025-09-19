# Yummy — Your Command-Line Recipe Companion

<div align="center">
  <img src="./assets/yummy_logo.svg" alt="Yummy Logo" />
  
  [![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)](https://golang.org/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![Development Status](https://img.shields.io/badge/status-in%20development-orange.svg)](https://github.com/GarroshIcecream/yummy)
  [![Go Report Card](https://goreportcard.com/badge/github.com/GarroshIcecream/yummy)](https://goreportcard.com/report/github.com/GarroshIcecream/yummy)
  [![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/GarroshIcecream/yummy)
  [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)
  [![GitHub stars](https://img.shields.io/github/stars/GarroshIcecream/yummy.svg?style=social&label=Star)](https://github.com/GarroshIcecream/yummy)
</div>

> A fast, delightful command-line application for managing recipes. Built with care and powered by Bubble Tea, Yummy brings a beautiful terminal-first experience to every home cook, developer, and recipe curator.

## ✨ Why Yummy Stands Out

- **🎨 Polished Terminal UI**: A modern, accessible TUI built with Bubble Tea that feels intuitive and responsive
- **⚡ Lightweight & Fast**: Zero bloat, instant startup, and smooth navigation across large recipe collections
- **💾 Portable Storage**: Recipes saved locally in simple, exportable formats (JSON/CSV), making backups and sharing effortless
- **🔄 Focus on Workflow**: Quick commands for adding, searching, categorizing and exporting recipes — spend less time managing and more time cooking
- **🔧 Extensible Design**: Modular internal packages (recipe, tui, db, tools) make it easy to extend features or integrate with other tools

## 🚀 Core Features

- **Recipe Management**: Add, edit, and organize recipes with ingredient lists, measures, instructions, and metadata
- **Powerful Search**: Quick search and categorization to find the recipe you need
- **Export Options**: Export collections to JSON or CSV for sharing or migration
- **Clean TUI**: Navigable interface with list/detail views, editable forms, and status indicators
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

## 🛠️ Development

### Prerequisites

- Go 1.24.3 or later
- Git

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