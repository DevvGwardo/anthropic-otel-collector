package panels

import (
	"fmt"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/piechart"
	"github.com/grafana/grafana-foundation-sdk/go/stat"
)

// TotalLinesChanged returns a stat panel showing total lines added and removed.
func TotalLinesChanged() cog.Builder[dashboard.Panel] {
	return stat.NewPanelBuilder().
		Title("Total Lines Changed").
		Datasource(datasourceRef()).
		Height(8).
		Span(6).
		Unit("short").
		Thresholds(greenThresholds()).
		GraphMode(common.BigValueGraphModeArea).
		ColorMode(common.BigValueColorModeValue).
		ReduceOptions(
			common.NewReduceDataOptionsBuilder().
				Calcs([]string{"lastNotNull"}),
		).
		WithTarget(
			promInstantQuery(
				`sum(increase(anthropic_tool_use_lines_added_total{`+filter+`}[$__range])) + sum(increase(anthropic_tool_use_lines_removed_total{`+filter+`}[$__range]))`,
				"Total Lines Changed",
			),
		)
}

// TotalFilesTouched returns a stat panel showing total files touched.
func TotalFilesTouched() cog.Builder[dashboard.Panel] {
	return stat.NewPanelBuilder().
		Title("Total Files Touched").
		Datasource(datasourceRef()).
		Height(8).
		Span(6).
		Unit("short").
		Thresholds(greenThresholds()).
		GraphMode(common.BigValueGraphModeArea).
		ColorMode(common.BigValueColorModeValue).
		ReduceOptions(
			common.NewReduceDataOptionsBuilder().
				Calcs([]string{"lastNotNull"}),
		).
		WithTarget(
			promInstantQuery(
				`sum(increase(anthropic_tool_use_files_touched_total{`+filter+`}[$__range]))`,
				"Total Files Touched",
			),
		)
}

// ProjectRequestsBreakdown returns a donut piechart showing requests by project.
func ProjectRequestsBreakdown() cog.Builder[dashboard.Panel] {
	return piechart.NewPanelBuilder().
		Title("Project Requests Breakdown").
		Description("Request distribution across projects over the selected range").
		Datasource(datasourceRef()).
		Height(8).
		Span(8).
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
				fmt.Sprintf(`sum by (claude_code_project_name) (increase(%s{%s}[$__range]))`, MetricProjectRequests, projectFilter),
				"{{claude_code_project_name}}",
			),
		).
		Thresholds(greenThresholds()).
		ColorScheme(paletteColor())
}
