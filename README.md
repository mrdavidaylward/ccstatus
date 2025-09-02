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

📊 **Advanced Metrics**
- 💰 **Cost tracking** - Session costs based on token usage and model pricing
- 💬 **Message count** - Current messages used in 5-hour rate limit window
- 📊 **Context efficiency** - Context window utilization percentage
- 🗜️ **Compaction warning** - Percentage until message compaction
- 📅 **Weekly vs Daily limits** - Smart tracking of both Claude limit types
- ⏱ **5-hour rolling windows** - Accurate reset timers based on actual Claude limits
- Git branch with change count (`master±3`)

🔧 **Smart Integration**
- Enhanced `ccusage` CLI tool integration with session tracking
- 5-hour rolling window calculations (not daily estimates)
- Context window usage tracking (200K limit)
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
- **Model** - Claude model (opus/sonnet)
- **Usage %** - Remaining tokens (color-coded: red<10%, yellow<30%, green>30%)
- **Tokens** - Daily usage count (🔤 172.1k)  
- **Timer** - Block elapsed time (⏱ 2m)
- **Reset** - Time until next rate limit reset

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
- **ccusage CLI**: Primary metrics source via `ccusage stats --json`
- **calculate-usage.sh**: Fallback script in `~/.claude/`
- **Claude JSON**: Input context from Claude Code
- **Git commands**: Live repository status

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