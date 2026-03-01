package anthropicreceiver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Usage tests ---

func TestUsage_TotalInputTokens(t *testing.T) {
	tests := []struct {
		name     string
		usage    Usage
		expected int
	}{
		{
			name:     "all zeros",
			usage:    Usage{},
			expected: 0,
		},
		{
			name: "only input tokens",
			usage: Usage{
				InputTokens: 100,
			},
			expected: 100,
		},
		{
			name: "input plus cache read",
			usage: Usage{
				InputTokens:          100,
				CacheReadInputTokens: 50,
			},
			expected: 150,
		},
		{
			name: "input plus cache creation",
			usage: Usage{
				InputTokens:              100,
				CacheCreationInputTokens: 30,
			},
			expected: 130,
		},
		{
			name: "all token types",
			usage: Usage{
				InputTokens:              100,
				CacheReadInputTokens:     50,
				CacheCreationInputTokens: 30,
			},
			expected: 180,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.usage.TotalInputTokens())
		})
	}
}

// --- ToolChoiceType tests ---

func TestAnthropicRequest_ToolChoiceType(t *testing.T) {
	t.Run("nil tool_choice", func(t *testing.T) {
		req := &AnthropicRequest{}
		assert.Equal(t, "", req.ToolChoiceType())
	})

	t.Run("null tool_choice", func(t *testing.T) {
		req := &AnthropicRequest{ToolChoice: json.RawMessage(`null`)}
		assert.Equal(t, "", req.ToolChoiceType())
	})

	t.Run("auto", func(t *testing.T) {
		req := &AnthropicRequest{ToolChoice: json.RawMessage(`{"type":"auto"}`)}
		assert.Equal(t, "auto", req.ToolChoiceType())
	})

	t.Run("any", func(t *testing.T) {
		req := &AnthropicRequest{ToolChoice: json.RawMessage(`{"type":"any"}`)}
		assert.Equal(t, "any", req.ToolChoiceType())
	})

	t.Run("tool with name", func(t *testing.T) {
		req := &AnthropicRequest{ToolChoice: json.RawMessage(`{"type":"tool","name":"my_tool"}`)}
		assert.Equal(t, "tool", req.ToolChoiceType())
	})

	t.Run("none", func(t *testing.T) {
		req := &AnthropicRequest{ToolChoice: json.RawMessage(`{"type":"none"}`)}
		assert.Equal(t, "none", req.ToolChoiceType())
	})

	t.Run("invalid json", func(t *testing.T) {
		req := &AnthropicRequest{ToolChoice: json.RawMessage(`invalid`)}
		assert.Equal(t, "", req.ToolChoiceType())
	})
}

// --- CacheHitRatio tests ---

func TestCacheHitRatio(t *testing.T) {
	tests := []struct {
		name     string
		usage    Usage
		expected float64
	}{
		{
			name:     "zero total returns zero",
			usage:    Usage{},
			expected: 0,
		},
		{
			name: "no cache reads",
			usage: Usage{
				InputTokens: 100,
			},
			expected: 0,
		},
		{
			name: "all from cache",
			usage: Usage{
				CacheReadInputTokens: 100,
			},
			expected: 1.0,
		},
		{
			name: "partial cache hit",
			usage: Usage{
				InputTokens:          50,
				CacheReadInputTokens: 50,
			},
			expected: 0.5,
		},
		{
			name: "mixed tokens with cache",
			usage: Usage{
				InputTokens:              100,
				CacheReadInputTokens:     200,
				CacheCreationInputTokens: 100,
			},
			expected: 0.5, // 200 / 400
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.expected, CacheHitRatio(tt.usage), 0.001)
		})
	}
}

// --- ExtractRateLimitInfo tests ---

