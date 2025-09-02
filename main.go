package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Constants for Claude Max 5x Plan Limits ($100/month)
const (
	OpusContextLimit = 200000  // 200K context window
	OpusDailyLimit   = 430000  // Daily token estimate for Max 5x
	OpusWeeklyLimit  = 3000000 // Weekly token estimate for Max 5x (7 days)
	OpusMessageLimit = 225     // Messages per 5 hour window
	OpusRateWindow   = 18000   // 5 hours in seconds (5 * 60 * 60)
	SecondsInWeek    = 604800  // 7 days in seconds (7 * 24 * 60 * 60)
)

// Claude pricing constants (per 1M tokens)
const (
	SonnetInputCost  = 3.00  // $3.00 per 1M input tokens
	SonnetOutputCost = 15.00 // $15.00 per 1M output tokens
	HaikuInputCost   = 0.25  // $0.25 per 1M input tokens
	HaikuOutputCost  = 1.25  // $1.25 per 1M output tokens
	OpusInputCost    = 15.00 // $15.00 per 1M input tokens
	OpusOutputCost   = 75.00 // $75.00 per 1M output tokens
)

// Powerline symbols
const (
	PowerlineRightArrow     = "\uE0B0" //
	PowerlineRightThinArrow = "\uE0B1" //
	PowerlineLeftArrow      = "\uE0B2" //
	PowerlineLeftThinArrow  = "\uE0B3" //
	GitBranch               = "\uE0A0" //
	GitIcon                 = "𖠰"
	BlockIcon               = "⏱"
	TokenIcon               = "🔤"
	TimerIcon               = "⏱"
	PercentIcon             = "%"
	DollarIcon              = "$"
	MessageIcon             = "💬"
	EfficiencyIcon          = "📊"
	LatencyIcon             = "⚡"
	CompactionIcon          = "🗜️"
	WeeklyIcon              = "📅"
	DailyIcon               = "📊"
)

// Enhanced ANSI color codes with truecolor support
const (
	ColorReset = "\033[0m"
	ColorBold  = "\033[1m"
	ColorDim   = "\033[2m"

	// Standard colors
	ColorBlack   = "\033[30m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"

	// Bright colors
	ColorBrightBlack   = "\033[90m"
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// Bright background colors
	BgBrightBlack   = "\033[100m"
	BgBrightRed     = "\033[101m"
	BgBrightGreen   = "\033[102m"
	BgBrightYellow  = "\033[103m"
	BgBrightBlue    = "\033[104m"
	BgBrightMagenta = "\033[105m"
	BgBrightCyan    = "\033[106m"
	BgBrightWhite   = "\033[107m"
)

// Theme configuration
type Theme struct {
	Name            string
	UserColor       string
	UserBg          string
	HostColor       string
	HostBg          string
	PathColor       string
	PathBg          string
	ModelColor      string
	ModelBg         string
	PercentColor    func(int) string
	PercentBg       func(int) string
	TokensColor     string
	TokensBg        string
	TimeColor       string
	TimeBg          string
	GitColor        string
	GitBg           string
	CostColor       string
	CostBg          string
	MessageColor    string
	MessageBg       string
	EfficiencyColor string
	EfficiencyBg    string
	LatencyColor    string
	LatencyBg       string
	CompactionColor func(int) string
	CompactionBg    func(int) string
	WeeklyColor     func(int) string
	WeeklyBg        func(int) string
	SeparatorColor  string
	UsePowerline    bool
}

