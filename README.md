# CCStatus - Enhanced Claude Code Status Line

A powerful Go-based status line for Claude Code with Powerline styling, multiple themes, and advanced metrics inspired by ccstatusline and ccusage.

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
- Session and daily token usage with k/M formatting
- Claude rate limit tracking (5hr blocks)
- Remaining usage percentage with color coding
- Block timer showing elapsed time in current 5hr window
- Git branch with change count (`master±3`)

🔧 **Smart Integration**
- `ccusage` CLI tool support for accurate metrics
- Custom calculation script integration
- Context window usage tracking (200K limit)
- Multiple token data source fallbacks

## Installation

1. Build the binary:
```bash
go build -o ccstatus main.go
chmod +x ccstatus
```

2. Configure Claude Code:
```bash
# Update ~/.claude/settings.json
{
  "statusLine": {
    "type": "command", 
    "command": "/path/to/ccstatus/ccstatus"
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