func TestExtractRateLimitInfo(t *testing.T) {
	t.Run("all headers present", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("anthropic-ratelimit-requests-limit", "1000")
		headers.Set("anthropic-ratelimit-requests-remaining", "900")
		headers.Set("anthropic-ratelimit-input-tokens-limit", "100000")
		headers.Set("anthropic-ratelimit-input-tokens-remaining", "80000")
		headers.Set("anthropic-ratelimit-output-tokens-limit", "50000")
		headers.Set("anthropic-ratelimit-output-tokens-remaining", "40000")

		info := ExtractRateLimitInfo(headers)

		assert.Equal(t, 1000, info.RequestsLimit)
		assert.Equal(t, 900, info.RequestsRemaining)
		assert.Equal(t, 100000, info.InputTokensLimit)
		assert.Equal(t, 80000, info.InputTokensRemaining)
		assert.Equal(t, 50000, info.OutputTokensLimit)
		assert.Equal(t, 40000, info.OutputTokensRemaining)
	})

	t.Run("no headers", func(t *testing.T) {
		headers := http.Header{}
		info := ExtractRateLimitInfo(headers)

		assert.Equal(t, 0, info.RequestsLimit)
		assert.Equal(t, 0, info.RequestsRemaining)
		assert.Equal(t, 0, info.InputTokensLimit)
		assert.Equal(t, 0, info.InputTokensRemaining)
		assert.Equal(t, 0, info.OutputTokensLimit)
		assert.Equal(t, 0, info.OutputTokensRemaining)
	})

	t.Run("invalid header value defaults to zero", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("anthropic-ratelimit-requests-limit", "not-a-number")

		info := ExtractRateLimitInfo(headers)
		assert.Equal(t, 0, info.RequestsLimit)
	})
}

// --- RateLimitInfo utilization tests ---

func TestRateLimitInfo_Utilization(t *testing.T) {
	t.Run("zero limits return zero utilization", func(t *testing.T) {
		info := RateLimitInfo{}
		assert.Equal(t, 0.0, info.RequestsUtilization())
		assert.Equal(t, 0.0, info.InputTokensUtilization())
		assert.Equal(t, 0.0, info.OutputTokensUtilization())
	})

	t.Run("partial utilization", func(t *testing.T) {
		info := RateLimitInfo{
			RequestsLimit:         100,
			RequestsRemaining:     50,
			InputTokensLimit:      1000,
			InputTokensRemaining:  750,
			OutputTokensLimit:     500,
			OutputTokensRemaining: 100,
		}

		assert.InDelta(t, 0.5, info.RequestsUtilization(), 0.001)
		assert.InDelta(t, 0.25, info.InputTokensUtilization(), 0.001)
		assert.InDelta(t, 0.8, info.OutputTokensUtilization(), 0.001)
	})

	t.Run("full utilization", func(t *testing.T) {
		info := RateLimitInfo{
			RequestsLimit:     100,
			RequestsRemaining: 0,
		}
		assert.InDelta(t, 1.0, info.RequestsUtilization(), 0.001)
	})
}

// --- AnthropicRequest.SystemPromptSize tests ---

func TestAnthropicRequest_SystemPromptSize(t *testing.T) {
	t.Run("nil system returns zero", func(t *testing.T) {
		req := &AnthropicRequest{}
		assert.Equal(t, 0, req.SystemPromptSize())
	})

	t.Run("string system prompt", func(t *testing.T) {
		system := json.RawMessage(`"You are a helpful assistant"`)
		req := &AnthropicRequest{System: system}
		assert.Equal(t, len("You are a helpful assistant"), req.SystemPromptSize())
	})

	t.Run("array system prompt", func(t *testing.T) {
		system := json.RawMessage(`[{"text":"Block one"},{"text":"Block two"}]`)
		req := &AnthropicRequest{System: system}
		assert.Equal(t, len("Block one")+len("Block two"), req.SystemPromptSize())
	})

	t.Run("unparseable system returns raw length", func(t *testing.T) {
		system := json.RawMessage(`12345`)
		req := &AnthropicRequest{System: system}
		// Neither string nor array parse succeeds, falls back to len(r.System)
		assert.Equal(t, len(`12345`), req.SystemPromptSize())
	})
}

// --- AnthropicRequest.HasSystemPrompt tests ---

func TestAnthropicRequest_HasSystemPrompt(t *testing.T) {
	t.Run("nil system", func(t *testing.T) {
		req := &AnthropicRequest{}
		assert.False(t, req.HasSystemPrompt())
	})

	t.Run("empty system", func(t *testing.T) {
		req := &AnthropicRequest{System: json.RawMessage{}}
		assert.False(t, req.HasSystemPrompt())
	})

	t.Run("null system", func(t *testing.T) {
		req := &AnthropicRequest{System: json.RawMessage(`null`)}
		assert.False(t, req.HasSystemPrompt())
	})

	t.Run("present system prompt", func(t *testing.T) {
		req := &AnthropicRequest{System: json.RawMessage(`"Hello"`)}
		assert.True(t, req.HasSystemPrompt())
	})
}