// Predefined themes
var themes = map[string]Theme{
	"powerline": {
		Name:       "Powerline",
		UserColor:  ColorBrightWhite,
		UserBg:     BgBlue,
		HostColor:  ColorBrightWhite,
		HostBg:     BgBlue,
		PathColor:  ColorBlack,
		PathBg:     BgBrightCyan,
		ModelColor: ColorBrightWhite,
		ModelBg:    BgMagenta,
		PercentColor: func(p int) string {
			if p < 10 {
				return ColorBrightWhite
			}
			if p < 30 {
				return ColorBlack
			}
			return ColorBlack
		},
		PercentBg: func(p int) string {
			if p < 10 {
				return BgRed
			}
			if p < 30 {
				return BgYellow
			}
			return BgGreen
		},
		TokensColor:     ColorBrightWhite,
		TokensBg:        BgBrightBlack,
		TimeColor:       ColorBrightWhite,
		TimeBg:          BgBrightBlue,
		GitColor:        ColorBrightWhite,
		GitBg:           BgBrightGreen,
		CostColor:       ColorBrightWhite,
		CostBg:          BgRed,
		MessageColor:    ColorBrightWhite,
		MessageBg:       BgMagenta,
		EfficiencyColor: ColorBrightWhite,
		EfficiencyBg:    BgBrightBlue,
		LatencyColor:    ColorBrightWhite,
		LatencyBg:       BgBrightGreen,
		CompactionColor: func(p int) string {
			if p < 50 {
				return ColorBrightWhite
			}
			if p < 80 {
				return ColorBlack
			}
			return ColorBrightWhite
		},
		CompactionBg: func(p int) string {
			if p < 50 {
				return BgGreen
			}
			if p < 80 {
				return BgYellow
			}
			return BgRed
		},
		WeeklyColor: func(p int) string {
			if p < 60 {
				return ColorBrightWhite
			}
			if p < 85 {
				return ColorBlack
			}
			return ColorBrightWhite
		},
		WeeklyBg: func(p int) string {
			if p < 60 {
				return BgBrightBlue
			}
			if p < 85 {
				return BgYellow
			}
			return BgRed
		},
		SeparatorColor: ColorReset,
		UsePowerline:   true,
	},
	"minimal": {
		Name:       "Minimal",
		UserColor:  ColorBrightGreen,
		UserBg:     "",
		HostColor:  ColorBrightGreen,
		HostBg:     "",
		PathColor:  ColorBrightBlue,
		PathBg:     "",
		ModelColor: ColorBrightMagenta,
		ModelBg:    "",
		PercentColor: func(p int) string {
			if p < 10 {
				return ColorBrightRed
			}
			if p < 30 {
				return ColorBrightYellow
			}
			return ColorBrightGreen
		},
		PercentBg:       func(p int) string { return "" },
		TokensColor:     ColorBrightBlack,
		TokensBg:        "",
		TimeColor:       ColorBrightCyan,
		TimeBg:          "",
		GitColor:        ColorBrightYellow,
		GitBg:           "",
		CostColor:       ColorBrightRed,
		CostBg:          "",
		MessageColor:    ColorBrightMagenta,
		MessageBg:       "",
		EfficiencyColor: ColorBrightBlue,
		EfficiencyBg:    "",
		LatencyColor:    ColorBrightGreen,
		LatencyBg:       "",
		CompactionColor: func(p int) string {
			if p < 50 {
				return ColorBrightGreen
			}
			if p < 80 {
				return ColorBrightYellow
			}
			return ColorBrightRed
		},
		CompactionBg: func(p int) string { return "" },
		WeeklyColor: func(p int) string {
			if p < 60 {
				return ColorBrightBlue
			}
			if p < 85 {
				return ColorBrightYellow
			}
			return ColorBrightRed
		},
		WeeklyBg:       func(p int) string { return "" },
		SeparatorColor: ColorBrightBlack,
		UsePowerline:   false,
	},
	"gruvbox": {
		Name:       "Gruvbox",
		UserColor:  trueColor(254, 128, 25),  // orange
		UserBg:     trueColorBg(40, 40, 40),  // dark gray
		HostColor:  trueColor(184, 187, 38),  // yellow-green
		HostBg:     trueColorBg(60, 56, 54),  // gray
		PathColor:  trueColor(131, 165, 152), // aqua
		PathBg:     trueColorBg(80, 73, 69),  // darker gray
		ModelColor: trueColor(211, 134, 155), // purple
		ModelBg:    trueColorBg(102, 92, 84), // brown-gray
		PercentColor: func(p int) string {
			if p < 10 {
				return trueColor(251, 73, 52)
			} // red
			if p < 30 {
				return trueColor(250, 189, 47)
			} // yellow
			return trueColor(184, 187, 38) // green
		},
		PercentBg:       func(p int) string { return trueColorBg(60, 56, 54) },
		TokensColor:     trueColor(235, 219, 178), // light
		TokensBg:        trueColorBg(50, 48, 47),  // darker
		TimeColor:       trueColor(142, 192, 124), // bright green
		TimeBg:          trueColorBg(40, 40, 40),
		GitColor:        trueColor(254, 128, 25), // orange
		GitBg:           trueColorBg(60, 56, 54),
		CostColor:       trueColor(251, 73, 52), // red
		CostBg:          trueColorBg(40, 40, 40),
		MessageColor:    trueColor(211, 134, 155), // purple
		MessageBg:       trueColorBg(60, 56, 54),
		EfficiencyColor: trueColor(131, 165, 152), // aqua
		EfficiencyBg:    trueColorBg(80, 73, 69),
		LatencyColor:    trueColor(142, 192, 124), // bright green
		LatencyBg:       trueColorBg(50, 48, 47),
		CompactionColor: func(p int) string {
			if p < 50 {
				return trueColor(142, 192, 124)
			} // bright green
			if p < 80 {
				return trueColor(250, 189, 47)
			} // yellow
			return trueColor(251, 73, 52) // red
		},
		CompactionBg: func(p int) string { return trueColorBg(60, 56, 54) },
		WeeklyColor: func(p int) string {
			if p < 60 {
				return trueColor(131, 165, 152)
			} // aqua
			if p < 85 {
				return trueColor(250, 189, 47)
			} // yellow
			return trueColor(251, 73, 52) // red
		},
		WeeklyBg:       func(p int) string { return trueColorBg(60, 56, 54) },
		SeparatorColor: trueColor(80, 73, 69),
		UsePowerline:   true,
	},
}

// ModelInfo represents the model information from Claude Code
type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// WorkspaceInfo represents the workspace information from Claude Code
type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// ContextUsage represents context usage information
type ContextUsage struct {
	Characters int `json:"characters"`
	Tokens     int `json:"tokens"`
}

// CostData represents cost information
type CostData struct {
	SessionCost float64 `json:"sessionCost"`
	DailyCost   float64 `json:"dailyCost"`
}

// StatusLineInput represents the JSON input structure from Claude Code
type StatusLineInput struct {
	Model              ModelInfo     `json:"model"`
	Workspace          WorkspaceInfo `json:"workspace"`
	WorkspaceDirectory string        `json:"workspaceDirectory"` // Alternative field
	Usage              *UsageInfo    `json:"usage,omitempty"`
	InputTokens        int           `json:"inputTokens,omitempty"`
	OutputTokens       int           `json:"outputTokens,omitempty"`
	TotalTokens        int           `json:"totalTokens,omitempty"`
	ContextUsage       *ContextUsage `json:"contextUsage,omitempty"`
	Context            *ContextUsage `json:"context,omitempty"`
	CostData           *CostData     `json:"costData,omitempty"`
	SessionCost        float64       `json:"sessionCost,omitempty"`
	DailyCost          float64       `json:"dailyCost,omitempty"`
}

