package panels

import (
	"fmt"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/piechart"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"

	"github.com/grafana/grafana-foundation-sdk/go/common"
)

// Prometheus metric names for project tracking.
const (
	MetricProjectRequests = "claude_code_project_requests_total"
	MetricProjectCost     = "claude_code_project_cost_total"
)

// projectFilter filters by project variable only (for project-level metrics).
const projectFilter = `claude_code_project_name=~"$project"`

// ProjectCostBreakdown returns a piechart showing cost breakdown by project.
func ProjectCostBreakdown() cog.Builder[dashboard.Panel] {
	return piechart.NewPanelBuilder().
		Title("Project Cost Breakdown").
		Description("Cost distribution across projects over the selected range").
		Datasource(datasourceRef()).
		Height(8).
		Span(12).
		Unit("currencyUSD").
		PieType(piechart.PieChartTypeDonut).
		Legend(
			piechart.NewPieChartLegendOptionsBuilder().
				DisplayMode(common.LegendDisplayModeList).
				Placement(common.LegendPlacementBottom).
				ShowLegend(true),
		).
		Tooltip(singleTooltip()).
		ReduceOptions(
			common.NewReduceDataOptionsBuilder().
				Calcs([]string{"lastNotNull"}),
		).
		WithTarget(
			promInstantQuery(
				fmt.Sprintf(`sum by (claude_code_project_name) (increase(%s{%s}[$__range]))`, MetricProjectCost, projectFilter),
				"{{claude_code_project_name}}",
			),
		).
		Thresholds(greenThresholds()).
		ColorScheme(paletteColor())
}

// ProjectRequestsOverTime returns a stacked timeseries showing requests per project.
func ProjectRequestsOverTime() cog.Builder[dashboard.Panel] {
	return timeseries.NewPanelBuilder().
		Title("Project Requests Over Time").
		Description("Request rate broken down by project").
		Datasource(datasourceRef()).
		Height(8).
		Span(12).
		FillOpacity(30).
		Stacking(
			common.NewStackingConfigBuilder().
				Mode(common.StackingModeNormal),
		).
		Legend(defaultLegend()).
		Tooltip(multiTooltip()).
		WithTarget(
			promRangeQuery(
				fmt.Sprintf(`sum by (claude_code_project_name) (rate(%s{%s}[$__rate_interval]))`, MetricProjectRequests, projectFilter),
				"{{claude_code_project_name}}",
			),
		)
}
