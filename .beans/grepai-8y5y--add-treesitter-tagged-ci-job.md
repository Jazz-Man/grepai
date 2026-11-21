---
# grepai-8y5y
title: Add treesitter-tagged CI job
status: todo
type: task
priority: low
created_at: 2026-11-21T14:10:23Z
updated_at: 2026-11-21T14:16:40Z
---

Add a CI job running 'make test-ast' (added to the Makefile: CGO_ENABLED=1 go test -tags treesitter ./...) — currently AST/treesitter tests never run in CI, and 'make test' skips them locally too (build tag). Deferred: user works fork-locally for now.