// --- AnthropicRequest.MessageRoleCounts tests ---

func TestAnthropicRequest_MessageRoleCounts(t *testing.T) {
	t.Run("empty messages", func(t *testing.T) {
		req := &AnthropicRequest{}
		counts := req.MessageRoleCounts()
		assert.Empty(t, counts)
	})

	t.Run("single role", func(t *testing.T) {
		req := &AnthropicRequest{
			Messages: []Message{
				{Role: "user", Content: json.RawMessage(`"Hello"`)},
			},
		}
		counts := req.MessageRoleCounts()
		assert.Equal(t, 1, counts["user"])
		assert.Equal(t, 0, counts["assistant"])
	})

	t.Run("mixed roles", func(t *testing.T) {
		req := &AnthropicRequest{
			Messages: []Message{
				{Role: "user", Content: json.RawMessage(`"Hello"`)},
				{Role: "assistant", Content: json.RawMessage(`"Hi"`)},
				{Role: "user", Content: json.RawMessage(`"How are you?"`)},
				{Role: "assistant", Content: json.RawMessage(`"Good"`)},
				{Role: "user", Content: json.RawMessage(`"Great"`)},
			},
		}
		counts := req.MessageRoleCounts()
		assert.Equal(t, 3, counts["user"])
		assert.Equal(t, 2, counts["assistant"])
	})
}

// --- AnthropicResponse helper tests ---

func TestAnthropicResponse_TextContent(t *testing.T) {
	t.Run("empty content", func(t *testing.T) {
		resp := &AnthropicResponse{}
		assert.Equal(t, "", resp.TextContent())
	})

	t.Run("single text block", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "text", Text: "Hello world"},
			},
		}
		assert.Equal(t, "Hello world", resp.TextContent())
	})

	t.Run("multiple text blocks", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "text", Text: "Hello "},
				{Type: "tool_use", Name: "test"},
				{Type: "text", Text: "world"},
			},
		}
		assert.Equal(t, "Hello world", resp.TextContent())
	})

	t.Run("no text blocks", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "tool_use", Name: "test"},
			},
		}
		assert.Equal(t, "", resp.TextContent())
	})
}

func TestAnthropicResponse_ToolCalls(t *testing.T) {
	t.Run("no tool calls", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "text", Text: "Hello"},
			},
		}
		assert.Empty(t, resp.ToolCalls())
	})

	t.Run("has tool calls", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "text", Text: "Let me help"},
				{Type: "tool_use", Name: "Edit", ID: "tc_1"},
				{Type: "tool_use", Name: "Bash", ID: "tc_2"},
			},
		}
		calls := resp.ToolCalls()
		require.Len(t, calls, 2)
		assert.Equal(t, "Edit", calls[0].Name)
		assert.Equal(t, "Bash", calls[1].Name)
	})
}

func TestAnthropicResponse_ThinkingBlocks(t *testing.T) {
	t.Run("no thinking blocks", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "text", Text: "Hello"},
			},
		}
		assert.Empty(t, resp.ThinkingBlocks())
	})

	t.Run("has thinking blocks", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "thinking", Thinking: "Let me think..."},
				{Type: "text", Text: "The answer is 42"},
			},
		}
		blocks := resp.ThinkingBlocks()
		require.Len(t, blocks, 1)
		assert.Equal(t, "Let me think...", blocks[0].Thinking)
	})
}

func TestAnthropicResponse_ThinkingLength(t *testing.T) {
	t.Run("no thinking blocks", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "text", Text: "Hello"},
			},
		}
		assert.Equal(t, 0, resp.ThinkingLength())
	})

	t.Run("single thinking block", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "thinking", Thinking: "Let me think about this..."},
				{Type: "text", Text: "The answer is 42"},
			},
		}
		assert.Equal(t, len("Let me think about this..."), resp.ThinkingLength())
	})

	t.Run("multiple thinking blocks", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "thinking", Thinking: "First thought"},
				{Type: "text", Text: "Some text"},
				{Type: "thinking", Thinking: "Second thought"},
			},
		}
		assert.Equal(t, len("First thought")+len("Second thought"), resp.ThinkingLength())
	})

	t.Run("empty content", func(t *testing.T) {
		resp := &AnthropicResponse{}
		assert.Equal(t, 0, resp.ThinkingLength())
	})
}

