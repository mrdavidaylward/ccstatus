# CCStatus - Enhanced Claude Code Status Line

A powerful Go-based status line for Claude Code with Powerline styling, multiple themes, and advanced metrics.

## Features

🎨 **Multiple Themes**
- `powerline` - Full Powerline theme with backgrounds and arrows
- `minimal` - Clean theme with simple separators
- `gruvbox` - Beautiful truecolor theme with warm colors

⚡ **Powerline Support**
- Arrow separators (``) for seamless visual flow
- Background colors for distinct segments
- Git branch symbols (``) and icons

📊 **Advanced Metrics (2025 Claude Pro Limits)**
- 💰 **Cost tracking** - Session costs based on token usage and model pricing
- 💬 **Message count** - Tracks ~45 messages per 5-hour rolling window
- 📊 **Context efficiency** - Context window utilization (200K standard, 1M available)
- 🗜️ **Compaction warning** - Percentage until message compaction threshold
- 📅 **Weekly limits** - New August 2025 weekly rate limits (40-80 hours Sonnet 4)
- ⏱ **5-hour rolling windows** - Accurate reset timers per Claude Pro rate limits
- Git branch with change count (`master±3`)

🔧 **Smart Integration**
- Enhanced `ccusage` CLI tool integration with session tracking
- 5-hour rolling window calculations aligned with Claude Pro limits
- Context window usage tracking (200K standard / 1M extended)
- Weekly limit tracking (resets every 7 days)
- Multiple data source fallbacks for reliability

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/mrdavidaylward/ccstatus/main/install.sh | bash
```

## Manual Installation

### Option 1: Download Pre-built Binary

Download the latest release for your platform:
- [Linux (x64)](https://github.com/mrdavidaylward/ccstatus/releases/latest/download/ccstatus_linux_amd64.tar.gz)
- [Linux (ARM64)](https://github.com/mrdavidaylward/ccstatus/releases/latest/download/ccstatus_linux_arm64.tar.gz)
- [macOS (Intel)](https://github.com/mrdavidaylward/ccstatus/releases/latest/download/ccstatus_darwin_amd64.tar.gz)
- [macOS (Apple Silicon)](https://github.com/mrdavidaylward/ccstatus/releases/latest/download/ccstatus_darwin_arm64.tar.gz)
- [Windows (x64)](https://github.com/mrdavidaylward/ccstatus/releases/latest/download/ccstatus_windows_amd64.zip)

### Option 2: Build from Source

```bash
git clone https://github.com/mrdavidaylward/ccstatus.git
cd ccstatus
go build -o ccstatus main.go
./install.sh
```

### Configure Claude Code

The install script automatically configures Claude Code, but you can also do it manually:

```json
{
  "statusLine": {
    "type": "command", 
    "command": "/usr/local/bin/ccstatus"
  }
}
```

## Usage

### Theme Selection
```bash
# Set theme via environment variable
export CCSTATUS_THEME=powerline  # Default
export CCSTATUS_THEME=minimal
export CCSTATUS_THEME=gruvbox

# Test themes
echo '{"model":{"display_name":"Sonnet 4"},"workspace":{"current_dir":"'$(pwd)'"}}' | ./ccstatus
```

### Widget Overview
- **User@Host** - Username and hostname
- **Path** - Current directory (truncated if long)
- **Git** - Branch name with change count (`master±3`)
- **Model** - Claude model (sonnet/opus/haiku)
- **Usage %** - Remaining capacity (color-coded: red<10%, yellow<30%, green>30%)
- **Weekly/Daily** - Shows most restrictive limit (weekly or daily usage %)
- **Tokens** - Token usage count (🔤 172.1k)
- **Cost** - Session cost estimate ($)
- **Messages** - Message count vs 5-hour window limit (💬 23/45)
- **Efficiency** - Context window utilization (📊 45.2%)
- **Compaction** - Distance to compaction threshold (🗜️ 68%)
- **Timer** - Time elapsed in current 5-hour block (⏱ 2h 15m)
- **Reset** - Time until next rate limit reset (5hr or weekly)

### Color Coding
- **Red**: Critical usage (>90% consumed)
- **Yellow**: High usage (70-90% consumed)
- **Green**: Normal usage (<70% consumed)

### Powerline Fonts
For best display, install a Powerline-compatible font like:
- Source Code Pro for Powerline
- Meslo LG S for Powerline  
- Cascadia Code PL
- Any Nerd Font

## Claude Pro Plan Limits (2025)

### Rate Limits
- **Messages**: ~45 messages per 5-hour rolling window (varies by conversation length)
- **Weekly**: 40-80 hours of Sonnet 4 usage per week (introduced August 2025)
- **Context Window**: 200K tokens standard, 1M tokens available with beta flag
- **Output**: Up to 64K output tokens for Sonnet 4

### Reset Cycles
- **5-hour windows**: Rolling window starting from first prompt, resets every 5 hours
- **Weekly limits**: Reset every Monday at 00:00 UTC (7-day cycle)

### Pricing (Per 1M Tokens)
- **Sonnet 4**: $3 input / $15 output (≤200K), $6 input / $22.50 output (>200K)
- **Opus 4**: $15 input / $75 output
- **Haiku**: $0.25 input / $1.25 output

## Architecture

### Theme System
Themes define colors for each widget and whether to use Powerline separators:

```go
type Theme struct {
    UserColor      string              // User@host color
    UserBg         string              // Background color
    PercentColor   func(int) string    // Dynamic color based on usage
    UsePowerline   bool                // Enable arrow separators
}
```

### Widget System
Each status component is a widget with content and styling:

```go
type Widget struct {
    Name    string    // Widget identifier
    Content string    // Display text
    Color   string    // Foreground color
    BgColor string    // Background color
}
```

### Integration Points
- **ccusage CLI**: Primary metrics source via `ccusage blocks --active --json`
- **calculate-usage.sh**: Fallback script in `~/.claude/`
- **Claude JSON**: Input context from Claude Code statusLine API
- **Git commands**: Live repository status
- **Session tracking**: 5-hour window and weekly limit tracking

## Contributing

The status line is designed to be:
- **Performant**: Fast startup, minimal dependencies
- **Extensible**: Easy to add new widgets and themes
- **Compatible**: Works with existing Claude Code patterns

## Inspiration

Built with inspiration from:
- [ccstatusline](https://github.com/sirmalloc/ccstatusline) - Powerline styling and widget architecture
- [ccusage](https://github.com/ryoppippi/ccusage) - Token usage tracking and metrics
- Powerline and airline themes from vim/terminal ecosystems