// CCUsageData represents parsed ccusage output
type CCUsageData struct {
	SessionTokens int
	DailyTokens   int
	WeeklyTokens  int
	Messages      int
	InputTokens   int
	OutputTokens  int
	SessionID     string
}

// CalculatedUsage represents calculated usage from conversation history
type CalculatedUsage struct {
	SessionTokens int
	DailyTokens   int
	WeeklyTokens  int
	Messages      int
	InputTokens   int
	OutputTokens  int
}

// LatencyData represents request latency tracking
type LatencyData struct {
	AverageMs     float64
	LastRequestMs float64
	RequestCount  int
}

// Widget represents a status line widget
type Widget struct {
	Name    string
	Content string
	Color   string
	BgColor string
}

// StatusLine holds the complete status line configuration
type StatusLine struct {
	Theme     Theme
	Widgets   []Widget
	StartTime time.Time
}

func main() {
	// Read JSON input from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON input
	var statusInput StatusLineInput
	if err := json.Unmarshal(input, &statusInput); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Initialize status line with theme (default: powerline)
	themeName := os.Getenv("CCSTATUS_THEME")
	if themeName == "" {
		themeName = "powerline"
	}

	theme, exists := themes[themeName]
	if !exists {
		theme = themes["powerline"]
	}

	// Create status line
	statusLine := &StatusLine{
		Theme:     theme,
		StartTime: time.Now(), // This would be session start in real implementation
	}

	// Generate enhanced status line
	output := statusLine.generatePowerlineStatusLine(statusInput)

	// Output status line to stdout
	fmt.Println(output)
}

// generatePowerlineStatusLine creates a powerline-style status line
func (s *StatusLine) generatePowerlineStatusLine(input StatusLineInput) string {
	// Collect data
	ccusageData := getCCUsageData()
	calculatedUsage := getCalculatedUsage()

	// Extract token usage from JSON input
	inputTokens := getInputTokens(input)
	outputTokens := getOutputTokens(input)
	totalTokens := getTotalTokens(input)
	contextTokens := getContextTokens(input)
	contextChars := getContextCharacters(input)

	// Calculate actual token usage (prefer calculated data, then ccusage, then JSON)
	var dailyTokensUsed int
	var sessionInputTokens, sessionOutputTokens int

	if calculatedUsage.DailyTokens > 0 {
		dailyTokensUsed = calculatedUsage.DailyTokens
		sessionInputTokens = calculatedUsage.InputTokens
		sessionOutputTokens = calculatedUsage.OutputTokens
	} else if ccusageData.DailyTokens > 0 {
		dailyTokensUsed = ccusageData.DailyTokens
		sessionInputTokens = ccusageData.InputTokens
		sessionOutputTokens = ccusageData.OutputTokens
	} else if totalTokens > 0 {
		dailyTokensUsed = totalTokens
		sessionInputTokens = inputTokens
		sessionOutputTokens = outputTokens
	} else {
		dailyTokensUsed = inputTokens + outputTokens
		sessionInputTokens = inputTokens
		sessionOutputTokens = outputTokens
	}

	// Build widgets
	s.Widgets = []Widget{}

	// User@Host widget
	username := getUsername()
	hostname := getHostname()
	s.addWidget("user", fmt.Sprintf("%s@%s", username, hostname), s.Theme.UserColor, s.Theme.UserBg)

	// Path widget
	workspacePath := formatWorkspacePath(getWorkspacePath(input))
	pathDisplay := truncatePath(workspacePath, 30)
	s.addWidget("path", pathDisplay, s.Theme.PathColor, s.Theme.PathBg)

	// Git widget
	if gitInfo := getGitInfo(getWorkspacePath(input)); gitInfo != "" {
		s.addWidget("git", gitInfo, s.Theme.GitColor, s.Theme.GitBg)
	}

	// Model widget
	modelDisplay := getModelDisplay(input.Model)
	s.addWidget("model", modelDisplay, s.Theme.ModelColor, s.Theme.ModelBg)

	// Usage percentage widget (daily)
	usagePercent := calculateUsagePercentage(dailyTokensUsed, contextTokens, contextChars)
	remainingPercent := 100 - usagePercent
	if remainingPercent < 0 {
		remainingPercent = 0 // Don't show negative percentages
	}
	s.addWidget("percent", fmt.Sprintf("%d%%", remainingPercent),
		s.Theme.PercentColor(remainingPercent), s.Theme.PercentBg(remainingPercent))

	// Weekly vs Daily comparison widget
	weeklyTokensUsed := getWeeklyTokensUsed(ccusageData, calculatedUsage)
	if weeklyTokensUsed > 0 || dailyTokensUsed > 0 {
		dailyPercent := calculateDailyUsagePercentage(dailyTokensUsed)
		weeklyPercent := calculateWeeklyUsagePercentage(weeklyTokensUsed)

		// Show the more restrictive limit (higher percentage)
		if weeklyPercent > dailyPercent && weeklyPercent > 0 {
			s.addWidget("weekly", fmt.Sprintf("%s %d%%", WeeklyIcon, weeklyPercent),
				s.Theme.WeeklyColor(weeklyPercent), s.Theme.WeeklyBg(weeklyPercent))
		} else if dailyPercent > 0 {
			s.addWidget("daily", fmt.Sprintf("%s %d%%", DailyIcon, dailyPercent),
				s.Theme.WeeklyColor(dailyPercent), s.Theme.WeeklyBg(dailyPercent))
		}
	}

	// Token usage widget
	if dailyTokensUsed > 0 {
		tokensDisplay := formatTokensAdvanced(dailyTokensUsed)
		s.addWidget("tokens", fmt.Sprintf("%s %s", TokenIcon, tokensDisplay),
			s.Theme.TokensColor, s.Theme.TokensBg)
	}

	// Cost widget - show session cost
	if sessionInputTokens > 0 || sessionOutputTokens > 0 {
		sessionCost, _ := calculateCost(input.Model.DisplayName, sessionInputTokens, sessionOutputTokens)
		costDisplay := formatCost(sessionCost)
		s.addWidget("cost", fmt.Sprintf("%s %s", DollarIcon, costDisplay),
			s.Theme.CostColor, s.Theme.CostBg)
	}

	// Message count widget
	messageCount := getMessageCount(ccusageData, calculatedUsage)
	if messageCount > 0 {
		s.addWidget("messages", fmt.Sprintf("%s %d/%d", MessageIcon, messageCount, OpusMessageLimit),
			s.Theme.MessageColor, s.Theme.MessageBg)
	}

	// Context efficiency widget
	if contextTokens > 0 {
		efficiency := calculateContextEfficiency(contextTokens)
		efficiencyDisplay := formatEfficiency(efficiency)
		s.addWidget("efficiency", fmt.Sprintf("%s %s", EfficiencyIcon, efficiencyDisplay),
			s.Theme.EfficiencyColor, s.Theme.EfficiencyBg)
	}

	// Message compaction percentage widget
	if contextTokens > 0 {
		compactionPercent := calculateCompactionPercentage(contextTokens)
		s.addWidget("compaction", fmt.Sprintf("%s %d%%", CompactionIcon, compactionPercent),
			s.Theme.CompactionColor(compactionPercent), s.Theme.CompactionBg(compactionPercent))
	}

	// Request latency widget - removed as it's not useful

	// Block timer widget
	blockTime := getBlockTimerDisplay()
	if blockTime != "" {
		s.addWidget("timer", fmt.Sprintf("%s %s", BlockIcon, blockTime),
			s.Theme.TimeColor, s.Theme.TimeBg)
	}

	// Time to reset widget - show both 5hr and weekly
	timeToReset, resetType := calculateTimeToReset()
	weeklyTimeToReset, weeklyResetType := calculateTimeToWeeklyReset()

	// Show whichever reset is sooner or more relevant
	if resetType == "5hr" && timeToReset != "0m" {
		s.addWidget("reset", fmt.Sprintf("%s reset %s", resetType, timeToReset),
			s.Theme.TimeColor, s.Theme.TimeBg)
	} else {
		// Show weekly if 5hr window has expired or is unknown
		s.addWidget("reset", fmt.Sprintf("%s reset %s", weeklyResetType, weeklyTimeToReset),
			s.Theme.TimeColor, s.Theme.TimeBg)
	}

	// Render widgets with powerline separators
	return s.renderPowerline()
}

