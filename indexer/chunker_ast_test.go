//go:build treesitter

package indexer

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestASTChunker_ZigFile(t *testing.T) {
	src := `const std = @import("std");

fn add(a: i32, b: i32) i32 {
    return a + b;
}

pub fn main() void {
    const result = add(1, 2);
    std.debug.print("{d}\n", .{result});
}

test "addition" {
    try std.testing.expectEqual(@as(i32, 3), add(1, 2));
}

const Point = struct {
    x: i32,
    y: i32,

    fn norm(self: Point) i32 {
        return self.x + self.y;
    }
};
`
	ac := NewASTChunker(NewChunker(512, 50))
	chunks := ac.ChunkWithContext("main.zig", src)

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	for i, c := range chunks {
		if !strings.HasPrefix(c.Content, "File: main.zig") {
			t.Errorf("chunk %d missing file context prefix", i)
		}
		if c.FilePath != "main.zig" {
			t.Errorf("chunk %d: expected file path main.zig, got %s", i, c.FilePath)
		}
		if c.StartLine < 1 {
			t.Errorf("chunk %d: invalid start line %d", i, c.StartLine)
		}
	}

	combined := ""
	for _, c := range chunks {
		combined += strings.TrimPrefix(c.Content, "File: main.zig\n\n")
	}
	for _, want := range []string{
		"fn add(a: i32, b: i32) i32 {",
		"pub fn main() void {",
		`test "addition" {`,
		"const Point = struct {",
		"fn norm(self: Point) i32 {",
	} {
		if !strings.Contains(combined, want) {
			t.Errorf("missing %q in chunked output", want)
		}
	}
}

func TestASTChunker_RustFile(t *testing.T) {
	src := `use std::collections::HashMap;

fn add(a: i32, b: i32) -> i32 {
    a + b
}

pub fn main() {
    let result = add(1, 2);
    println!("{}", result);
}

struct Point {
    x: i32,
    y: i32,
}

impl Point {
    fn norm(&self) -> i32 {
        self.x + self.y
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn addition() {
        assert_eq!(add(1, 2), 3);
    }
}
`
	ac := NewASTChunker(NewChunker(512, 50))
	chunks := ac.ChunkWithContext("main.rs", src)

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	for i, c := range chunks {
		if !strings.HasPrefix(c.Content, "File: main.rs") {
			t.Errorf("chunk %d missing file context prefix", i)
		}
		if c.FilePath != "main.rs" {
			t.Errorf("chunk %d: expected file path main.rs, got %s", i, c.FilePath)
		}
		if c.StartLine < 1 {
			t.Errorf("chunk %d: invalid start line %d", i, c.StartLine)
		}
	}

	combined := ""
	for _, c := range chunks {
		combined += strings.TrimPrefix(c.Content, "File: main.rs\n\n")
	}
	for _, want := range []string{
		"fn add(a: i32, b: i32) -> i32 {",
		"pub fn main() {",
		"struct Point {",
		"impl Point {",
		"fn norm(&self) -> i32 {",
		"fn addition() {",
	} {
		if !strings.Contains(combined, want) {
			t.Errorf("missing %q in chunked output", want)
		}
	}
}

