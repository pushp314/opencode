package parse

import (
	"testing"
)

func TestDetect(t *testing.T) {
	cases := []struct {
		path string
		name string
		ok   bool
	}{
		{"main.go", "go", true},
		{"app.ts", "typescript", true},
		{"app.tsx", "typescript", true},
		{"index.js", "javascript", true},
		{"file.jsx", "javascript", true},
		{"file.mjs", "javascript", true},
		{"file.cjs", "javascript", true},
		{"util.py", "python", true},
		{"lib.rs", "rust", true},
		{"Main.java", "java", true},
		{"util.c", "c", true},
		{"util.h", "c", true},
		{"util.cc", "cpp", true},
		{"util.cpp", "cpp", true},
		{"util.cxx", "cpp", true},
		{"util.hpp", "cpp", true},
		{"README.md", "", false},
		{"data.json", "", false},
		{"file.swift", "", false},
		{"Makefile", "", false},
	}
	for _, c := range cases {
		lang, ok := Detect(c.path)
		if ok != c.ok {
			t.Errorf("Detect(%q) ok = %t, want %t", c.path, ok, c.ok)
			continue
		}
		if ok && lang.Name() != c.name {
			t.Errorf("Detect(%q) name = %q, want %q", c.path, lang.Name(), c.name)
		}
	}
}

func TestExtract(t *testing.T) {
	cases := []struct {
		name  string
		path  string
		code  string
		want  []Symbol
		error bool
	}{
		{
			name: "go",
			path: "main.go",
			code: `package main

import "fmt"

func Foo(a, b int) int { return a + b }

func (r *Receiver) Bar() {}

type Point struct{ X, Y int }

type Handler interface{ Run() }

type Alias = string

const Max = 100

var Count = 0
`,
			want: []Symbol{
				{Name: "Foo", Kind: Function},
				{Name: "Bar", Kind: Method},
				{Name: "Point", Kind: Struct},
				{Name: "Handler", Kind: Interface},
				{Name: "Alias", Kind: Type},
				{Name: "Max", Kind: Const},
				{Name: "Count", Kind: Var},
			},
		},
		{
			name: "typescript",
			path: "app.ts",
			code: `export function foo(a: number): void {}

export class A {
  m(): void {}
}

interface I {
  p: string
  run(x: number): void
}

enum E { X, Y }

type T = string

const c = 1
`,
			want: []Symbol{
				{Name: "foo", Kind: Function},
				{Name: "A", Kind: Class},
				{Name: "m", Kind: Method},
				{Name: "I", Kind: Interface},
				{Name: "run", Kind: Method},
				{Name: "E", Kind: Enum},
				{Name: "T", Kind: Type},
				{Name: "c", Kind: Var},
			},
		},
		{
			name: "javascript",
			path: "app.js",
			code: `function foo(a) { return a }

class A { m() {} }

const c = 1
let v = 2
var w = 3
`,
			want: []Symbol{
				{Name: "foo", Kind: Function},
				{Name: "A", Kind: Class},
				{Name: "m", Kind: Method},
				{Name: "c", Kind: Var},
				{Name: "v", Kind: Var},
				{Name: "w", Kind: Var},
			},
		},
		{
			name: "python",
			path: "app.py",
			code: `import os

def foo(a, b=1):
    return a + b

async def fetch():
    pass

@decorator
def decorated():
    pass

class Bar:
    def method(self):
        pass
`,
			want: []Symbol{
				{Name: "foo", Kind: Function},
				{Name: "fetch", Kind: Function},
				{Name: "decorated", Kind: Function},
				{Name: "Bar", Kind: Class},
				{Name: "method", Kind: Function},
			},
		},
		{
			name: "rust",
			path: "lib.rs",
			code: `fn foo(a: i32) -> i32 { a }

struct S { x: i32 }

enum E { A, B }

trait T { fn m(&self); }

impl S {
  fn build() -> S { S { x: 0 } }
}

type Alias = i32;

const MAX: i32 = 1;

static COUNT: i32 = 0;
`,
			want: []Symbol{
				{Name: "foo", Kind: Function},
				{Name: "S", Kind: Struct},
				{Name: "E", Kind: Enum},
				{Name: "T", Kind: Interface},
				{Name: "build", Kind: Function},
				{Name: "Alias", Kind: Type},
				{Name: "MAX", Kind: Const},
				{Name: "COUNT", Kind: Var},
			},
		},
		{
			name: "java",
			path: "Foo.java",
			code: `package p;

public class Foo {
  int field;
  public Foo() {}
  public void method(int x) {}
}

interface Iface { void run(); }

enum Color { RED, GREEN }

record Point(int x, int y) {}
`,
			want: []Symbol{
				{Name: "Foo", Kind: Class},
				{Name: "Foo", Kind: Method},
				{Name: "method", Kind: Method},
				{Name: "Iface", Kind: Interface},
				{Name: "run", Kind: Method},
				{Name: "Color", Kind: Enum},
				{Name: "Point", Kind: Class},
			},
		},
		{
			name: "c",
			path: "util.c",
			code: `#include <stdio.h>

int global_var;

int foo(int a) { return a; }

int *pointer_ret(void) { return 0; }

struct Point { int x; };

typedef struct Point Point;

enum Color { RED };

union Value { int a; };
`,
			want: []Symbol{
				{Name: "foo", Kind: Function},
				{Name: "pointer_ret", Kind: Function},
				{Name: "Point", Kind: Struct},
				{Name: "Point", Kind: Type},
				{Name: "Color", Kind: Enum},
				{Name: "Value", Kind: Type},
			},
		},
		{
			name: "cpp",
			path: "util.cpp",
			code: `int foo(int a) { return a; }

class A { public: void m(); };

void A::m() {}

struct B { int x; };

enum C { X };

union U { int a; };

typedef int MyInt;
`,
			want: []Symbol{
				{Name: "foo", Kind: Function},
				{Name: "A", Kind: Class},
				{Name: "m", Kind: Function},
				{Name: "B", Kind: Struct},
				{Name: "C", Kind: Enum},
				{Name: "U", Kind: Type},
				{Name: "MyInt", Kind: Type},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lang, ok := Detect(c.path)
			if !ok {
				t.Fatalf("Detect(%q) returned no language", c.path)
			}
			result := lang.Parse([]byte(c.code))
			if result.HasErrors != c.error {
				t.Errorf("HasErrors = %t, want %t", result.HasErrors, c.error)
			}
			if len(result.Symbols) != len(c.want) {
				t.Fatalf("got %d symbols %v, want %d %v", len(result.Symbols), symbolNames(result.Symbols), len(c.want), symbolNames(c.want))
			}
			for i := range c.want {
				got, want := result.Symbols[i], c.want[i]
				if got.Name != want.Name || got.Kind != want.Kind {
					t.Errorf("symbol %d = %s/%s, want %s/%s", i, got.Name, got.Kind, want.Name, want.Kind)
				}
				if got.EndByte <= got.StartByte {
					t.Errorf("symbol %s has empty range %d-%d", got.Name, got.StartByte, got.EndByte)
				}
				if got.EndLine < got.StartLine || got.StartLine == 0 {
					t.Errorf("symbol %s has invalid lines %d-%d", got.Name, got.StartLine, got.EndLine)
				}
			}
		})
	}
}

