package consts

// Mein menu constants
const (
	MainMenuLogoText = `
██╗   ██╗██╗   ██╗███╗   ███╗███╗   ███╗██╗   ██╗
╚██╗ ██╔╝██║   ██║████╗ ████║████╗ ████║╚██╗ ██╔╝
 ╚████╔╝ ██║   ██║██╔████╔██║██╔████╔██║ ╚████╔╝
  ╚██╔╝  ██║   ██║██║╚██╔╝██║██║╚██╔╝██║  ╚██╔╝
   ██║   ╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║   ██║
   ╚═╝    ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝   ╚═╝`
)

// Ollama help messages
const (
	OllamaNotInstalledHelp = `ollama is not installed or not found in PATH.

To fix this:
1. Install Ollama from https://ollama.ai
2. Make sure Ollama is added to your system PATH
3. Restart your terminal/command prompt
4. Try running this application again

For more help, visit: https://ollama.ai/install`

	OllamaServiceNotRunningHelp = `ollama service is not running and could not be started automatically.

To fix this:
1. Start the Ollama service manually by running: ollama serve
2. Or restart your computer if Ollama is set to start automatically
3. Make sure no firewall is blocking Ollama
4. Check if there are any error messages in the Ollama logs
5. Try running this application again`
)
