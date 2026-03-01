package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/piechart"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"
)

// TokenUsageOverTime returns a stacked area timeseries showing token usage rates.
func TokenUsageOverTime() cog.Builder[dashboard.Panel] {
	return timeseries.NewPanelBuilder().
		Title("Token Usage Over Time").
		Description("Rate of token usage by type over time").
		Datasource(datasourceRef()).
		Height(8).
		Span(12).
		Unit("short").
		FillOpacity(30).
		Stacking(common.NewStackingConfigBuilder().Mode(common.StackingModeNormal)).
		Legend(defaultLegend()).
		Tooltip(multiTooltip()).
		WithTarget(
			promRangeQuery(
				f(`sum(rate(anthropic_tokens_input_total{%s}[$__rate_interval]))`),
				"Input",
			),
		).
		WithTarget(
			promRangeQuery(
				f(`sum(rate(anthropic_tokens_output_total{%s}[$__rate_interval]))`),
				"Output",
			),
		).
		WithTarget(
			promRangeQuery(
				f(`sum(rate(anthropic_tokens_cache_read_total{%s}[$__rate_interval]))`),
				"Cache Read",
			),
		).
		WithTarget(
			promRangeQuery(
				f(`sum(rate(anthropic_tokens_cache_creation_total{%s}[$__rate_interval]))`),
				"Cache Creation",
			),
		)
}

// CacheHitRatioOverTime returns a timeseries showing the cache hit ratio over time.
func CacheHitRatioOverTime() cog.Builder[dashboard.Panel] {
	return timeseries.NewPanelBuilder().
		Title("Cache Hit Ratio Over Time").
		Description("Average cache hit ratio over time").
		Datasource(datasourceRef()).
		Height(8).
		Span(12).
		Unit("percentunit").
		Min(0).
		Max(1).
		Legend(defaultLegend()).
		Tooltip(singleTooltip()).
		WithTarget(
			promRangeQuery(
				f(`avg(anthropic_cache_hit_ratio{%s})`),
				"Cache Hit Ratio",
			),
		)
}

// InputVsOutputBreakdown returns a donut piechart showing input vs output token distribution.
func InputVsOutputBreakdown() cog.Builder[dashboard.Panel] {
	return piechart.NewPanelBuilder().
		Title("Input vs Output Breakdown").
		Description("Distribution of input and output tokens over the selected range").
		Datasource(datasourceRef()).
		Height(8).
		Span(6).
		PieType(piechart.PieChartTypeDonut).
		Legend(
			piechart.NewPieChartLegendOptionsBuilder().
				DisplayMode(common.LegendDisplayModeList).
				Placement(common.LegendPlacementBottom).
				ShowLegend(true).
				Values([]piechart.PieChartLegendValues{
					piechart.PieChartLegendValuesValue,
					piechart.PieChartLegendValuesPercent,
				}),
		).
		Tooltip(multiTooltip()).
		ReduceOptions(
			common.NewReduceDataOptionsBuilder().
				Calcs([]string{"lastNotNull"}),
		).
		WithTarget(
			promInstantQuery(
				f(`sum(increase(anthropic_tokens_input_total{%s}[$__range]))`),
				"Input",
			),
		).
		WithTarget(
			promInstantQuery(
				f(`sum(increase(anthropic_tokens_output_total{%s}[$__range]))`),
				"Output",
			),
		)
}

// CacheTokensDetail returns a timeseries showing cache read and creation token rates.
func CacheTokensDetail() cog.Builder[dashboard.Panel] {
	return timeseries.NewPanelBuilder().
		Title("Cache Tokens Detail").
		Description("Rate of cache read and cache creation tokens").
		Datasource(datasourceRef()).
		Height(8).
		Span(9).
		Unit("short").
		Legend(defaultLegend()).
		Tooltip(multiTooltip()).
		WithTarget(
			promRangeQuery(
				f(`sum(rate(anthropic_tokens_cache_read_total{%s}[$__rate_interval]))`),
				"Cache Read",
			),
		).
		WithTarget(
			promRangeQuery(
				f(`sum(rate(anthropic_tokens_cache_creation_total{%s}[$__rate_interval]))`),
				"Cache Creation",
			),
		)
}

// OutputUtilization returns a timeseries showing output token utilization over time.
func OutputUtilization() cog.Builder[dashboard.Panel] {
	return timeseries.NewPanelBuilder().
		Title("Output Utilization").
		Description("Average output token utilization ratio over time").
		Datasource(datasourceRef()).
		Height(8).
		Span(9).
		Unit("percentunit").
		Legend(defaultLegend()).
		Tooltip(multiTooltip()).
		WithTarget(
			promRangeQuery(
				f(`avg(anthropic_tokens_output_utilization_ratio{%s})`),
				"Output Utilization",
			),
		)
}
