# grepai — code discovery routing for agents

grepai provides semantic search and call-graph analysis over this indexed
codebase. Prefer it over grep/rg for code discovery.

## Tool selection by question

- intent ("where is X handled", "what does Y do") → `grepai_search`
  (English intent queries, `compact: true`, `limit` 3-5)
- exact string / identifier / literal → text grep is correct
- impact before changing a function → `grepai_trace_callers`
- what a function depends on → `grepai_trace_callees`
- which feature does this belong to / what surrounds it → `grepai_rpg_fetch`,
  `grepai_rpg_explore`
- "missing" code / index health → `grepai_index_status`

## Composing loop

`grepai_search` finds candidates → trace/RPG verify structure → Read the
bounded line range the verification points to. Never re-verify a classified
grepai result with a text sweep.

Fall back to grep only when grepai tools error or the index is stale
(check `grepai_index_status` first).
