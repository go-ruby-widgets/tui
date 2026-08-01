// Copyright (c) 2026 the go-ruby-widgets/tui authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package tui_test

import (
	"fmt"

	"github.com/go-ruby-widgets/tui"
)

// Example mirrors the README: build a border-layout container with a titled
// header over a clickable body, render it to a cell grid, then route a click —
// every result a Ruby-shaped value.
func Example() {
	m := tui.NewModule()

	button := m.Button("OK")
	_ = button.OnClick("ok")

	root := m.Container()
	_ = root.SetHeaderHeight(1)
	_ = root.SetHeader(m.Label("Title"))
	_ = root.SetBody(button)

	_ = m.SetSize(root, 20, 5)

	grid, _ := m.RenderCells(root, 20, 5) // a Ruby Hash of decoded cells
	fmt.Println("cols:", grid["cols"], "rows:", grid["rows"])

	res, _ := m.Dispatch(button, map[string]any{"kind": "click"})
	fmt.Println("fired:", res["fired"], "repaint:", res["repaint"])

	// Output:
	// cols: 20 rows: 5
	// fired: [ok] repaint: true
}
