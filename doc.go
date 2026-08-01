// Copyright (c) 2026 the go-ruby-widgets/tui authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

// Package tui is the pure-Go, Ruby-runtime-independent core of the Ruby `tui`
// gem: a terminal-cell user-interface toolkit — widgets, layout containers,
// frame rendering and input routing — shaped so that
// github.com/go-embedded-ruby/ruby (rbgo) can bind it as `require "tui"`.
//
// It is a thin adapter over github.com/go-widgets/tui, the terminal-cell widget
// toolkit. It exposes that toolkit through Ruby-facing handles (Module, Widget)
// whose methods return Ruby-shaped values: a Hash (map[string]any), an Array
// ([]any) or a scalar. A single dynamic entry point, Call, dispatches a
// Ruby-style snake_case method name to the matching handle method and coerces
// the arguments, which is exactly what an rbgo binding drives from
// method_missing. Nothing here imports the Ruby runtime, so the package is
// equally usable as a standalone Go library — a sibling of
// go-ruby-opentype/opentype, go-ruby-regexp/regexp and go-ruby-erb/erb.
//
// # Handles
//
//   - Module is the package-level receiver: it constructs widgets and layout
//     containers, lays a tree out (SetSize), renders it (Render / RenderCells),
//     decodes an ANSI frame (DecodeCells) and routes input (Dispatch).
//   - Widget is an opaque handle over one widget or container. Its methods
//     mutate props (SetText, SetFraction, …), compose a tree (SetBody, AddTab,
//     SetLeft, …) and wire callbacks by id (OnClick, OnToggle, …).
//
// # Widgets and layouts
//
// Leaves: label, button, entry, check_button, list_box, progress_bar.
// Containers: container (a border layout — fixed header + footer bands around a
// filling body, plus floating overlays), h_split / v_split (draggable split
// panes) and notebook (a card layout — a tab strip selects the visible page).
//
// # The render seam
//
// A widget tree renders to a terminal cell grid, emitted as a self-contained
// ANSI stream. Render returns that stream as a String a Ruby host can print
// straight to a terminal. RenderCells (and DecodeCells over any ANSI string)
// return the decoded grid as a Ruby Hash — "cols", "rows", a "cells" Array of
// rows of {"char", "fg", "bg"} cell Hashes (a color is {"r","g","b"} or nil),
// and a "text" Array of per-row strings — so a test can assert the output at
// cell precision. The decode path is backed by go-widgets/tui's DecodeANSI.
//
// # Events
//
// Dispatch routes one input event (a Ruby Hash with "kind", "x", "y", "code"
// and modifier flags) into the tree, whose containers translate coordinates and
// deliver it to the right leaf. It returns the ids of the callbacks that fired
// and whether a repaint is warranted.
//
// # Usage from Go
//
//	m := tui.NewModule()
//	root := m.Container()
//	_ = root.SetHeaderHeight(1)
//	_ = root.SetHeader(m.Label("Title"))
//	_ = root.SetBody(m.Button("OK"))
//	_ = m.SetSize(root, 40, 10)
//	frame, _ := m.Render(root, 40, 10)         // an ANSI String
//	grid, _ := m.RenderCells(root, 40, 10)     // a Ruby Hash of cells
//	_ = frame
//	_ = grid
//
// # Usage from Ruby
//
// Under rbgo, `require "tui"` gives a Tui module whose snake_case methods are
// these operations, returning Ruby Hashes, Arrays and scalars:
//
//	require "tui"
//
//	root = Tui.container
//	root.set_header_height(1)
//	root.set_header(Tui.label("Title"))
//	root.set_body(Tui.button("OK"))
//	Tui.set_size(root, 40, 10)
//	frame = Tui.render(root, 40, 10)           # => String (ANSI)
//	fired = Tui.dispatch(root, {"kind" => "click", "x" => 1, "y" => 2})
//
// The `require "tui"` binding lives in rbgo (a thin method_missing shim over
// Call); it is pending in that repo.
package tui
