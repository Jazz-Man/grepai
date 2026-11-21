# grepai — code discovery routing for agents

grepai indexes this codebase: semantic search, symbol call graphs, and a
feature graph. For ANY code discovery here, grepai tools come BEFORE grep/rg;
text grep is correct only for exact strings, identifiers, and literals.

## Tool selection by question

- intent ("where is X handled", "what does Y do") → `grepai_search`
  (English intent queries, not identifiers; `compact: true`, `limit` 3-5)
- impact before changing a function → `grepai_trace_callers`
- what a function depends on → `grepai_trace_callees`
- call graph around a symbol → `grepai_trace_graph` (`depth`)
- where a property is read or written (JS/TS only) → `grepai_refs_readers` /
  `grepai_refs_writers` / `grepai_refs_graph`
- which feature does this belong to → `grepai_rpg_search` FIRST — it yields
  the `node_id` — then `grepai_rpg_fetch` (node detail) or
  `grepai_rpg_explore` (BFS neighborhood)
- "missing" code, wrong results, tool errors → `grepai_index_status`
- workspace / project names → `grepai_list_workspaces`,
  `grepai_list_projects`

A code-research task is not one search: combine search + trace + RPG before
concluding.

## How the tools match

- `grepai_search` is the only embeddings tool — intent phrasing, synonyms work.
- trace and refs match by bare symbol NAME: same-name symbols in other files
  are indistinguishable, and a miss returns an EMPTY result, not an error.
- `grepai_rpg_search` is word-overlap, not semantic — use the literal words
  of the feature or area, not paraphrases.

## Composing loop

`grepai_search` finds candidates → trace/RPG verify structure → Read the
bounded line range the verification points to. Never re-verify a classified
grepai result with a text sweep.

## Prerequisites, scope, economy

trace/refs need the symbol index and RPG needs the feature graph — both are
built by the owner (`grepai watch` / `grepai index`); when results look
wrong, check `grepai_index_status` (`symbols_ready`, `rpg_enabled`) before
blaming the query. Tools default to the server's startup project; a per-call
`root` or `workspace` overrides it (mutually exclusive). Fall back to text
grep only when grepai tools error or the index is stale. Economy on nearly
every tool: `compact: true`, `format: "toon"`.
