package main

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/guicaulada/anthropic-otel-collector/dashboard/panels"
)

func buildUserDashboard() (dashboard.Dashboard, error) {
	builder := dashboard.NewDashboardBuilder("Claude Code User Dashboard").
		Uid("claude-code-user-dashboard").
		Description("User-centric monitoring dashboard for Claude Code usage patterns, productivity, and efficiency").
		Tags([]string{"anthropic", "claude", "user", "productivity"}).
		Refresh("30s").
		Time("now-24h", "now").
		Timezone("browser").
		Tooltip(dashboard.DashboardCursorSyncCrosshair).
		Variables(buildVariables()).
		// Row 1: Usage Overview
		WithRow(dashboard.NewRowBuilder("Usage Overview")).
		WithPanel(panels.UserTotalCost()).
		WithPanel(panels.UserActiveSessions()).
		WithPanel(panels.UserTotalRequests()).
		WithPanel(panels.UserAvgCostPerSession()).
		WithPanel(panels.UserCacheSavings()).
		WithPanel(panels.UserTotalTokens()).
		WithPanel(panels.UserErrorRate()).
		WithPanel(panels.UserAvgLatency()).
		// Row 2: Cost Intelligence
		WithRow(dashboard.NewRowBuilder("Cost Intelligence")).
		WithPanel(panels.UserCostOverTime()).
		WithPanel(panels.UserCostByProject()).
		WithPanel(panels.UserCostByModel()).
		WithPanel(panels.UserCacheSavingsOverTime()).
		WithPanel(panels.UserCumulativeCostByProject()).
		// Row 3: Session Activity
		WithRow(dashboard.NewRowBuilder("Session Activity")).
		WithPanel(panels.UserSessionsOverTime()).
		WithPanel(panels.UserConversationDepthOverTime()).
		WithPanel(panels.UserSessionDuration()).
		WithPanel(panels.UserSessionCostDistribution()).
		WithPanel(panels.UserAvgRequestsPerSession()).
		// Row 4: Productivity
		WithRow(dashboard.NewRowBuilder("Productivity")).
		WithPanel(panels.UserToolCallDistribution()).
		WithPanel(panels.UserLinesChangedOverTime()).
		WithPanel(panels.UserFileOperations()).
		WithPanel(panels.UserFilesTouched()).
		WithPanel(panels.UserToolCallsBySession()).
		// Row 5: Efficiency & Performance
		WithRow(dashboard.NewRowBuilder("Efficiency & Performance")).
		WithPanel(panels.UserCacheHitRatio()).
		WithPanel(panels.UserOutputUtilization()).
		WithPanel(panels.UserAvgTTFT()).
		WithPanel(panels.UserOutputThroughput()).
		WithPanel(panels.UserInputVsOutput()).
		WithPanel(panels.UserCostPerOutputToken()).
		// Row 6: Errors & Rate Limits
		WithRow(dashboard.NewRowBuilder("Errors & Rate Limits")).
		WithPanel(panels.UserErrorRateOverTime()).
		WithPanel(panels.UserErrorsByType()).
		WithPanel(panels.UserRateLimitUtilization()).
		WithPanel(panels.UserRateLimitRemaining())

	return builder.Build()
}