func TestAnthropicResponse_ContentBlockCounts(t *testing.T) {
	t.Run("empty content", func(t *testing.T) {
		resp := &AnthropicResponse{}
		counts := resp.ContentBlockCounts()
		assert.Empty(t, counts)
	})

	t.Run("mixed content blocks", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "thinking"},
				{Type: "text"},
				{Type: "text"},
				{Type: "tool_use"},
				{Type: "tool_use"},
				{Type: "tool_use"},
			},
		}
		counts := resp.ContentBlockCounts()
		assert.Equal(t, 1, counts["thinking"])
		assert.Equal(t, 2, counts["text"])
		assert.Equal(t, 3, counts["tool_use"])
	})
}

// --- ServerToolUse tests ---

func TestUsage_ServerToolUse_Unmarshal(t *testing.T) {
	data := `{"input_tokens":100,"output_tokens":50,"speed":"fast","server_tool_use":{"web_search_requests":3,"web_fetch_requests":1,"code_execution_requests":2}}`
	var usage Usage
	err := json.Unmarshal([]byte(data), &usage)
	require.NoError(t, err)

	assert.Equal(t, 100, usage.InputTokens)
	assert.Equal(t, 50, usage.OutputTokens)
	assert.Equal(t, "fast", usage.Speed)
	require.NotNil(t, usage.ServerToolUse)
	assert.Equal(t, 3, usage.ServerToolUse.WebSearchRequests)
	assert.Equal(t, 1, usage.ServerToolUse.WebFetchRequests)
	assert.Equal(t, 2, usage.ServerToolUse.CodeExecutionRequests)
}

func TestUsage_Speed_Unmarshal(t *testing.T) {
	t.Run("standard speed", func(t *testing.T) {
		data := `{"input_tokens":100,"output_tokens":50,"speed":"standard"}`
		var usage Usage
		err := json.Unmarshal([]byte(data), &usage)
		require.NoError(t, err)
		assert.Equal(t, "standard", usage.Speed)
	})

	t.Run("no speed field", func(t *testing.T) {
		data := `{"input_tokens":100,"output_tokens":50}`
		var usage Usage
		err := json.Unmarshal([]byte(data), &usage)
		require.NoError(t, err)
		assert.Equal(t, "", usage.Speed)
	})
}

// --- Expanded MessageDeltaUsage tests ---

func TestMessageDeltaUsage_Unmarshal(t *testing.T) {
	data := `{"output_tokens":42,"input_tokens":100,"cache_read_input_tokens":50,"cache_creation_input_tokens":25}`
	var usage MessageDeltaUsage
	err := json.Unmarshal([]byte(data), &usage)
	require.NoError(t, err)

	assert.Equal(t, 42, usage.OutputTokens)
	assert.Equal(t, 100, usage.InputTokens)
	assert.Equal(t, 50, usage.CacheReadInputTokens)
	assert.Equal(t, 25, usage.CacheCreationInputTokens)
}

// --- Expanded RateLimitInfo tests ---

func TestExtractRateLimitInfo_NewFields(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-requests-limit", "1000")
	headers.Set("anthropic-ratelimit-requests-remaining", "900")
	headers.Set("anthropic-ratelimit-requests-reset", "2025-01-15T12:00:00Z")
	headers.Set("anthropic-ratelimit-input-tokens-reset", "2025-01-15T12:01:00Z")
	headers.Set("anthropic-ratelimit-output-tokens-reset", "2025-01-15T12:02:00Z")
	headers.Set("anthropic-ratelimit-tokens-limit", "200000")
	headers.Set("anthropic-ratelimit-tokens-remaining", "150000")
	headers.Set("x-anthropic-organization-id", "org_abc123")
	headers.Set("retry-after", "30")
	headers.Set("x-ratelimit-status", "rate_limited")

	info := ExtractRateLimitInfo(headers)

	assert.Equal(t, 1000, info.RequestsLimit)
	assert.Equal(t, 900, info.RequestsRemaining)
	assert.Equal(t, "2025-01-15T12:00:00Z", info.RequestsReset)
	assert.Equal(t, "2025-01-15T12:01:00Z", info.InputTokensReset)
	assert.Equal(t, "2025-01-15T12:02:00Z", info.OutputTokensReset)
	assert.Equal(t, 200000, info.TokensLimit)
	assert.Equal(t, 150000, info.TokensRemaining)
	assert.Equal(t, "org_abc123", info.OrganizationID)
	assert.Equal(t, "30", info.RetryAfter)
	assert.Equal(t, "rate_limited", info.UnifiedStatus)
}