// addWidget adds a widget to the status line
func (s *StatusLine) addWidget(name, content, color, bgColor string) {
	s.Widgets = append(s.Widgets, Widget{
		Name:    name,
		Content: content,
		Color:   color,
		BgColor: bgColor,
	})
}

// renderPowerline renders widgets with powerline-style separators
func (s *StatusLine) renderPowerline() string {
	if len(s.Widgets) == 0 {
		return ""
	}

	var parts []string

	for i, widget := range s.Widgets {
		// Widget content with colors
		var segment string
		if s.Theme.UsePowerline && widget.BgColor != "" {
			// Powerline segment with background
			segment = fmt.Sprintf("%s%s %s %s", widget.BgColor, widget.Color, widget.Content, ColorReset)
		} else {
			// Simple colored text
			segment = fmt.Sprintf("%s%s%s", widget.Color, widget.Content, ColorReset)
		}

		parts = append(parts, segment)

		// Add separator (except for last widget)
		if i < len(s.Widgets)-1 {
			nextWidget := s.Widgets[i+1]
			separator := s.getSeparator(widget, nextWidget)
			parts = append(parts, separator)
		}
	}

	return strings.Join(parts, "")
}

// getSeparator returns appropriate separator between two widgets
func (s *StatusLine) getSeparator(current, next Widget) string {
	if !s.Theme.UsePowerline {
		return fmt.Sprintf(" %s|%s ", s.Theme.SeparatorColor, ColorReset)
	}

	// Powerline arrow separator
	if current.BgColor != "" && next.BgColor != "" {
		// Background to background transition
		return fmt.Sprintf("%s%s%s%s", next.BgColor, getBgToFgColor(current.BgColor), PowerlineRightArrow, ColorReset)
	} else if current.BgColor != "" && next.BgColor == "" {
		// Background to normal transition
		return fmt.Sprintf("%s%s%s", getBgToFgColor(current.BgColor), PowerlineRightArrow, ColorReset)
	} else {
		// Normal to normal transition
		return fmt.Sprintf(" %s%s%s ", s.Theme.SeparatorColor, PowerlineRightThinArrow, ColorReset)
	}
}

// Helper functions for advanced features

