// Copyright (c) 2026 the go-ruby-widgets/tui authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package tui_test

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-ruby-widgets/tui"
)

// TestDashboardExample is the automated acceptance test for examples/dashboard.rb.
// It reproduces that demo through tui.Call -- the single dynamic-dispatch entry
// point an rbgo `method_missing` shim drives -- passing Ruby-shaped values (a
// []any Array, a map[string]any Hash, snake_case method names) exactly as the
// binding does, then asserts the rendered cell grid at cell precision. This
// proves the adapter end-to-end without a dependency on the interpreter.
func TestDashboardExample(t *testing.T) {
	m := tui.NewModule()

	call := func(recv any, method string, args ...any) any {
		t.Helper()
		v, err := tui.Call(recv, method, args...)
		if err != nil {
			t.Fatalf("Call %s: %v", method, err)
		}
		return v
	}

	items := []any{"Inbox", "Drafts", "Sent", "Archive", "Trash"}
	list := call(m, "list_box", items).(*tui.Widget)
	call(list, "on_select", "selection_changed")

	header := call(m, "label", "  Mailboxes").(*tui.Widget)
	call(header, "set_align", "left")

	status := call(m, "label", "1/5: Inbox").(*tui.Widget)

	root := call(m, "container").(*tui.Widget)
	call(root, "set_header_height", 1)
	call(root, "set_footer_height", 1)
	call(root, "set_header", header)
	call(root, "set_body", list)
	call(root, "set_footer", status)
	call(m, "set_size", root, 24, 8)

	// refresh keeps the footer bound to the list selection, mirroring the demo's
	// lambda: read the selected index back and rewrite the status label.
	refresh := func() {
		i := call(list, "selected").(int)
		call(status, "set_text",
			strconv.Itoa(i+1)+"/"+strconv.Itoa(len(items))+": "+items[i].(string))
	}

	// Two "Down" key presses drive the selection to the third row.
	for range 2 {
		call(m, "dispatch", list, map[string]any{"kind": "key_down", "code": "Down"})
		refresh()
	}

	if got := call(list, "selected").(int); got != 2 {
		t.Fatalf("selected = %d, want 2", got)
	}

	// Assert the decoded cell grid at cell precision.
	grid := call(m, "render_cells", root, 24, 8).(map[string]any)
	if grid["cols"] != 24 || grid["rows"] != 8 {
		t.Fatalf("grid dims = %vx%v, want 24x8", grid["cols"], grid["rows"])
	}
	rows := grid["text"].([]any)
	var joined strings.Builder
	for _, r := range rows {
		joined.WriteString(r.(string))
		joined.WriteByte('\n')
	}
	text := joined.String()
	for _, want := range []string{"Mailboxes", "Sent", "3/5: Sent"} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered frame missing %q; got:\n%s", want, text)
		}
	}

	// A cell Hash carries a "char" key (plus optional colors).
	cell := grid["cells"].([]any)[0].([]any)[0].(map[string]any)
	if _, ok := cell["char"]; !ok {
		t.Fatalf("cell missing char key: %v", cell)
	}
}

// TestDashboardExampleThroughRbgo runs examples/dashboard.rb through the actual
// rbgo interpreter when it is on PATH, asserting the demo's success line. It is
// skipped when rbgo is unavailable (e.g. in CI, which has no interpreter
// dependency), so the CI-safe proof above always runs.
func TestDashboardExampleThroughRbgo(t *testing.T) {
	rbgo, err := exec.LookPath("rbgo")
	if err != nil {
		t.Skip("rbgo not on PATH; skipping interpreter run (see TestDashboardExample)")
	}
	script, err := filepath.Abs(filepath.Join("examples", "dashboard.rb"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(rbgo, script).CombinedOutput()
	if err != nil {
		t.Fatalf("rbgo %s: %v\n%s", script, err, out)
	}
	if got := strings.TrimSpace(string(out)); !strings.HasPrefix(got, "OK ") {
		t.Fatalf("rbgo output = %q, want an OK line", got)
	}
}