func TestMalformedInputDoesNotPanic(t *testing.T) {
	paths := []string{"a.go", "a.ts", "a.js", "a.py", "a.rs", "a.java", "a.c", "a.cpp"}
	garbage := []string{
		"",
		"\x00\x01\x02",
		"func (((((((((((((((((((",
		"class class class }}}}}}}}",
		"\xff\xfe\xfd",
		"===== <<<<<< >>>>>",
	}
	for _, path := range paths {
		for _, input := range garbage {
			lang, ok := Detect(path)
			if !ok {
				t.Fatalf("Detect(%q) returned no language", path)
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Detect(%q) panicked on %q: %v", path, input, r)
					}
				}()
				lang.Parse([]byte(input))
			}()
		}
	}
}

func TestSymbolRangesAreDeterministic(t *testing.T) {
	lang, ok := Detect("main.go")
	if !ok {
		t.Fatal("no go language")
	}
	code := []byte("package main\n\nfunc Foo() {}\n\ntype Bar struct{}\n")
	first := lang.Parse(code)
	second := lang.Parse(code)
	if len(first.Symbols) != len(second.Symbols) {
		t.Fatalf("symbol counts differ: %d vs %d", len(first.Symbols), len(second.Symbols))
	}
	for i := range first.Symbols {
		if first.Symbols[i] != second.Symbols[i] {
			t.Errorf("symbol %d differs: %+v vs %+v", i, first.Symbols[i], second.Symbols[i])
		}
	}
}

func symbolNames(symbols []Symbol) []string {
	names := make([]string, 0, len(symbols))
	for _, s := range symbols {
		names = append(names, string(s.Kind)+":"+s.Name)
	}
	return names
}

