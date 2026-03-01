package main

// MetricRequests is the Prometheus metric name for request counts,
// used in template variable queries.
const MetricRequests = "anthropic_requests_total"

// MetricProjectRequests is the Prometheus metric name for project-level request counts.
const MetricProjectRequests = "claude_code_project_requests_total"

// MetricCacheSavings is the Prometheus metric for cache savings.
const MetricCacheSavings = "anthropic_cost_cache_savings_total"

// MetricOutputUtilization is the Prometheus metric for output token utilization.
const MetricOutputUtilization = "anthropic_tokens_output_utilization_ratio"

// MetricProjectTokensInput is the Prometheus metric for project input tokens.
const MetricProjectTokensInput = "claude_code_project_tokens_input_total"

// MetricProjectTokensOutput is the Prometheus metric for project output tokens.
const MetricProjectTokensOutput = "claude_code_project_tokens_output_total"

// MetricProjectErrors is the Prometheus metric for project errors.
const MetricProjectErrors = "claude_code_project_errors_total"

// MetricErrorsByType is the Prometheus metric for errors broken down by type.
const MetricErrorsByType = "anthropic_errors_by_type_total"

func strPtr(s string) *string {
	return &s
}