// TestASTChunker_LanguageSuite runs a small real-code fixture through the AST
// chunker for every grammar wired into the languages map beyond the Go/Zig/Rust
// standalone tests. Each case asserts the same reconstruction contract:
// chunks carry the File: prefix, the file path, and reassemble the fixture
// with the language's key constructs intact.
func TestASTChunker_LanguageSuite(t *testing.T) {
	cases := []struct {
		file  string
		src   string
		wants []string
	}{
		{
			file: "page.html",
			src: `<!DOCTYPE html>
<html>
<head><title>demo</title></head>
<body>
  <nav id="main">
    <a href="/one">one</a>
    <a href="/two">two</a>
  </nav>
  <section id="content">
    <p>hello</p>
  </section>
</body>
</html>
`,
			wants: []string{`<nav id="main">`, `<section id="content">`},
		},
		{
			file: "style.css",
			src: `.header {
  color: red;
}

.body .container {
  margin: 0 auto;
}
`,
			wants: []string{".header {", ".body .container {"},
		},
		{
			file: "workflow.yaml",
			src: `name: demo
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout
`,
			wants: []string{"jobs:", "runs-on: ubuntu-latest"},
		},
		{
			file: "config.toml",
			src: `[server]
port = 8080

[database]
url = "postgres://localhost/app"
`,
			wants: []string{"[server]", "[database]"},
		},
		{
			file: "user.proto",
			src: `syntax = "proto3";

message User {
  string name = 1;
  int32 age = 2;
}
`,
			wants: []string{"message User {"},
		},
		{
			file: "main.tf",
			src: `resource "aws_instance" "web" {
  ami = "ami-123456"
  count = 2
}
`,
			wants: []string{`resource "aws_instance" "web" {`},
		},
		{
			file: "Main.elm",
			src: `module Main exposing (main)

add : Int -> Int -> Int
add a b =
    a + b

main =
    add 1 2
`,
			wants: []string{"add : Int -> Int -> Int"},
		},
		{
			file: "util.lua",
			src: `local M = {}

function M.add(a, b)
    return a + b
end

return M
`,
			wants: []string{"function M.add(a, b)"},
		},
		{
			file: "Main.java",
			src: `public class Main {
    public static int add(int a, int b) {
        return a + b;
    }
}
`,
			wants: []string{"public class Main {", "public static int add(int a, int b) {"},
		},
		{
			file: "main.c",
			src: `#include <stdio.h>

int add(int a, int b) {
    return a + b;
}

int main(void) {
    printf("%d\n", add(1, 2));
    return 0;
}
`,
			wants: []string{"int add(int a, int b) {", "int main(void) {"},
		},
		{
			file: "util.php",
			src: `<?php

function add(int $a, int $b): int {
    return $a + $b;
}

class Point {
    public int $x;
}
`,
			wants: []string{"function add(int $a, int $b): int {", "class Point {"},
		},
		{
			file: "run.sh",
			src: `#!/usr/bin/env bash
set -euo pipefail

add() {
    echo $(( $1 + $2 ))
}

add 1 2
`,
			wants: []string{"add() {"},
		},
		{
			file: "math.ex",
			src: `defmodule Math do
  def add(a, b) do
    a + b
  end
end
`,
			wants: []string{"defmodule Math do", "def add(a, b) do"},
		},
		{
			file: "schema.sql",
			src: `CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE INDEX idx_users_name ON users(name);
`,
			wants: []string{"CREATE TABLE users (", "CREATE INDEX idx_users_name"},
		},
		{
			file: "point.rb",
			src: `class Point
  attr_accessor :x, :y

  def norm
    x + y
  end
end
`,
			wants: []string{"class Point", "def norm"},
		},
		{
			file: "Point.swift",
			src: `struct Point {
    var x: Int
    var y: Int

    func norm() -> Int {
        return x + y
    }
}

func add(_ a: Int, _ b: Int) -> Int {
    return a + b
}
`,
			wants: []string{"struct Point {", "func add(_ a: Int, _ b: Int) -> Int {"},
		},
		{
			file: "point.cpp",
			src: `#include <vector>

class Point {
public:
    int x;
    int norm();
};

int Point::norm() { return x; }
`,
			wants: []string{"class Point {", "int Point::norm()"},
		},
		{
			file: "Main.scala",
			src: `object Main {
  def add(a: Int, b: Int): Int = a + b

  def main(args: Array[String]): Unit =
    println(add(1, 2))
}
`,
			wants: []string{"object Main {", "def add(a: Int, b: Int): Int = a + b"},
		},
		{
			file: "Point.kt",
			src: `data class Point(val x: Int, val y: Int)

fun add(a: Int, b: Int): Int {
    return a + b
}
`,
			wants: []string{"data class Point(", "fun add(a: Int, b: Int): Int {"},
		},
		{
			file: "util.ml",
			src: `let add a b = a + b

let () =
  print_int (add 1 2);
  print_newline ()
`,
			wants: []string{"let add a b = a + b"},
		},
		{
			file: "Util.fs",
			src: `module Util

let add a b = a + b

type Point =
    { X: int
      Y: int }

let norm (p: Point) = p.X + p.Y
`,
			wants: []string{"let add a b = a + b", "type Point ="},
		},
		{
			file: "Point.cs",
			src: `using System;

namespace Demo {
    public class Point {
        public int X { get; set; }

        public int Norm() {
            return X;
        }
    }
}
`,
			wants: []string{"public class Point {", "public int Norm() {"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			ac := NewASTChunker(NewChunker(512, 50))
			chunks := ac.ChunkWithContext(tc.file, tc.src)

			if len(chunks) == 0 {
				t.Fatal("expected at least one chunk")
			}

			prefix := "File: " + tc.file
			combined := ""
			for i, c := range chunks {
				if !strings.HasPrefix(c.Content, prefix) {
					t.Errorf("chunk %d missing file context prefix", i)
				}
				if c.FilePath != tc.file {
					t.Errorf("chunk %d: expected file path %s, got %s", i, tc.file, c.FilePath)
				}
				if c.StartLine < 1 {
					t.Errorf("chunk %d: invalid start line %d", i, c.StartLine)
				}
				combined += strings.TrimPrefix(c.Content, prefix+"\n\n")
			}

			for _, want := range tc.wants {
				if !strings.Contains(combined, want) {
					t.Errorf("missing %q in chunked output", want)
				}
			}
		})
	}
}