// trueColor returns truecolor ANSI sequence
func trueColor(r, g, b int) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// trueColorBg returns truecolor background ANSI sequence
func trueColorBg(r, g, b int) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// getBgToFgColor converts background color code to foreground
func getBgToFgColor(bgColor string) string {
	// Simple mapping - in practice would need full color translation
	switch bgColor {
	case BgBlue:
		return ColorBlue
	case BgBrightCyan:
		return ColorBrightCyan
	case BgMagenta:
		return ColorMagenta
	case BgBrightBlack:
		return ColorBrightBlack
	case BgRed:
		return ColorRed
	case BgYellow:
		return ColorYellow
	case BgGreen:
		return ColorGreen
	case BgBrightBlue:
		return ColorBrightBlue
	case BgBrightGreen:
		return ColorBrightGreen
	default:
		return ColorWhite
	}
}

// calculateUsagePercentage calculates usage percentage from various sources
func calculateUsagePercentage(dailyTokens, contextTokens, contextChars int) int {
	contextPercentage := 0
	if contextTokens > 0 {
		contextPercentage = int((contextTokens * 100) / OpusContextLimit)
	} else if contextChars > 0 {
		estimatedTokens := contextChars / 4
		contextPercentage = int((estimatedTokens * 100) / OpusContextLimit)
	}
	
	// Cap context percentage at 100%
	if contextPercentage > 100 {
		contextPercentage = 100
	}

	dailyPercentage := int((dailyTokens * 100) / OpusDailyLimit)
	
	// Cap daily percentage at 100%
	if dailyPercentage > 100 {
		dailyPercentage = 100
	}

	if contextPercentage > dailyPercentage {
		return contextPercentage
	}
	return dailyPercentage
}

// getBlockTimerDisplay returns the block timer display
func getBlockTimerDisplay() string {
	// Mock implementation - would calculate from session start
	now := time.Now()
	// Simulate block start (5-hour windows from midnight)
	hour := now.Hour()
	blockStart := (hour / 5) * 5
	blockStartTime := time.Date(now.Year(), now.Month(), now.Day(), blockStart, 0, 0, 0, now.Location())

	elapsed := now.Sub(blockStartTime)
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60

	if hours == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// formatTokensAdvanced formats token counts with advanced display
func formatTokensAdvanced(tokens int) string {
	if tokens > 1000000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
	} else if tokens > 1000 {
		return fmt.Sprintf("%.1fk", float64(tokens)/1000)
	}
	return strconv.Itoa(tokens)
}

// truncatePath truncates path to fit within maxLength
func truncatePath(path string, maxLength int) string {
	if len(path) <= maxLength {
		return path
	}

	// Try to show end of path with ellipsis
	if maxLength > 5 {
		return "…" + path[len(path)-(maxLength-1):]
	}
	return path[:maxLength]
}

// calculateCost calculates session and daily costs based on model and token usage
func calculateCost(modelName string, inputTokens, outputTokens int) (sessionCost, dailyCost float64) {
	var inputCostPer1M, outputCostPer1M float64

	modelLower := strings.ToLower(modelName)
	switch {
	case strings.Contains(modelLower, "sonnet"):
		inputCostPer1M = SonnetInputCost
		outputCostPer1M = SonnetOutputCost
	case strings.Contains(modelLower, "haiku"):
		inputCostPer1M = HaikuInputCost
		outputCostPer1M = HaikuOutputCost
	case strings.Contains(modelLower, "opus"):
		inputCostPer1M = OpusInputCost
		outputCostPer1M = OpusOutputCost
	default:
		inputCostPer1M = SonnetInputCost
		outputCostPer1M = SonnetOutputCost
	}

	sessionCost = (float64(inputTokens)*inputCostPer1M + float64(outputTokens)*outputCostPer1M) / 1000000
	dailyCost = sessionCost

	return sessionCost, dailyCost
}

// getMessageCount gets the current message count in rate limit window
func getMessageCount(ccusageData CCUsageData, calculatedUsage CalculatedUsage) int {
	if calculatedUsage.Messages > 0 {
		return calculatedUsage.Messages
	}
	if ccusageData.Messages > 0 {
		return ccusageData.Messages
	}
	return 0
}

// calculateContextEfficiency calculates context window utilization percentage
func calculateContextEfficiency(contextTokens int) float64 {
	if contextTokens == 0 {
		return 0.0
	}
	efficiency := (float64(contextTokens) / float64(OpusContextLimit)) * 100
	if efficiency > 100.0 {
		efficiency = 100.0 // Cap at 100%
	}
	return efficiency
}

// getLatencyData retrieves request latency information
func getLatencyData() LatencyData {
	var latency LatencyData

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return latency
	}

	latencyPath := filepath.Join(homeDir, ".claude", "latency.txt")
	if content, err := os.ReadFile(latencyPath); err == nil {
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		if len(lines) >= 3 {
			if avg, err := strconv.ParseFloat(strings.TrimSpace(lines[0]), 64); err == nil {
				latency.AverageMs = avg
			}
			if last, err := strconv.ParseFloat(strings.TrimSpace(lines[1]), 64); err == nil {
				latency.LastRequestMs = last
			}
			if count, err := strconv.Atoi(strings.TrimSpace(lines[2])); err == nil {
				latency.RequestCount = count
			}
		}
	}

	return latency
}

// formatCost formats cost values for display
func formatCost(cost float64) string {
	if cost < 0.01 {
		return fmt.Sprintf("%.3f¢", cost*100)
	} else if cost < 1.0 {
		return fmt.Sprintf("%.2f¢", cost*100)
	}
	return fmt.Sprintf("$%.2f", cost)
}