// --- Credit usage header tests ---

func TestExtractRateLimitInfo_CreditUsageUSD(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("anthropic-ratelimit-requests-limit", "1000")
		headers.Set("anthropic-organization-user-credit-usage-usd", "12.50")

		info := ExtractRateLimitInfo(headers)
		assert.InDelta(t, 12.50, info.CreditUsageUSD, 0.001)
	})

	t.Run("absent", func(t *testing.T) {
		headers := http.Header{}
		info := ExtractRateLimitInfo(headers)
		assert.Equal(t, 0.0, info.CreditUsageUSD)
	})
}

func TestHeaderFloat64(t *testing.T) {
	t.Run("valid float", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("x-cost", "3.14")
		assert.InDelta(t, 3.14, headerFloat64(headers, "x-cost"), 0.001)
	})

	t.Run("empty header", func(t *testing.T) {
		headers := http.Header{}
		assert.Equal(t, 0.0, headerFloat64(headers, "x-cost"))
	})

	t.Run("invalid value", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("x-cost", "not-a-number")
		assert.Equal(t, 0.0, headerFloat64(headers, "x-cost"))
	})
}

// --- Rate limit header fallback tests ---

func TestExtractRateLimitInfo_NewFormatOnly(t *testing.T) {
	headers := http.Header{}
	headers.Set("ratelimit-limit", "500")
	headers.Set("ratelimit-remaining", "400")
	headers.Set("ratelimit-reset", "2025-06-01T12:00:00Z")

	info := ExtractRateLimitInfo(headers)

	assert.Equal(t, 500, info.RequestsLimit)
	assert.Equal(t, 400, info.RequestsRemaining)
	assert.Equal(t, "2025-06-01T12:00:00Z", info.RequestsReset)
	assert.Equal(t, 500, info.TokensLimit)
	assert.Equal(t, 400, info.TokensRemaining)
}

func TestExtractRateLimitInfo_OldTakesPrecedence(t *testing.T) {
	headers := http.Header{}
	// Old format (more granular)
	headers.Set("anthropic-ratelimit-requests-limit", "1000")
	headers.Set("anthropic-ratelimit-requests-remaining", "900")
	headers.Set("anthropic-ratelimit-requests-reset", "2025-06-01T12:00:00Z")
	// New format (should be ignored when old is present)
	headers.Set("ratelimit-limit", "500")
	headers.Set("ratelimit-remaining", "400")
	headers.Set("ratelimit-reset", "2025-06-01T13:00:00Z")

	info := ExtractRateLimitInfo(headers)

	assert.Equal(t, 1000, info.RequestsLimit, "old format should take precedence")
	assert.Equal(t, 900, info.RequestsRemaining, "old format should take precedence")
	assert.Equal(t, "2025-06-01T12:00:00Z", info.RequestsReset, "old format should take precedence")
}

func TestExtractRateLimitInfo_MixedFormats(t *testing.T) {
	headers := http.Header{}
	// Old format for some fields
	headers.Set("anthropic-ratelimit-input-tokens-limit", "100000")
	headers.Set("anthropic-ratelimit-input-tokens-remaining", "80000")
	// New format for request-level
	headers.Set("ratelimit-limit", "500")
	headers.Set("ratelimit-remaining", "400")

	info := ExtractRateLimitInfo(headers)

	assert.Equal(t, 500, info.RequestsLimit, "should use new format for requests")
	assert.Equal(t, 400, info.RequestsRemaining, "should use new format for requests")
	assert.Equal(t, 100000, info.InputTokensLimit, "should use old format for input tokens")
	assert.Equal(t, 80000, info.InputTokensRemaining, "should use old format for input tokens")
}

func TestHeaderIntFallback(t *testing.T) {
	t.Run("primary present", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("primary", "100")
		headers.Set("fallback", "50")
		assert.Equal(t, 100, headerIntFallback(headers, "primary", "fallback"))
	})

	t.Run("only fallback present", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("fallback", "50")
		assert.Equal(t, 50, headerIntFallback(headers, "primary", "fallback"))
	})

	t.Run("neither present", func(t *testing.T) {
		headers := http.Header{}
		assert.Equal(t, 0, headerIntFallback(headers, "primary", "fallback"))
	})
}

