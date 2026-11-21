package mcp

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// The tests in this file pin the behavioral claims made in the tool
// descriptions and the server instructions. A description that tells agents
// how to interpret a result is a contract: when the handler changes in a way
// that breaks the claim, these tests fail and force the description to be
// updated together with the behavior.

type advertisedTool struct {
	Name        string
	Description string
	Properties  map[string]json.RawMessage
}

// listToolsViaWire sends a tools/list request through the MCP server and
// returns the advertised tools with name, description and schema properties.
func listToolsViaWire(t *testing.T, s *Server) []advertisedTool {
	t.Helper()

	raw, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	if err != nil {
		t.Fatalf("marshal tools/list request: %v", err)
	}

	wire, err := json.Marshal(s.mcpServer.HandleMessage(context.Background(), raw))
	if err != nil {
		t.Fatalf("marshal tools/list response: %v", err)
	}

	var parsed struct {
		Result struct {
			Tools []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				InputSchema struct {
					Properties map[string]json.RawMessage `json:"properties"`
				} `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(wire, &parsed); err != nil {
		t.Fatalf("decode tools/list response: %v (%s)", err, wire)
	}
	if len(parsed.Result.Tools) == 0 {
		t.Fatalf("expected advertised tools, got: %s", wire)
	}

	tools := make([]advertisedTool, 0, len(parsed.Result.Tools))
	for _, tt := range parsed.Result.Tools {
		tools = append(tools, advertisedTool{
			Name:        tt.Name,
			Description: tt.Description,
			Properties:  tt.InputSchema.Properties,
		})
	}
	return tools
}

func advertisedToolsByName(t *testing.T) map[string]advertisedTool {
	t.Helper()
	s, err := NewServer("")
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	byName := make(map[string]advertisedTool)
	for _, tool := range listToolsViaWire(t, s) {
		byName[tool.Name] = tool
	}
	return byName
}

// TestToolDescriptions_ContainBehaviorContracts asserts that the
// interpretation-critical claims live in the descriptions agents actually
// see when they pull a tool schema: silent-empty semantics, matching mode,
// language limits and provenance of node ids.
func TestToolDescriptions_ContainBehaviorContracts(t *testing.T) {
	contracts := map[string][]string{
		"grepai_search":          {"grep/rg"},
		"grepai_trace_callers":   {"bare symbol NAME", "empty result", "symbols_ready"},
		"grepai_trace_callees":   {"bare symbol NAME", "empty result", "symbols_ready"},
		"grepai_trace_graph":     {"grepai_trace_callers"},
		"grepai_refs_readers":    {"JavaScript/TypeScript"},
		"grepai_refs_writers":    {"JavaScript/TypeScript"},
		"grepai_refs_graph":      {"JavaScript/TypeScript"},
		"grepai_index_status":    {"symbols_ready", "rpg_enabled"},
		"grepai_rpg_search":      {"word-overlap", "node_id"},
		"grepai_rpg_fetch":       {"grepai_rpg_search", "not guessable"},
		"grepai_rpg_explore":     {"grepai_rpg_search", "not guessable"},
		"grepai_list_workspaces": {"workspace"},
		"grepai_list_projects":   {"workspace"},
	}

	byName := advertisedToolsByName(t)
	for name, wantPhrases := range contracts {
		tool, ok := byName[name]
		if !ok {
			t.Errorf("tool %q is not advertised", name)
			continue
		}
		for _, phrase := range wantPhrases {
			if !strings.Contains(tool.Description, phrase) {
				t.Errorf("%s: description no longer contains %q — update the description or this contract", name, phrase)
			}
		}
	}
}

// TestIndexStatusSchema_HasNoVerboseParam pins the removal of the dead
// "verbose" parameter: it was declared in the schema but never read by the
// handler. If it comes back, it must come back implemented (see the deferred
// bean) — not as a phantom option.
func TestIndexStatusSchema_HasNoVerboseParam(t *testing.T) {
	tool := advertisedToolsByName(t)["grepai_index_status"]
	if tool.Properties == nil {
		t.Fatal("grepai_index_status has no input schema properties")
	}
	if _, ok := tool.Properties["verbose"]; ok {
		t.Error("grepai_index_status advertises a 'verbose' parameter that the handler does not read")
	}
}

// TestIndexStatusJSON_HasGateFields pins that the JSON the status handler
// returns really carries the gate field names the descriptions and the
// instructions tell agents to check.
func TestIndexStatusJSON_HasGateFields(t *testing.T) {
	wire, err := json.Marshal(IndexStatus{})
	if err != nil {
		t.Fatalf("marshal IndexStatus: %v", err)
	}
	for _, field := range []string{"symbols_ready", "rpg_enabled"} {
		if !strings.Contains(string(wire), `"`+field+`"`) {
			t.Errorf("IndexStatus JSON no longer carries the %q gate field", field)
		}
	}
}

// TestInstructions_MentionEveryTool keeps the server instructions and the
// registered tool set in sync: every discovery-relevant tool must be routed
// in the instructions, and every tool name the instructions mention must
// exist. grepai_stats is exempt (metrics reporting, not code discovery).
func TestInstructions_MentionEveryTool(t *testing.T) {
	exempt := map[string]bool{"grepai_stats": true}

	byName := advertisedToolsByName(t)
	for name := range byName {
		if exempt[name] {
			continue
		}
		if !strings.Contains(serverInstructions, name) {
			t.Errorf("server instructions do not mention tool %q — add routing for it or extend the exempt list", name)
		}
	}

	tokenRe := regexp.MustCompile(`grepai_[a-z_]+`)
	for _, token := range tokenRe.FindAllString(serverInstructions, -1) {
		if _, ok := byName[token]; !ok {
			t.Errorf("server instructions mention %q, which is not a registered tool", token)
		}
	}
}