func TestASTChunker_GoFile(t *testing.T) {
	src := `package main

import "fmt"

func hello() {
	fmt.Println("hello")
}

func world() {
	fmt.Println("world")
}

type Foo struct {
	Name string
}

func (f Foo) String() string {
	return f.Name
}
`
	ac := NewASTChunker(NewChunker(512, 50))
	chunks := ac.ChunkWithContext("main.go", src)

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	for i, c := range chunks {
		if !strings.HasPrefix(c.Content, "File: main.go") {
			t.Errorf("chunk %d missing file context prefix", i)
		}
		if c.FilePath != "main.go" {
			t.Errorf("chunk %d: expected file path main.go, got %s", i, c.FilePath)
		}
		if c.StartLine < 1 {
			t.Errorf("chunk %d: invalid start line %d", i, c.StartLine)
		}
	}

	combined := ""
	for _, c := range chunks {
		combined += strings.TrimPrefix(c.Content, "File: main.go\n\n")
	}
	if !strings.Contains(combined, "func hello()") {
		t.Error("missing hello function")
	}
	if !strings.Contains(combined, "func world()") {
		t.Error("missing world function")
	}
	if !strings.Contains(combined, "type Foo struct") {
		t.Error("missing Foo struct")
	}
}

func TestASTChunker_PythonFile(t *testing.T) {
	src := `import os

class Greeter:
    def __init__(self, name):
        self.name = name

    def greet(self):
        print(f"hello {self.name}")

def main():
    g = Greeter("world")
    g.greet()
`
	ac := NewASTChunker(NewChunker(512, 50))
	chunks := ac.ChunkWithContext("app.py", src)

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	combined := ""
	for _, c := range chunks {
		combined += strings.TrimPrefix(c.Content, "File: app.py\n\n")
	}
	if !strings.Contains(combined, "class Greeter") {
		t.Error("missing Greeter class")
	}
	if !strings.Contains(combined, "def main()") {
		t.Error("missing main function")
	}
}

func TestASTChunker_FallbackForUnsupportedExt(t *testing.T) {
	ac := NewASTChunker(NewChunker(512, 50))
	content := strings.Repeat("some yaml content\n", 50)
	chunks := ac.ChunkWithContext("config.yaml", content)

	if len(chunks) == 0 {
		t.Fatal("expected fallback chunks for unsupported extension")
	}
}

func TestASTChunker_OversizedFunction(t *testing.T) {
	var b strings.Builder
	b.WriteString("package main\n\n")
	b.WriteString("func tiny() {}\n\n")
	b.WriteString("func huge() {\n")
	for i := 0; i < 200; i++ {
		b.WriteString("\tfmt.Println(\"line\")\n")
	}
	b.WriteString("}\n")

	ac := NewASTChunker(NewChunker(64, 10))
	chunks := ac.ChunkWithContext("big.go", b.String())

	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks for oversized function, got %d", len(chunks))
	}
}

func TestASTChunker_EmptyContent(t *testing.T) {
	ac := NewASTChunker(NewChunker(512, 50))
	chunks := ac.ChunkWithContext("empty.go", "")
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks, got %d", len(chunks))
	}
}