// formatEfficiency formats efficiency percentage for display
func formatEfficiency(efficiency float64) string {
	return fmt.Sprintf("%.1f%%", efficiency)
}

// formatLatency formats latency for display
func formatLatency(latencyMs float64) string {
	if latencyMs < 1000 {
		return fmt.Sprintf("%.0fms", latencyMs)
	}
	return fmt.Sprintf("%.1fs", latencyMs/1000)
}

// calculateWeeklyUsagePercentage calculates percentage of weekly limit used
func calculateWeeklyUsagePercentage(weeklyTokens int) int {
	if weeklyTokens == 0 {
		return 0
	}
	percentage := int((float64(weeklyTokens) / float64(OpusWeeklyLimit)) * 100)
	if percentage > 100 {
		percentage = 100 // Cap at 100%
	}
	return percentage
}

// calculateDailyUsagePercentage calculates percentage of daily limit used
func calculateDailyUsagePercentage(dailyTokens int) int {
	if dailyTokens == 0 {
		return 0
	}
	percentage := int((float64(dailyTokens) / float64(OpusDailyLimit)) * 100)
	if percentage > 100 {
		percentage = 100 // Cap at 100%
	}
	return percentage
}

// getWeeklyTokensUsed gets weekly token usage with fallback logic
func getWeeklyTokensUsed(ccusageData CCUsageData, calculatedUsage CalculatedUsage) int {
	if calculatedUsage.WeeklyTokens > 0 {
		return calculatedUsage.WeeklyTokens
	}
	if ccusageData.WeeklyTokens > 0 {
		return ccusageData.WeeklyTokens
	}
	// Fallback: estimate from daily (assume 7-day average)
	if ccusageData.DailyTokens > 0 {
		return ccusageData.DailyTokens * 7 / 7 // This would need better logic
	}
	return 0
}

// calculateTimeToWeeklyReset calculates time until weekly limit resets
func calculateTimeToWeeklyReset() (string, string) {
	now := time.Now()

	// Weekly resets happen every Monday at 00:00 UTC (based on Claude's implementation)
	// Find next Monday
	daysUntilMonday := (7 - int(now.Weekday()) + 1) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7 // If today is Monday, next reset is next Monday
	}

	nextMonday := now.AddDate(0, 0, daysUntilMonday)
	nextMondayMidnight := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 0, 0, 0, 0, time.UTC)

	duration := nextMondayMidnight.Sub(now)

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	var timeStr string
	if days > 0 {
		timeStr = fmt.Sprintf("%dd %dh", days, hours)
	} else if hours > 0 {
		timeStr = fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		timeStr = fmt.Sprintf("%dm", minutes)
	}

	return timeStr, "weekly"
}

// calculateCompactionPercentage calculates how close we are to hitting compaction
func calculateCompactionPercentage(contextTokens int) int {
	if contextTokens == 0 {
		return 0
	}

	// Most models start compaction around 90% of context limit
	compactionThreshold := int(float64(OpusContextLimit) * 0.9) // 180K tokens for 200K limit

	if contextTokens >= compactionThreshold {
		// We're in the danger zone - show percentage from threshold to limit
		remaining := OpusContextLimit - contextTokens
		dangerZone := OpusContextLimit - compactionThreshold
		percentage := int((float64(dangerZone-remaining) / float64(dangerZone)) * 100)
		if percentage > 100 {
			return 100
		}
		return percentage
	}

	// Show percentage to threshold
	return int((float64(contextTokens) / float64(compactionThreshold)) * 100)
}

// Enhanced git information with more details
func getGitInfo(dir string) string {
	gitDir := findGitDir(dir)
	if gitDir == "" {
		return ""
	}

	// Get branch name
	headFile := filepath.Join(gitDir, "HEAD")
	content, err := os.ReadFile(headFile)
	if err != nil {
		return ""
	}

	headContent := strings.TrimSpace(string(content))
	var branch string

	if strings.HasPrefix(headContent, "ref: refs/heads/") {
		branch = strings.TrimPrefix(headContent, "ref: refs/heads/")
	} else if len(headContent) >= 7 {
		branch = headContent[:7] // Detached HEAD
	} else {
		return ""
	}

	// Check for changes (simplified)
	changes := getGitChanges(dir)
	if changes > 0 {
		return fmt.Sprintf("%s %s±%d", GitBranch, branch, changes)
	}

	return fmt.Sprintf("%s %s", GitBranch, branch)
}

