package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	agentTypes "github.com/pardnchiu/agenvoy/internal/agents/types"
)

const (
	colorOk      = "\033[32m" // ok — 綠
	colorWarn    = "\033[33m" // warn — 咖啡土黃
	colorError   = "\033[31m" // error — 紅
	colorConfirm = "\033[34m" // confirm — 藍
	colorNormal  = "\033[97m" // normal — 白
	colorHint    = "\033[90m" // hint — 灰
	colorReset   = "\033[0m"
)

func printTool(event agentTypes.Event) {
	var args map[string]any
	json.Unmarshal([]byte(event.ToolArgs), &args)

	switch event.ToolName {
	// case "read_file":
	// 	fmt.Printf("[*] Read File — \033[36m%s\033[0m\n", args["path"])
	// case "list_files":
	// 	fmt.Printf("[*] List Directory — \033[36m%s\033[0m\n", args["path"])
	// case "glob_files":
	// 	fmt.Printf("[*] Glob Files — \033[35m%s\033[0m\n", args["pattern"])
	case "write_file":
		printOk("Write File", args["path"].(string))
		printHint("──────────────────────────────────────────────────")
		content := strings.TrimSpace(args["content"].(string))
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			printHint(line)
		}
	case "run_command":
		printNormal("Run Command", args["command"].(string))
	// case "search_content":
	// 	fmt.Printf("[*] Search Content — \033[35m%s\033[0m\n", args["pattern"])
	// case "patch_edit":
	// 	fmt.Printf("[*] Patch Edit — \033[33m%s\033[0m\n", args["path"])
	// case "run_command":
	// 	fmt.Printf("[*] Run Command — \033[32m%s\033[0m\n", args["command"])
	// case "fetch_yahoo_finance":
	// 	fmt.Printf("[*] Fetch Ticker — \033[34m%s (%s)\033[0m\n", args["symbol"], args["range"])
	// case "fetch_google_rss":
	// 	fmt.Printf("[*] Fetch News — \033[34m%s (%s)\033[0m\n", args["keyword"], args["time"])
	// case "fetch_page":
	// 	url := fmt.Sprintf("%v", args["url"])
	// 	if len(url) > 64 {
	// 		url = url[:61] + "..."
	// 	}
	// 	fmt.Printf("[*] Fetch Page — \033[34m%s\033[0m\n", url)
	default:
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.Encode(args)
		fmt.Printf("[*] Tool: %s — \033[90m%s\033[0m\n", event.ToolName, strings.TrimSpace(buf.String()))
	}
}

func printOk(action, text string) {
	fmt.Printf("[*] %s — %s%s%s\n", strings.TrimSpace(action), colorOk, strings.TrimSpace(text), colorReset)
}

func printWarn(action, text string) {
	fmt.Printf("[~] %s — %s%s%s\n", strings.TrimSpace(action), colorWarn, strings.TrimSpace(text), colorReset)
}

func printError(action, text string) {
	fmt.Printf("[!] %s — %s%s%s\n", strings.TrimSpace(action), colorError, strings.TrimSpace(text), colorReset)
}

func printConfirm(action, text string) {
	fmt.Printf("[?] %s — %s%s%s\n", strings.TrimSpace(action), colorConfirm, strings.TrimSpace(text), colorReset)
}

func printNormal(action, text string) {
	fmt.Printf("[*] %s — %s%s%s\n", strings.TrimSpace(action), colorNormal, strings.TrimSpace(text), colorReset)
}

func printHint(text string) {
	fmt.Printf("%s%s%s\n", colorHint, strings.TrimSpace(text), colorReset)
}