func TestNewFileChunker_AST(t *testing.T) {
	logs := captureLogs(t)

	fc := NewFileChunker("ast", 512, 50)

	if _, ok := fc.(*ASTChunker); !ok {
		t.Errorf("NewFileChunker(\"ast\") = %T, want *ASTChunker", fc)
	}
	if got := logs(); strings.Contains(got, "chunking strategy") {
		t.Errorf("unexpected warning: %q", got)
	}
}

func TestASTChunker_VerbatimReconstruction(t *testing.T) {
	src := "package main\n\nimport \"fmt\"\n\nfunc tiny() {}\n\nfunc medium() {\n\tfor i := 0; i < 10; i++ {\n\t\tfmt.Println(i)\n\t}\n}\n\nfunc huge() {\n"
	for i := 0; i < 100; i++ {
		src += "\tfmt.Println(\"line\")\n"
	}
	src += "}\n"

	ac := NewASTChunker(NewChunker(64, 10))
	chunks := ac.ChunkWithContext("main.go", src)

	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}

	prefix := "File: main.go\n\n"
	var combined string
	for _, c := range chunks {
		combined += strings.TrimPrefix(c.Content, prefix)
	}

	if combined != src {
		t.Errorf("verbatim reconstruction failed\ngot length: %d\nwant length: %d", len(combined), len(src))
		for i := 0; i < len(src) && i < len(combined); i++ {
			if combined[i] != src[i] {
				t.Errorf("first diff at byte %d: got %q want %q", i, combined[i], src[i])
				break
			}
		}
	}
}

func TestASTChunker_NonWhitespaceSizeMetric(t *testing.T) {
	cumsum := buildNWSCumSum("  func hello()  {\n  }\n")
	nws := nwsInRange(cumsum, 0, len("  func hello()  {\n  }\n"))
	expected := len("funchello(){}")
	if nws != expected {
		t.Errorf("non-whitespace count: got %d, want %d", nws, expected)
	}
}

func TestASTChunker_RecursiveDescentNotFixedFallback(t *testing.T) {
	var b strings.Builder
	b.WriteString("package main\n\n")
	b.WriteString("func huge() {\n")
	for i := 0; i < 50; i++ {
		b.WriteString("\tx := 1\n")
	}
	b.WriteString("}\n")

	ac := NewASTChunker(NewChunker(32, 5))
	chunks := ac.ChunkWithContext("recursive.go", b.String())

	for _, c := range chunks {
		raw := strings.TrimPrefix(c.Content, "File: recursive.go\n\n")
		if strings.Contains(raw, "func huge()") && strings.Contains(raw, "x := 1") {
			continue
		}
		nws := 0
		for _, r := range raw {
			if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
				nws++
			}
		}
		if nws > ac.maxSize*2 {
			t.Errorf("chunk has %d non-whitespace chars, max is %d: likely fell back to fixed-size", nws, ac.maxSize)
		}
	}
}

func TestASTChunker_MergeAdjacentRanges(t *testing.T) {
	content := "aaaa    bbbb    cccc    dddd"
	cumsum := buildNWSCumSum(content)
	ac := &ASTChunker{maxSize: 10}

	ranges := []byteRange{
		{0, 4},   // "aaaa" nws=4
		{8, 12},  // "bbbb" nws=4
		{16, 20}, // "cccc" nws=4
		{24, 28}, // "dddd" nws=4
	}

	merged := ac.mergeAdjacentRanges(ranges, cumsum)
	if len(merged) != 2 {
		t.Fatalf("expected 2 merged ranges, got %d", len(merged))
	}
	if merged[0].start != 0 || merged[0].end != 12 {
		t.Errorf("first merged range: got {%d,%d}, want {0,12}", merged[0].start, merged[0].end)
	}
	if merged[1].start != 16 || merged[1].end != 28 {
		t.Errorf("second merged range: got {%d,%d}, want {16,28}", merged[1].start, merged[1].end)
	}
}

// goFuncs returns a Go file with n small functions.
func goFuncs(n int) string {
	var b strings.Builder
	b.WriteString("package main\n\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "func f%d() int {\n\treturn %d\n}\n\n", i, i)
	}
	return b.String()
}

