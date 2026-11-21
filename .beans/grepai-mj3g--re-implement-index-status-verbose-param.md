---
# grepai-mj3g
title: Re-implement index_status verbose param
status: todo
type: task
priority: low
created_at: 2026-11-21T14:10:23Z
updated_at: 2026-11-21T14:10:23Z
---

The verbose parameter was removed from the grepai_index_status schema (declared but never read by the handler). Future option: implement it meaningfully — config summary (embedder provider/model, store backend, chunking strategy), watcher PID state, index-meta embedder identity. Re-adding a working param is upstream-compatible (additions-only).
