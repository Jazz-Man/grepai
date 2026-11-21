package zig

//#include "tree_sitter/parser.h"
//TSLanguage *tree_sitter_zig();
import "C"
import (
	"unsafe"

	sitter "github.com/smacker/go-tree-sitter"
)

// GetLanguage returns the tree-sitter grammar for Zig.
// Grammar source: https://github.com/tree-sitter-grammars/tree-sitter-zig (MIT license)
// Vendored at tag v1.1.2 (commit b670c8df): its LANGUAGE_VERSION 14 matches the
// ABI of the smacker/go-tree-sitter runtime. Do not upgrade the vendored files
// past ABI 14 without also upgrading the runtime.
func GetLanguage() *sitter.Language {
	ptr := unsafe.Pointer(C.tree_sitter_zig())
	return sitter.NewLanguage(ptr)
}
