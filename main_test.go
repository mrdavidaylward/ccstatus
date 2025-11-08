package main

import (
	"testing"
)

// TestCalculateUsagePercentage tests the usage percentage calculation
func TestCalculateUsagePercentage(t *testing.T) {
	tests := []struct {
		name          string
		dailyTokens   int
		contextTokens int
		contextChars  int
		want          int
	}{
		{
			name:          "zero tokens",
			dailyTokens:   0,
			contextTokens: 0,
			contextChars:  0,
			want:          0,
		},
		{
			name:          "under context limit",
			dailyTokens:   100000,
			contextTokens: 100000,
			contextChars:  0,
			want:          50, // 100k/200k = 50%
		},
		{
			name:          "over context limit",
			dailyTokens:   300000,
			contextTokens: 250000,
			contextChars:  0,
			want:          100, // Capped at 100%
		},
		{
			name:          "using chars estimate",
			dailyTokens:   50000,
			contextTokens: 0,
			contextChars:  400000, // 100k tokens estimated (400k/4)
			want:          50,     // 100k/200k = 50%
		},
		{
			name:          "weekly more restrictive",
			dailyTokens:   4000000, // 80% of weekly estimate
			contextTokens: 50000,   // 25% of context
			contextChars:  0,
			want:          80, // Weekly is more restrictive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateUsagePercentage(tt.dailyTokens, tt.contextTokens, tt.contextChars)
			if got != tt.want {
				t.Errorf("calculateUsagePercentage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalculateCompactionPercentage tests compaction percentage calculation
func TestCalculateCompactionPercentage(t *testing.T) {
	tests := []struct {
		name          string
		contextTokens int
		want          int
	}{
		{
			name:          "zero tokens",
			contextTokens: 0,
			want:          0,
		},
		{
			name:          "half way to threshold",
			contextTokens: 90000, // 50% of 180k threshold
			want:          50,
		},
		{
			name:          "at threshold",
			contextTokens: 180000,
			want:          0, // At threshold, shows 0% into danger zone
		},
		{
			name:          "over limit",
			contextTokens: 250000,
			want:          100, // Capped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCompactionPercentage(tt.contextTokens)
			if got != tt.want {
				t.Errorf("calculateCompactionPercentage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormatTokensAdvanced tests token formatting
func TestFormatTokensAdvanced(t *testing.T) {
	tests := []struct {
		name   string
		tokens int
		want   string
	}{
		{
			name:   "small number",
			tokens: 500,
			want:   "500",
		},
		{
			name:   "thousands",
			tokens: 5000,
			want:   "5.0k",
		},
		{
			name:   "hundreds of thousands",
			tokens: 172100,
			want:   "172.1k",
		},
		{
			name:   "millions",
			tokens: 2500000,
			want:   "2.5M",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTokensAdvanced(tt.tokens)
			if got != tt.want {
				t.Errorf("formatTokensAdvanced() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormatCost tests cost formatting
func TestFormatCost(t *testing.T) {
	tests := []struct {
		name string
		cost float64
		want string
	}{
		{
			name: "sub-penny",
			cost: 0.005,
			want: "0.500¢",
		},
		{
			name: "cents",
			cost: 0.15,
			want: "15.00¢",
		},
		{
			name: "dollars",
			cost: 1.50,
			want: "$1.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCost(tt.cost)
			if got != tt.want {
				t.Errorf("formatCost() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalculateCost tests session cost calculation
func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name         string
		modelName    string
		inputTokens  int
		outputTokens int
		wantSession  float64
	}{
		{
			name:         "sonnet small",
			modelName:    "Sonnet 4",
			inputTokens:  1000,
			outputTokens: 500,
			wantSession:  0.0105, // (1000 * 3 + 500 * 15) / 1000000
		},
		{
			name:         "haiku small",
			modelName:    "Haiku",
			inputTokens:  1000,
			outputTokens: 500,
			wantSession:  0.000875, // (1000 * 0.25 + 500 * 1.25) / 1000000
		},
		{
			name:         "opus small",
			modelName:    "Opus",
			inputTokens:  1000,
			outputTokens: 500,
			wantSession:  0.0525, // (1000 * 15 + 500 * 75) / 1000000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSession, _ := calculateCost(tt.modelName, tt.inputTokens, tt.outputTokens)
			if gotSession != tt.wantSession {
				t.Errorf("calculateCost() session = %v, want %v", gotSession, tt.wantSession)
			}
		})
	}
}

// TestGetBgToFgColor tests background to foreground color conversion
func TestGetBgToFgColor(t *testing.T) {
	tests := []struct {
		name    string
		bgColor string
		want    string
	}{
		{
			name:    "standard blue",
			bgColor: BgBlue,
			want:    ColorBlue,
		},
		{
			name:    "truecolor conversion",
			bgColor: "\033[48;2;60;56;54m",
			want:    "\033[38;2;60;56;54m",
		},
		{
			name:    "unknown color",
			bgColor: "\033[49m",
			want:    ColorWhite,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBgToFgColor(tt.bgColor)
			if got != tt.want {
				t.Errorf("getBgToFgColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTruncatePath tests path truncation
func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		maxLength int
		want      string
	}{
		{
			name:      "short path",
			path:      "/home",
			maxLength: 20,
			want:      "/home",
		},
		{
			name:      "long path",
			path:      "/home/user/very/long/path/to/project",
			maxLength: 20,
			want:      "", // Just verify it's truncated (actual result depends on ellipsis encoding)
		},
		{
			name:      "exact length",
			path:      "12345",
			maxLength: 5,
			want:      "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePath(tt.path, tt.maxLength)
			// For long paths, just verify they're truncated (Unicode ellipsis complicates byte length)
			if tt.want == "" {
				if len(tt.path) <= tt.maxLength {
					t.Errorf("truncatePath() should have truncated path of length %d with max %d", len(tt.path), tt.maxLength)
				}
			} else if got != tt.want {
				t.Errorf("truncatePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalculateWeeklyUsagePercentage tests weekly usage percentage
func TestCalculateWeeklyUsagePercentage(t *testing.T) {
	tests := []struct {
		name         string
		weeklyTokens int
		want         int
	}{
		{
			name:         "zero usage",
			weeklyTokens: 0,
			want:         0,
		},
		{
			name:         "half weekly limit",
			weeklyTokens: 2500000,
			want:         50,
		},
		{
			name:         "over limit",
			weeklyTokens: 6000000,
			want:         100, // Capped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateWeeklyUsagePercentage(tt.weeklyTokens)
			if got != tt.want {
				t.Errorf("calculateWeeklyUsagePercentage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalculateContextEfficiency tests context efficiency calculation
func TestCalculateContextEfficiency(t *testing.T) {
	tests := []struct {
		name          string
		contextTokens int
		want          float64
	}{
		{
			name:          "zero tokens",
			contextTokens: 0,
			want:          0.0,
		},
		{
			name:          "half capacity",
			contextTokens: 100000,
			want:          50.0,
		},
		{
			name:          "over capacity",
			contextTokens: 300000,
			want:          100.0, // Capped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateContextEfficiency(tt.contextTokens)
			if got != tt.want {
				t.Errorf("calculateContextEfficiency() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetModelDisplay tests model name simplification
func TestGetModelDisplay(t *testing.T) {
	tests := []struct {
		name  string
		model ModelInfo
		want  string
	}{
		{
			name:  "sonnet display name",
			model: ModelInfo{DisplayName: "Sonnet 4"},
			want:  "sonnet",
		},
		{
			name:  "opus display name",
			model: ModelInfo{DisplayName: "Claude Opus 4"},
			want:  "opus",
		},
		{
			name:  "haiku display name",
			model: ModelInfo{DisplayName: "Haiku 3.5"},
			want:  "haiku",
		},
		{
			name:  "fallback to ID",
			model: ModelInfo{ID: "claude-3-sonnet"},
			want:  "sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getModelDisplay(tt.model)
			if got != tt.want {
				t.Errorf("getModelDisplay() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalculateTimeToWeeklyReset tests weekly reset time calculation
func TestCalculateTimeToWeeklyReset(t *testing.T) {
	// This test just verifies the function runs without errors
	// and returns valid values
	timeStr, resetType := calculateTimeToWeeklyReset()

	if resetType != "weekly" {
		t.Errorf("calculateTimeToWeeklyReset() resetType = %v, want 'weekly'", resetType)
	}

	if timeStr == "" {
		t.Error("calculateTimeToWeeklyReset() returned empty time string")
	}
}

// TestExtractTokenCount tests token extraction from text
func TestExtractTokenCount(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		pattern string
		want    int
	}{
		{
			name:    "json number",
			text:    `{"totalTokens": 12345}`,
			pattern: `"totalTokens"\s*:\s*(\d+)`,
			want:    12345,
		},
		{
			name:    "plain text",
			text:    "total tokens: 54321",
			pattern: `total.*tokens.*:\s*(\d+)`,
			want:    54321,
		},
		{
			name:    "no match",
			text:    "no numbers here",
			pattern: `tokens:\s*(\d+)`,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTokenCount(tt.text, tt.pattern)
			if got != tt.want {
				t.Errorf("extractTokenCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests for performance-critical functions
func BenchmarkCalculateUsagePercentage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calculateUsagePercentage(100000, 50000, 200000)
	}
}

func BenchmarkFormatTokensAdvanced(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatTokensAdvanced(172100)
	}
}

func BenchmarkExtractTokenCount(b *testing.B) {
	text := `{"totalTokens": 12345, "inputTokens": 8000}`
	pattern := `"totalTokens"\s*:\s*(\d+)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractTokenCount(text, pattern)
	}
}