// nestedGoBlocks returns a Go function whose body sits inside depth nested
// blocks. The innermost block holds enough statements to exceed any small
// chunk budget, so every enclosing node is oversized and chunkNodes has to
// descend through all of them.
func nestedGoBlocks(depth int) string {
	var b strings.Builder
	b.WriteString("package main\n\nfunc f() {\n")
	for i := 0; i < depth; i++ {
		b.WriteString("{\n")
	}
	for i := 0; i < 40; i++ {
		b.WriteString("\tx := 1\n")
	}
	for i := 0; i < depth; i++ {
		b.WriteString("}\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func TestASTChunker_ChunksCarryEmbedContent(t *testing.T) {
	tests := []struct {
		name       string
		size       int
		src        string
		wantChunks int
	}{
		{"single chunk", 512, "package main\n\nfunc main() {}\n", 1},
		{"multiple chunks", 16, goFuncs(8), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := NewASTChunker(NewChunker(tt.size, 2)).ChunkWithContext("main.go", tt.src)

			if len(chunks) < tt.wantChunks {
				t.Fatalf("got %d chunks, want at least %d", len(chunks), tt.wantChunks)
			}
			for i, c := range chunks {
				// The indexer sends EmbedContent to the embedder.
				if c.EmbedContent == "" {
					t.Errorf("chunk %d: EmbedContent is empty", i)
				}
				if c.EmbedContent != c.Content {
					t.Errorf("chunk %d: EmbedContent %q differs from Content %q", i, c.EmbedContent, c.Content)
				}
			}
		})
	}
}

func TestASTChunker_WhitespaceOnlyContent(t *testing.T) {
	ac := NewASTChunker(NewChunker(512, 50))
	if chunks := ac.ChunkWithContext("blank.go", "\n\n  \t\n"); len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for whitespace-only content, got %d", len(chunks))
	}
}

func TestASTChunker_DeepNestingFallsBackToFixed(t *testing.T) {
	src := nestedGoBlocks(maxASTDepth + 8)

	got := NewASTChunker(NewChunker(16, 2)).ChunkWithContext("deep.go", src)
	want := NewChunker(16, 2).ChunkWithContext("deep.go", src)

	if len(want) == 0 {
		t.Fatal("fixed chunker produced no chunks")
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected the fixed-size result once nesting exceeds %d levels\ngot  %d chunks\nwant %d chunks", maxASTDepth, len(got), len(want))
	}
}

func TestASTChunker_ModerateNestingStaysAST(t *testing.T) {
	src := nestedGoBlocks(4)

	got := NewASTChunker(NewChunker(16, 2)).ChunkWithContext("shallow.go", src)
	fixed := NewChunker(16, 2).ChunkWithContext("shallow.go", src)

	if reflect.DeepEqual(got, fixed) {
		t.Fatal("expected AST-aligned chunks, got the fixed-size result")
	}

	var combined strings.Builder
	for _, c := range got {
		combined.WriteString(strings.TrimPrefix(c.Content, "File: shallow.go\n\n"))
	}
	if combined.String() != src {
		t.Error("chunks do not reassemble into the original source")
	}
}

func TestASTChunker_TSXKeepsComponentsIntact(t *testing.T) {
	// The arrow function returning <li key={n}> is not valid in the plain
	// TypeScript grammar, which turns the whole file into a flat run of error
	// tokens; only the TSX grammar parses it into one node per component.
	component := func(name string) string {
		return "export function " + name + "({ items }: { items: number[] }) {\n" +
			"  const rows = items.map((n) => <li key={n}>{n}</li>);\n" +
			"  return <ul>{rows}</ul>;\n" +
			"}\n\n"
	}
	src := "import React from \"react\";\n\n" + component("First") + component("Second")

	// The import and the first component fit the budget together, but not
	// with the second component, so the split has to fall between the two.
	chunks := NewASTChunker(NewChunker(40, 2)).ChunkWithContext("app.tsx", src)

	if len(chunks) != 2 {
		t.Fatalf("got %d chunks, want 2", len(chunks))
	}
	prefix := "File: app.tsx\n\n"
	first := strings.TrimPrefix(chunks[0].Content, prefix)
	second := strings.TrimPrefix(chunks[1].Content, prefix)
	if !strings.HasPrefix(second, "export function Second") {
		t.Errorf("second chunk should start at the second component, got %q", second)
	}
	if !strings.Contains(first, "export function First") || !strings.HasSuffix(strings.TrimSpace(first), "}") {
		t.Errorf("first chunk should end with the first component, got %q", first)
	}
	if first+second != src {
		t.Error("chunks do not reassemble into the original source")
	}
}