func TestHeaderStrFallback(t *testing.T) {
	t.Run("primary present", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("primary", "value1")
		headers.Set("fallback", "value2")
		assert.Equal(t, "value1", headerStrFallback(headers, "primary", "fallback"))
	})

	t.Run("only fallback present", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("fallback", "value2")
		assert.Equal(t, "value2", headerStrFallback(headers, "primary", "fallback"))
	})

	t.Run("neither present", func(t *testing.T) {
		headers := http.Header{}
		assert.Equal(t, "", headerStrFallback(headers, "primary", "fallback"))
	})
}

// --- ContentBlock.Citations tests ---

func TestContentBlock_Citations_Unmarshal(t *testing.T) {
	data := `{"type":"text","text":"According to sources...","citations":[{"type":"web","url":"https://example.com"}]}`
	var block ContentBlock
	err := json.Unmarshal([]byte(data), &block)
	require.NoError(t, err)

	assert.Equal(t, "text", block.Type)
	assert.Equal(t, "According to sources...", block.Text)
	assert.NotNil(t, block.Citations)
	assert.Contains(t, string(block.Citations), "web")
}

// --- RedactedThinking tests ---

func TestAnthropicResponse_RedactedThinkingBlocks(t *testing.T) {
	t.Run("no redacted thinking", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "thinking", Thinking: "visible"},
				{Type: "text", Text: "answer"},
			},
		}
		assert.Empty(t, resp.RedactedThinkingBlocks())
		assert.Equal(t, 0, resp.RedactedThinkingCount())
	})

	t.Run("has redacted thinking", func(t *testing.T) {
		resp := &AnthropicResponse{
			Content: []ContentBlock{
				{Type: "thinking", Thinking: "visible"},
				{Type: "redacted_thinking", Data: "base64data1"},
				{Type: "redacted_thinking", Data: "base64data2"},
				{Type: "text", Text: "answer"},
			},
		}
		blocks := resp.RedactedThinkingBlocks()
		require.Len(t, blocks, 2)
		assert.Equal(t, "base64data1", blocks[0].Data)
		assert.Equal(t, "base64data2", blocks[1].Data)
		assert.Equal(t, 2, resp.RedactedThinkingCount())
	})
}

func TestContentBlock_Data_Unmarshal(t *testing.T) {
	data := `{"type":"redacted_thinking","data":"aGVsbG8gd29ybGQ="}`
	var block ContentBlock
	err := json.Unmarshal([]byte(data), &block)
	require.NoError(t, err)

	assert.Equal(t, "redacted_thinking", block.Type)
	assert.Equal(t, "aGVsbG8gd29ybGQ=", block.Data)
}

func TestAnthropicResponse_ContentBlockCounts_IncludesRedactedThinking(t *testing.T) {
	resp := &AnthropicResponse{
		Content: []ContentBlock{
			{Type: "thinking"},
			{Type: "redacted_thinking"},
			{Type: "redacted_thinking"},
			{Type: "text"},
		},
	}
	counts := resp.ContentBlockCounts()
	assert.Equal(t, 1, counts["thinking"])
	assert.Equal(t, 2, counts["redacted_thinking"])
	assert.Equal(t, 1, counts["text"])
}

// --- Container tests ---

func TestAnthropicResponse_Container_Unmarshal(t *testing.T) {
	data := `{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-6","stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":5},"container":{"id":"ctr_abc123","expires_at":"2025-06-01T12:00:00Z"}}`
	var resp AnthropicResponse
	err := json.Unmarshal([]byte(data), &resp)
	require.NoError(t, err)

	require.NotNil(t, resp.Container)
	assert.Equal(t, "ctr_abc123", resp.Container.ID)
	assert.Equal(t, "2025-06-01T12:00:00Z", resp.Container.ExpiresAt)
}

func TestAnthropicResponse_Container_Nil(t *testing.T) {
	data := `{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-6","stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":5}}`
	var resp AnthropicResponse
	err := json.Unmarshal([]byte(data), &resp)
	require.NoError(t, err)

	assert.Nil(t, resp.Container)
}

// --- Delta.Signature tests ---

func TestDelta_Signature_Unmarshal(t *testing.T) {
	data := `{"type":"signature_delta","signature":"abc123sig"}`
	var delta Delta
	err := json.Unmarshal([]byte(data), &delta)
	require.NoError(t, err)

	assert.Equal(t, "signature_delta", delta.Type)
	assert.Equal(t, "abc123sig", delta.Signature)
}