// findGitDir finds the .git directory
func findGitDir(startDir string) string {
	dir := startDir
	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil {
			if info.IsDir() {
				return gitDir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir || parent == "/" {
			break
		}
		dir = parent
	}
	return ""
}

// getGitChanges gets count of git changes (simplified implementation)
func getGitChanges(dir string) int {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

// Existing helper functions (kept from original implementation)
func getCCUsageData() CCUsageData {
	var data CCUsageData

	if _, err := exec.LookPath("ccusage"); err != nil {
		return data
	}

	// Try blocks command first for current 5-hour session data
	cmd := exec.Command("ccusage", "blocks", "--json")
	if output, err := cmd.Output(); err == nil {
		outputStr := string(output)
		// Parse current block data (most recent/active block)
		data.SessionTokens = extractTokenCount(outputStr, `"tokens"\s*:\s*(\d+)|"totalTokens"\s*:\s*(\d+)`)
		data.InputTokens = extractTokenCount(outputStr, `"inputTokens"\s*:\s*(\d+)`)
		data.OutputTokens = extractTokenCount(outputStr, `"outputTokens"\s*:\s*(\d+)`)
		data.Messages = extractTokenCount(outputStr, `"messages"\s*:\s*(\d+)|"messageCount"\s*:\s*(\d+)`)
	}

	// Get session-specific data if we have a session ID
	sessionID := getCurrentSessionID()
	if sessionID != "" {
		data.SessionID = sessionID
		cmd := exec.Command("ccusage", "session", sessionID, "--json")
		if output, err := cmd.Output(); err == nil {
			outputStr := string(output)
			// Override with session-specific data if available
			if sessionTokens := extractTokenCount(outputStr, `"tokens"\s*:\s*(\d+)|"totalTokens"\s*:\s*(\d+)`); sessionTokens > 0 {
				data.SessionTokens = sessionTokens
			}
			if inputTokens := extractTokenCount(outputStr, `"inputTokens"\s*:\s*(\d+)`); inputTokens > 0 {
				data.InputTokens = inputTokens
			}
			if outputTokens := extractTokenCount(outputStr, `"outputTokens"\s*:\s*(\d+)`); outputTokens > 0 {
				data.OutputTokens = outputTokens
			}
			if messages := extractTokenCount(outputStr, `"messages"\s*:\s*(\d+)|"messageCount"\s*:\s*(\d+)`); messages > 0 {
				data.Messages = messages
			}
		}
	}

	// Get overall daily stats for daily totals
	cmd = exec.Command("ccusage", "stats", "--json")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to plain text output
		cmd = exec.Command("ccusage", "stats")
		output, err = cmd.Output()
		if err != nil {
			return data
		}
	}

	outputStr := string(output)
	data.DailyTokens = extractTokenCount(outputStr, `"totalTokens"\s*:\s*(\d+)|total.*tokens.*:\s*(\d+)`)
	data.WeeklyTokens = extractTokenCount(outputStr, `"weeklyTokens"\s*:\s*(\d+)|weekly.*tokens.*:\s*(\d+)`)

	// If we still don't have session data, try to extract from general stats
	if data.SessionTokens == 0 {
		data.SessionTokens = extractTokenCount(outputStr, `"sessionTokens"\s*:\s*(\d+)|session.*tokens.*:\s*(\d+)`)
	}
	if data.InputTokens == 0 {
		data.InputTokens = extractTokenCount(outputStr, `"inputTokens"\s*:\s*(\d+)|input.*tokens.*:\s*(\d+)`)
	}
	if data.OutputTokens == 0 {
		data.OutputTokens = extractTokenCount(outputStr, `"outputTokens"\s*:\s*(\d+)|output.*tokens.*:\s*(\d+)`)
	}
	if data.Messages == 0 {
		data.Messages = extractTokenCount(outputStr, `"messages"\s*:\s*(\d+)|message.*count.*:\s*(\d+)`)
	}

	return data
}

// getCurrentSessionID attempts to get the current Claude Code session ID
func getCurrentSessionID() string {
	// Try multiple methods to get session ID

	// Method 1: Check environment variable
	if sessionID := os.Getenv("CLAUDE_SESSION_ID"); sessionID != "" {
		return sessionID
	}

	// Method 2: Check for session file in ~/.claude/
	homeDir, err := os.UserHomeDir()
	if err == nil {
		sessionFile := filepath.Join(homeDir, ".claude", "current_session")
		if content, err := os.ReadFile(sessionFile); err == nil {
			return strings.TrimSpace(string(content))
		}
	}

	// Method 3: Try to extract from Claude Code process info
	if sessionID := extractSessionFromProcess(); sessionID != "" {
		return sessionID
	}

	return ""
}

// extractSessionFromProcess tries to extract session ID from running Claude Code processes
func extractSessionFromProcess() string {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "claude") && strings.Contains(line, "session") {
			// Try to extract session ID from command line
			re := regexp.MustCompile(`session[=:]([a-zA-Z0-9-]+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				return matches[1]
			}
		}
	}

	return ""
}

func getCalculatedUsage() CalculatedUsage {
	var usage CalculatedUsage

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return usage
	}

	scriptPath := filepath.Join(homeDir, ".claude", "calculate-usage.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return usage
	}

	cmd := exec.Command(scriptPath)
	output, err := cmd.Output()
	if err != nil {
		return usage
	}

	parts := strings.Fields(string(output))
	if len(parts) >= 5 {
		if val, err := strconv.Atoi(parts[0]); err == nil {
			usage.SessionTokens = val
		}
		if val, err := strconv.Atoi(parts[1]); err == nil {
			usage.DailyTokens = val
		}
		if val, err := strconv.Atoi(parts[2]); err == nil {
			usage.Messages = val
		}
		if val, err := strconv.Atoi(parts[3]); err == nil {
			usage.InputTokens = val
		}
		if val, err := strconv.Atoi(parts[4]); err == nil {
			usage.OutputTokens = val
		}
	} else if len(parts) >= 3 {
		if val, err := strconv.Atoi(parts[0]); err == nil {
			usage.SessionTokens = val
		}
		if val, err := strconv.Atoi(parts[1]); err == nil {
			usage.DailyTokens = val
		}
		if val, err := strconv.Atoi(parts[2]); err == nil {
			usage.Messages = val
		}
	}

	return usage
}

func extractTokenCount(text, pattern string) int {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	for i := 1; i < len(matches); i++ {
		if matches[i] != "" {
			if count, err := strconv.Atoi(matches[i]); err == nil {
				return count
			}
		}
	}
	return 0
}

func getInputTokens(input StatusLineInput) int {
	if input.Usage != nil && input.Usage.InputTokens > 0 {
		return input.Usage.InputTokens
	}
	return input.InputTokens
}

func getOutputTokens(input StatusLineInput) int {
	if input.Usage != nil && input.Usage.OutputTokens > 0 {
		return input.Usage.OutputTokens
	}
	return input.OutputTokens
}

func getTotalTokens(input StatusLineInput) int {
	if input.Usage != nil && input.Usage.TotalTokens > 0 {
		return input.Usage.TotalTokens
	}
	return input.TotalTokens
}

func getContextTokens(input StatusLineInput) int {
	if input.ContextUsage != nil && input.ContextUsage.Tokens > 0 {
		return input.ContextUsage.Tokens
	}
	if input.Context != nil && input.Context.Tokens > 0 {
		return input.Context.Tokens
	}
	return 0
}

func getContextCharacters(input StatusLineInput) int {
	if input.ContextUsage != nil && input.ContextUsage.Characters > 0 {
		return input.ContextUsage.Characters
	}
	if input.Context != nil && input.Context.Characters > 0 {
		return input.Context.Characters
	}
	return 0
}

func calculateTimeToReset() (string, string) {
	// Claude uses 5-hour rolling windows, not fixed daily resets
	// The window starts with your first prompt and resets 5 hours later

	// Try to get actual session start time from ccusage or estimate
	sessionStartTime := getSessionStartTime()
	now := time.Now()

	if !sessionStartTime.IsZero() {
		// Calculate time since session started
		elapsed := now.Sub(sessionStartTime)
		fiveHours := 5 * time.Hour

		if elapsed < fiveHours {
			// Still in current 5-hour window
			remaining := fiveHours - elapsed
			hours := int(remaining.Hours())
			minutes := int(remaining.Minutes()) % 60

			var timeStr string
			if hours > 0 {
				timeStr = fmt.Sprintf("%dh %dm", hours, minutes)
			} else {
				timeStr = fmt.Sprintf("%dm", minutes)
			}
			return timeStr, "5hr"
		} else {
			// Session has expired, next message starts new window
			return "0m", "5hr"
		}
	}

	// Fallback: estimate based on daily reset (if no session data available)
	currentHour := now.Hour()
	currentMin := now.Minute()
	currentSec := now.Second()

	secondsSinceMidnight := currentHour*3600 + currentMin*60 + currentSec
	secondsToDailyReset := 86400 - secondsSinceMidnight

	hoursToReset := secondsToDailyReset / 3600
	minutesToReset := (secondsToDailyReset % 3600) / 60

	var timeStr string
	if hoursToReset > 0 {
		timeStr = fmt.Sprintf("%dh %dm", hoursToReset, minutesToReset)
	} else {
		timeStr = fmt.Sprintf("%dm", minutesToReset)
	}

	return timeStr, "daily"
}

// getSessionStartTime tries to determine when the current 5-hour session started
func getSessionStartTime() time.Time {
	// Try to get session start from ccusage blocks command
	if _, err := exec.LookPath("ccusage"); err == nil {
		cmd := exec.Command("ccusage", "blocks", "--json")
		if output, err := cmd.Output(); err == nil {
			// Parse JSON to find current active block start time
			outputStr := string(output)
			if startTime := extractSessionStartFromCCUsage(outputStr); !startTime.IsZero() {
				return startTime
			}
		}
	}

	// Fallback: check for session start time file
	homeDir, err := os.UserHomeDir()
	if err == nil {
		sessionStartFile := filepath.Join(homeDir, ".claude", "session_start")
		if content, err := os.ReadFile(sessionStartFile); err == nil {
			if startTime, err := time.Parse(time.RFC3339, strings.TrimSpace(string(content))); err == nil {
				return startTime
			}
		}
	}

	return time.Time{} // Zero time if no session data found
}

// extractSessionStartFromCCUsage parses ccusage blocks output to find active session start
func extractSessionStartFromCCUsage(jsonOutput string) time.Time {
	// Look for patterns like "start_time": "2024-01-01T10:00:00Z" in current block
	re := regexp.MustCompile(`"start_time"\s*:\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(jsonOutput)
	if len(matches) > 1 {
		if startTime, err := time.Parse(time.RFC3339, matches[1]); err == nil {
			// Check if this session is still active (within 5 hours)
			if time.Since(startTime) < 5*time.Hour {
				return startTime
			}
		}
	}
	return time.Time{}
}

func getModelDisplay(model ModelInfo) string {
	modelStr := model.DisplayName
	if modelStr == "" {
		modelStr = model.ID
	}

	if strings.Contains(strings.ToLower(modelStr), "sonnet") || strings.Contains(modelStr, "3.5") {
		return "opus"
	}

	return strings.ToLower(modelStr)
}

func getUsername() string {
	if currentUser, err := user.Current(); err == nil {
		return currentUser.Username
	}
	return "user"
}

func getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		if dotIndex := strings.Index(hostname, "."); dotIndex > 0 {
			return hostname[:dotIndex]
		}
		return hostname
	}
	return "localhost"
}

func getWorkspacePath(input StatusLineInput) string {
	if input.Workspace.CurrentDir != "" {
		return input.Workspace.CurrentDir
	}
	if input.WorkspaceDirectory != "" {
		return input.WorkspaceDirectory
	}
	return "~"
}

func formatWorkspacePath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}
	return path
}