func TestExtractImports(t *testing.T) {
	cases := []struct {
		name  string
		path  string
		code  string
		want  []Import
		error bool
	}{
		{
			name: "go",
			path: "main.go",
			code: `package main

import (
	"fmt"
	sub "math/rand"
	_ "net/http"
)

import "strings"
`,
			want: []Import{
				{Kind: KindImport, Spec: "fmt"},
				{Kind: KindImport, Spec: "math/rand"},
				{Kind: KindImport, Spec: "net/http"},
				{Kind: KindImport, Spec: "strings"},
			},
		},
		{
			name: "typescript",
			path: "app.ts",
			code: `import { a } from "./a"
import type { B } from "./b"
export { c } from "./c"
const d = require("./d")
`,
			want: []Import{
				{Kind: KindImport, Spec: "./a"},
				{Kind: KindImport, Spec: "./b"},
				{Kind: KindExport, Spec: "./c"},
				{Kind: KindRequire, Spec: "./d"},
			},
		},
		{
			name: "javascript",
			path: "app.js",
			code: `import x from "x"
const y = require("y")
const n = notRequire("n")
export { z } from "z"
`,
			want: []Import{
				{Kind: KindImport, Spec: "x"},
				{Kind: KindRequire, Spec: "y"},
				{Kind: KindExport, Spec: "z"},
			},
		},
		{
			name: "python",
			path: "app.py",
			code: `import os
import os.path
from collections import OrderedDict
from . import local
from .util import a
from ..pkg.helper import b
`,
			want: []Import{
				{Kind: KindImport, Spec: "os"},
				{Kind: KindImport, Spec: "os.path"},
				{Kind: KindImport, Spec: "collections"},
				{Kind: KindImport, Spec: ".util"},
				{Kind: KindImport, Spec: "..pkg.helper"},
			},
		},
		{
			name: "rust",
			path: "lib.rs",
			code: `use std::collections::HashMap;
use crate::models::{User, Post};
use super::util;
`,
			want: []Import{
				{Kind: KindUse, Spec: "std::collections::HashMap"},
				{Kind: KindUse, Spec: "crate::models::{User, Post}"},
				{Kind: KindUse, Spec: "super::util"},
			},
		},
		{
			name: "java",
			path: "Foo.java",
			code: `package p;

import java.util.List;
import static p.Constants.MAX;
`,
			want: []Import{
				{Kind: KindImport, Spec: "java.util.List"},
				{Kind: KindImport, Spec: "p.Constants.MAX"},
			},
		},
		{
			name: "c",
			path: "util.c",
			code: `#include <stdio.h>
#include "local.h"
`,
			want: []Import{
				{Kind: KindInclude, Spec: "stdio.h"},
				{Kind: KindInclude, Spec: "local.h"},
			},
		},
		{
			name: "cpp",
			path: "util.cpp",
			code: `#include <vector>
#include "../util.hpp"
`,
			want: []Import{
				{Kind: KindInclude, Spec: "vector"},
				{Kind: KindInclude, Spec: "../util.hpp"},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lang, ok := Detect(c.path)
			if !ok {
				t.Fatalf("Detect(%q) returned no language", c.path)
			}
			result := lang.Parse([]byte(c.code))
			if result.HasErrors != c.error {
				t.Errorf("HasErrors = %t, want %t", result.HasErrors, c.error)
			}
			got := result.Imports
			// Loose match: allow relative python "" to be skipped.
			want := c.want
			if len(got) != len(want) {
				t.Fatalf("got %d imports %v, want %d %v", len(got), importSpecs(got), len(want), importSpecs(want))
			}
			for i := range want {
				if got[i].Kind != want[i].Kind || got[i].Spec != want[i].Spec {
					t.Errorf("import %d = %s/%s, want %s/%s", i, got[i].Kind, got[i].Spec, want[i].Kind, want[i].Spec)
				}
				if got[i].EndByte <= got[i].StartByte {
					t.Errorf("import %s has empty range %d-%d", got[i].Spec, got[i].StartByte, got[i].EndByte)
				}
				if got[i].EndLine < got[i].StartLine || got[i].StartLine == 0 {
					t.Errorf("import %s has invalid lines %d-%d", got[i].Spec, got[i].StartLine, got[i].EndLine)
				}
			}
		})
	}
}

func importSpecs(imports []Import) []string {
	specs := make([]string, 0, len(imports))
	for _, i := range imports {
		specs = append(specs, string(i.Kind)+":"+i.Spec)
	}
	return specs
}
