---
# grepai-yb58
title: 'MCP doctrine: contract tests + verbose removal'
status: completed
type: task
priority: normal
created_at: 2026-11-21T14:10:23Z
updated_at: 2026-11-21T14:12:46Z
---

Pin tool-description claims with executable tests: tools/list schema assertions, instructions<->tools coverage lint, gate-field presence. Remove dead verbose param from index_status schema.

## Summary of Changes

- Removed dead verbose param from grepai_index_status schema (mcp/server.go); legacy test TestRegisterTools_IndexStatusSchema re-pinned to the format property.
- New mcp/descriptions_contract_test.go: TestToolDescriptions_ContainBehaviorContracts (phrases per tool via tools/list wire), TestIndexStatusSchema_HasNoVerboseParam, TestIndexStatusJSON_HasGateFields (symbols_ready/rpg_enabled tags), TestInstructions_MentionEveryTool (bidirectional instructions<->tools lint, grepai_stats exempt).
- CHANGELOG: Added (server instructions, honest tool descriptions), Removed (dead verbose param).
- Full ./mcp/ package: ok. Deferred: grepai-8y5y (treesitter CI job), grepai-mj3g (verbose re-implementation).
