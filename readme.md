# Yummy -- Your Command-Line Recipe Companion

<div align="center">
  <img src="./assets/yummy_logo.svg" alt="Yummy Logo" />

  ![Go Version](https://img.shields.io/badge/Go-1.27.2-blue.svg)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/GarroshIcecream/yummy/blob/master/LICENSE)
  [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)
  [![Lint](https://github.com/GarroshIcecream/yummy/actions/workflows/lint.yml/badge.svg)](https://github.com/GarroshIcecream/yummy/actions/workflows/lint.yml)
  ![Development Status](https://img.shields.io/badge/status-in%20development-orange.svg)
  [![CI](https://github.com/GarroshIcecream/yummy/actions/workflows/ci.yml/badge.svg)](https://github.com/GarroshIcecream/yummy/actions/workflows/ci.yml)
  [![Release Status](https://github.com/GarroshIcecream/yummy/actions/workflows/release.yml/badge.svg)](https://github.com/GarroshIcecream/yummy/actions/workflows/release.yml)
</div>

> A terminal-first cookbook manager with a Bubble Tea TUI, local SQLite storage, recipe import/export, URL scraping, theming, and a cooking assistant powered by Ollama or your ChatGPT subscription.

## Features

- Recipe browsing with list, detail, edit, rating, and cooking-mode views
- Local persistence in `~/.yummy` using SQLite databases
- Recipe import from Markdown or JSON
- Recipe export to Markdown
- Add-from-URL flow backed by Python `recipe-scrapers`
- Cooking chat with recipe-aware context, running on local Ollama or your ChatGPT Plus/Pro subscription
- Configurable key bindings, chat settings, and custom YAML themes

## Installation

### Homebrew (macOS)

```bash
brew tap GarroshIcecream/yummy
brew install --cask yummy
```

### From source

```bash
git clone https://github.com/GarroshIcecream/yummy.git
cd yummy
go build -o yummy
./yummy
```

### With Go install

```bash
go install github.com/GarroshIcecream/yummy@latest
```

## Runtime Requirements

### LLM provider

The cooking assistant runs on one of two providers, selected with `chat.provider` in `~/.yummy/config.json`:

- `"ollama"` (default): a local model served by Ollama.
- `"chatgpt"`: your ChatGPT Plus/Pro subscription, no API key needed.

The app opens even if the selected provider isn't ready (Ollama not running, not signed in to ChatGPT); the chat explains what's wrong, and you can sign in or switch provider from the command palette.

### Ollama

When using Ollama, the TUI checks it during startup. Make sure Ollama is installed, running, and has the configured model available before launching `yummy`.

By default, the app uses `gemma3:4b`.

Example:

```bash
ollama serve
ollama pull gemma3:4b
```

### ChatGPT subscription

From the app: open the command palette (`ctrl+p`) and choose **Switch LLM Provider** → **ChatGPT**. If you aren't signed in yet, a browser opens to sign in, then the app switches over. To sign out, select ChatGPT in the same dialog and press `x`.

From the terminal (opens a browser; uses a local callback on port 1455 or 1457):

```bash
yummy login chatgpt
```

The provider choice is saved in `~/.yummy/config.json`:

```json
{
  "chat": {
    "provider": "chatgpt",
    "chatgpt": {
      "default_model": ""
    }
  }
}
```

- `default_model` empty = first model your plan offers. The model selector lists the models available on your account.
- `models` (optional) is a fallback list used if the model catalog can't be fetched.
- The session is stored in `~/.yummy/auth/chatgpt.json` (readable only by you) and refreshed automatically. `yummy logout chatgpt` removes it.

> **Note:** this uses the same sign-in flow and backend as OpenAI's Codex CLI. It is not a documented public API, so it may change or break without notice.

### Add recipe from URL

The add-from-URL flow uses [recipe-scrapers](https://github.com/hhursev/recipe-scrapers).

- Python 3 must be installed
- `recipe-scrapers` is auto-installed on first use
- On externally managed Python installs, Yummy creates and uses `~/.yummy/recipe-scrapers-venv`

When ChatGPT is the active provider and recipe-scrapers can't read a page (unsupported site, no recipe schema, or no Python), the dialog offers to let ChatGPT read the page instead: press `enter` again and Yummy downloads the page and asks ChatGPT to pull the recipe out of it. This sends the page's text to ChatGPT and uses your subscription's quota.

If Python is not auto-detected, set it in `~/.yummy/config.json`:

```json
{
  "add_recipe_from_url_dialog": {
    "python_path": "/usr/bin/python3"
  }
}
```

## Usage

Run the TUI:

```bash
yummy
```

Enable debug logging:

```bash
yummy -d
```

Export a recipe to Markdown:

```bash
yummy export 123
```

Import a recipe from Markdown or JSON:

```bash
yummy import recipe.md
yummy import recipe.json --name "Weeknight Pasta"
```

## Configuration

Yummy stores its configuration in `~/.yummy/config.json` and creates it automatically on first run.

Common things to customize:

- `theme`
- `chat.default_model`
- key bindings under `keymap`
- database names and retention settings under `database`
- dialog sizes and UI behavior

Yummy ships with a built-in `default` theme and can load custom YAML themes from `~/.yummy/themes`. Example theme files live in `examples/themes/`.

## Data Layout

Yummy stores app data under `~/.yummy/`, including:

- `config.json`
- `cookbook.db`
- `session_log.db`
- `themes/`
- `recipe-scrapers-venv/` when needed for URL imports

## Project Structure

```text
yummy/
|- main.go
|- assets/
|- examples/
|- internal/
|  |- cmd/          # Cobra commands (root, import, export)
|  |- config/       # Config defaults and loading
|  |- db/           # SQLite-backed cookbook and session log
|  |- log/          # Structured logging setup
|  |- models/       # Shared model interfaces and messages
|  |- scrape/       # URL scraping integration
|  |- themes/       # Default theme and YAML theme loading
|  |- tui/          # Bubble Tea application and views
|  |- utils/        # Recipe parsing and helpers
|  `- version/      # Build-time version info
|- Taskfile.yml     # Primary local task runner
`- .goreleaser.yaml
```

## Development

Useful commands:

```bash
task deps
task check
task ci
task test
task test-coverage
task build
task fmt
task lint
task commitizen:install
task commit
task commit:check
task version:next
task tag
task release-snapshot
```

Install Task with your preferred package manager, for example:

```bash
brew install go-task/tap/go-task
```

Run command help locally:

```bash
go run . --help
go run . import --help
go run . export --help
```

## Commit Workflow

Commit messages are standardized with Commitizen using the repo config in `.cz.toml`.

Install Commitizen into the repo-local tools venv:

```bash
task commitizen:install
```

Create an interactive conventional commit:

```bash
task commit
```

Validate commit messages, defaulting to the latest commit:

```bash
task commit:check
```

Validate a wider range when needed:

```bash
task commit:check RANGE=main..HEAD
```

## Release Tags

This repo uses `svu` with the config in `.svu.yaml` and the release workflow triggers on tags matching `v*`.

Preview the next version:

```bash
task version:next
```

Create the next local tag automatically from commit history:

```bash
task tag
```

The tag tasks require a clean working tree.

Or force a specific bump:

```bash
task tag:patch
task tag:minor
task tag:major
```

Then push the tag to trigger the release workflow:

```bash
git push origin <tag>
```

## Contributing

Contributions are welcome. Small, focused pull requests are easiest to review.

1. Fork the repository
2. Create your branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push the branch (`git push origin feature/amazing-feature`)
5. Open a pull request

## License

This project is licensed under the MIT License. See `LICENSE`.

## Contact

Questions, ideas, or recipes to share? Email [garroshicecream@gmail.com](mailto:garroshicecream@gmail.com)

---

<div align="center">
  <strong>Cook boldly. Ship deliciousness.</strong>
</div>
