// Copyright (c) 2026 the go-ruby-widgets/tui authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package tui

import (
	"strings"
	"testing"
)

// --- construction ----------------------------------------------------------

func TestConstructorsAndKind(t *testing.T) {
	m := NewModule()
	cases := []struct {
		w    *Widget
		kind string
	}{
		{m.Label("hi"), "label"},
		{m.Button("go"), "button"},
		{m.Entry("x"), "entry"},
		{m.CheckButton("c", true), "check_button"},
		{m.ListBox([]any{"a", "b"}), "list_box"},
		{m.ProgressBar(), "progress_bar"},
		{m.Container(), "container"},
		{m.HSplit(), "h_split"},
		{m.VSplit(), "v_split"},
		{m.Notebook(), "notebook"},
	}
	for _, c := range cases {
		if c.w.Kind() != c.kind {
			t.Errorf("kind = %q, want %q", c.w.Kind(), c.kind)
		}
		if c.w.Underlying() == nil {
			t.Errorf("%s: nil underlying widget", c.kind)
		}
	}
}

func TestPackageLevelConstructors(t *testing.T) {
	widgets := []*Widget{
		Label("l"), Button("b"), Entry("e"), CheckButton("c", false),
		ListBox([]any{"x"}), ProgressBar(), Container(), HSplit(), VSplit(), Notebook(),
	}
	for i, w := range widgets {
		if w == nil || w.Underlying() == nil {
			t.Fatalf("package constructor %d returned an empty handle", i)
		}
	}
}

// --- mutation happy paths --------------------------------------------------

func TestTextMutators(t *testing.T) {
	m := NewModule()

	lbl := m.Label("a")
	if err := lbl.SetText("b"); err != nil {
		t.Fatal(err)
	}
	if got, _ := lbl.Text(); got != "b" {
		t.Errorf("label text = %q", got)
	}
	if err := lbl.SetAlign("center"); err != nil {
		t.Fatal(err)
	}
	if err := lbl.SetAlign("right"); err != nil {
		t.Fatal(err)
	}
	if err := lbl.SetAlign("weird"); err != nil { // default -> left
		t.Fatal(err)
	}

	ent := m.Entry("")
	if err := ent.SetText("hello"); err != nil {
		t.Fatal(err)
	}
	if got, _ := ent.Text(); got != "hello" {
		t.Errorf("entry text = %q", got)
	}
	if err := ent.SetPlaceholder("type…"); err != nil {
		t.Fatal(err)
	}
}

func TestLabelAndScalarMutators(t *testing.T) {
	m := NewModule()

	for _, w := range []*Widget{m.Button("x"), m.CheckButton("x", false), m.ProgressBar()} {
		if err := w.SetLabel("Y"); err != nil {
			t.Errorf("%s SetLabel: %v", w.Kind(), err)
		}
	}

	cb := m.CheckButton("c", false)
	if err := cb.SetChecked(true); err != nil {
		t.Fatal(err)
	}
	if v, _ := cb.Checked(); !v {
		t.Error("check button not checked")
	}

	pb := m.ProgressBar()
	if err := pb.SetFraction(2.0); err != nil { // clamped to 1
		t.Fatal(err)
	}
	if v, _ := pb.Fraction(); v != 1 {
		t.Errorf("fraction = %v, want 1", v)
	}
}

func TestListBoxMutators(t *testing.T) {
	m := NewModule()
	lb := m.ListBox([]any{"a", "b"})
	if err := lb.SetItems([]any{"x", "y", "z"}); err != nil {
		t.Fatal(err)
	}
	items, _ := lb.Items()
	if len(items) != 3 || items[2] != "z" {
		t.Errorf("items = %v", items)
	}
	if err := lb.SetSelected(2); err != nil {
		t.Fatal(err)
	}
	if v, _ := lb.Selected(); v != 2 {
		t.Errorf("selected = %d", v)
	}
	// A nil Array clears the list.
	if err := lb.SetItems(nil); err != nil {
		t.Fatal(err)
	}
	if items, _ := lb.Items(); len(items) != 0 {
		t.Errorf("items after nil = %v", items)
	}
}

func TestContainerComposition(t *testing.T) {
	m := NewModule()
	c := m.Container()
	if err := c.SetHeader(m.Label("h")); err != nil {
		t.Fatal(err)
	}
	if err := c.SetBody(m.Label("b")); err != nil {
		t.Fatal(err)
	}
	if err := c.SetFooter(m.Label("f")); err != nil {
		t.Fatal(err)
	}
	if err := c.SetHeaderHeight(1); err != nil {
		t.Fatal(err)
	}
	if err := c.SetFooterHeight(1); err != nil {
		t.Fatal(err)
	}
	if err := c.AddOverlay(m.Label("o")); err != nil {
		t.Fatal(err)
	}
	// A nil child is allowed (unset pane); exercises childWidget's nil path.
	if err := c.SetFooter(nil); err != nil {
		t.Fatal(err)
	}
}

func TestSplitComposition(t *testing.T) {
	m := NewModule()
	h := m.HSplit()
	if err := h.SetLeft(m.Label("l")); err != nil {
		t.Fatal(err)
	}
	if err := h.SetRight(m.Label("r")); err != nil {
		t.Fatal(err)
	}
	if err := h.SetLeftFraction(30); err != nil {
		t.Fatal(err)
	}

	v := m.VSplit()
	if err := v.SetTop(m.Label("t")); err != nil {
		t.Fatal(err)
	}
	if err := v.SetBottom(m.Label("b")); err != nil {
		t.Fatal(err)
	}
	if err := v.SetTopFraction(70); err != nil {
		t.Fatal(err)
	}
}

func TestNotebookComposition(t *testing.T) {
	m := NewModule()
	nb := m.Notebook()
	if err := nb.AddTab("One", m.Label("1")); err != nil {
		t.Fatal(err)
	}
	if err := nb.AddTab("Two", m.Label("2")); err != nil {
		t.Fatal(err)
	}
	if err := nb.SetActive(1); err != nil {
		t.Fatal(err)
	}
	if v, _ := nb.Active(); v != 1 {
		t.Errorf("active = %d", v)
	}
}

// --- wrong-kind error paths ------------------------------------------------

func TestWrongKindErrors(t *testing.T) {
	m := NewModule()
	lbl := m.Label("x")  // wrong kind for almost everything
	btn := m.Button("x") // wrong kind for text mutators
	checks := []struct {
		name string
		err  error
	}{
		{"SetText", btn.SetText("y")},
		{"SetAlign", btn.SetAlign("left")},
		{"SetPlaceholder", btn.SetPlaceholder("p")},
		{"SetLabel", lbl.SetLabel("y")},
		{"SetChecked", lbl.SetChecked(true)},
		{"SetFraction", lbl.SetFraction(0.5)},
		{"SetItems", lbl.SetItems(nil)},
		{"SetSelected", lbl.SetSelected(1)},
		{"SetHeader", lbl.SetHeader(nil)},
		{"SetBody", lbl.SetBody(nil)},
		{"SetFooter", lbl.SetFooter(nil)},
		{"SetHeaderHeight", lbl.SetHeaderHeight(1)},
		{"SetFooterHeight", lbl.SetFooterHeight(1)},
		{"AddOverlay", lbl.AddOverlay(nil)},
		{"SetLeft", lbl.SetLeft(nil)},
		{"SetRight", lbl.SetRight(nil)},
		{"SetLeftFraction", lbl.SetLeftFraction(1)},
		{"SetTop", lbl.SetTop(nil)},
		{"SetBottom", lbl.SetBottom(nil)},
		{"SetTopFraction", lbl.SetTopFraction(1)},
		{"AddTab", lbl.AddTab("t", nil)},
		{"SetActive", lbl.SetActive(1)},
		{"OnClick", lbl.OnClick("i")},
		{"OnToggle", lbl.OnToggle("i")},
		{"OnSelect", lbl.OnSelect("i")},
		{"OnChange", lbl.OnChange("i")},
		{"OnTabChanged", lbl.OnTabChanged("i")},
	}
	for _, c := range checks {
		if c.err == nil {
			t.Errorf("%s on wrong kind: want error, got nil", c.name)
		}
	}

	// Getters return the zero value plus an error on the wrong kind.
	if _, err := btn.Text(); err == nil {
		t.Error("Text wrong kind: want error")
	}
	if _, err := lbl.Checked(); err == nil {
		t.Error("Checked wrong kind: want error")
	}
	if _, err := lbl.Fraction(); err == nil {
		t.Error("Fraction wrong kind: want error")
	}
	if _, err := lbl.Items(); err == nil {
		t.Error("Items wrong kind: want error")
	}
	if _, err := lbl.Selected(); err == nil {
		t.Error("Selected wrong kind: want error")
	}
	if _, err := lbl.Active(); err == nil {
		t.Error("Active wrong kind: want error")
	}
}

// --- layout / render / bounds ----------------------------------------------

func TestSetSizeAndBounds(t *testing.T) {
	m := NewModule()
	lbl := m.Label("hi")
	if err := m.SetSize(lbl, 40, 10); err != nil {
		t.Fatal(err)
	}
	b, err := m.Bounds(lbl)
	if err != nil {
		t.Fatal(err)
	}
	if b["w"] != 40 || b["h"] != 10 || b["x"] != 0 || b["y"] != 0 {
		t.Errorf("bounds = %v", b)
	}
	// Non-positive dimensions fall back to the 80x24 default (clamp).
	if err := m.SetSize(lbl, 0, 0); err != nil {
		t.Fatal(err)
	}
	b, _ = m.Bounds(lbl)
	if b["w"] != 80 || b["h"] != 24 {
		t.Errorf("clamped bounds = %v", b)
	}
}

func TestRenderProducesANSIAndCells(t *testing.T) {
	m := NewModule()
	lbl := m.Label("HELLO")
	_ = lbl.SetAlign("center")

	ansi, err := m.Render(lbl, 20, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ansi, "\x1b[") {
		t.Error("render output is not an ANSI stream")
	}

	grid, err := m.RenderCells(lbl, 20, 3)
	if err != nil {
		t.Fatal(err)
	}
	if grid["cols"] != 20 || grid["rows"] != 3 {
		t.Errorf("grid dims = %v x %v", grid["cols"], grid["rows"])
	}
	text := grid["text"].([]any)
	joined := ""
	for _, r := range text {
		joined += r.(string)
	}
	if !strings.Contains(joined, "HELLO") {
		t.Errorf("rendered text missing label; got rows %v", text)
	}

	// A rendered frame carries set colors (background fill + text ink), so at
	// least one cell decodes to a color Hash — exercises colorToRuby's set path.
	sawColor := false
	for _, row := range grid["cells"].([]any) {
		for _, cell := range row.([]any) {
			c := cell.(map[string]any)
			if c["bg"] != nil {
				if bg, ok := c["bg"].(map[string]any); ok {
					if _, ok := bg["r"]; ok {
						sawColor = true
					}
				}
			}
		}
	}
	if !sawColor {
		t.Error("expected at least one colored cell in a rendered frame")
	}
}

func TestDecodeCellsPlainAndDefaults(t *testing.T) {
	m := NewModule()
	// A plain (color-free) stream: every cell decodes to the terminal default,
	// so fg/bg are nil — exercises colorToRuby's unset path and clamp defaults.
	grid := m.DecodeCells("hi", 0, 0)
	if grid["cols"] != 80 || grid["rows"] != 24 {
		t.Errorf("default dims = %v x %v", grid["cols"], grid["rows"])
	}
	first := grid["cells"].([]any)[0].([]any)[0].(map[string]any)
	if first["char"] != "h" {
		t.Errorf("first char = %v", first["char"])
	}
	if first["fg"] != nil || first["bg"] != nil {
		t.Errorf("plain stream should have default colors, got fg=%v bg=%v", first["fg"], first["bg"])
	}
}

// --- events ----------------------------------------------------------------

func TestDispatchFiresCallbacks(t *testing.T) {
	m := NewModule()

	btn := m.Button("OK")
	_ = btn.OnClick("clicked")
	res, err := m.Dispatch(btn, map[string]any{"kind": "click", "x": 1, "y": 1})
	if err != nil {
		t.Fatal(err)
	}
	if fired := res["fired"].([]any); len(fired) != 1 || fired[0] != "clicked" {
		t.Errorf("button fired = %v", fired)
	}
	if res["repaint"] != true {
		t.Error("repaint should be true")
	}

	cb := m.CheckButton("c", false)
	_ = cb.OnToggle("toggled")
	res, _ = m.Dispatch(cb, map[string]any{"kind": "click"})
	if fired := res["fired"].([]any); len(fired) != 1 || fired[0] != "toggled" {
		t.Errorf("check fired = %v", fired)
	}

	lb := m.ListBox([]any{"a", "b", "c"})
	_ = lb.OnSelect("selected")
	res, _ = m.Dispatch(lb, map[string]any{"kind": "key_down", "code": "Down"})
	if fired := res["fired"].([]any); len(fired) != 1 || fired[0] != "selected" {
		t.Errorf("list fired = %v", fired)
	}

	ent := m.Entry("")
	_ = ent.OnChange("changed")
	res, _ = m.Dispatch(ent, map[string]any{"kind": "char", "rune": "z"})
	if fired := res["fired"].([]any); len(fired) != 1 || fired[0] != "changed" {
		t.Errorf("entry fired = %v", fired)
	}

	nb := m.Notebook()
	_ = nb.AddTab("A", m.Label("1"))
	_ = nb.AddTab("B", m.Label("2"))
	_ = nb.OnTabChanged("tab")
	_ = m.SetSize(nb, 20, 10)
	res, _ = m.Dispatch(nb, map[string]any{"kind": "click", "x": 4, "y": 0})
	if fired := res["fired"].([]any); len(fired) != 1 || fired[0] != "tab" {
		t.Errorf("notebook fired = %v", fired)
	}
}

func TestEventKindsAndCodes(t *testing.T) {
	m := NewModule()
	root := m.Container()
	_ = root.SetBody(m.Label("x"))
	_ = m.SetSize(root, 20, 10)

	kinds := []string{
		"click", "key_down", "keydown", "key_up", "char",
		"mouse_drag", "mouse_up", "tick", "bogus", // bogus -> default branch
	}
	for _, k := range kinds {
		if _, err := m.Dispatch(root, map[string]any{"kind": k}); err != nil {
			t.Errorf("dispatch %q: %v", k, err)
		}
	}

	// eventCode source priority: code, then key, then rune, then none.
	ent := m.Entry("")
	for _, ev := range []map[string]any{
		{"kind": "char", "code": "a"},
		{"kind": "char", "key": "b"},
		{"kind": "char", "rune": "c"},
		{"kind": "char"}, // no code source -> empty string, no-op insert
	} {
		if _, err := m.Dispatch(ent, ev); err != nil {
			t.Fatal(err)
		}
	}

	// A nil event Hash routes as a zero key event without panicking.
	if _, err := m.Dispatch(root, nil); err != nil {
		t.Fatal(err)
	}
}

// --- nil-handle guards -----------------------------------------------------

func TestNilHandleGuards(t *testing.T) {
	m := NewModule()
	if err := m.SetSize(nil, 1, 1); err == nil {
		t.Error("SetSize(nil): want error")
	}
	if _, err := m.Bounds(nil); err == nil {
		t.Error("Bounds(nil): want error")
	}
	if _, err := m.Render(nil, 1, 1); err == nil {
		t.Error("Render(nil): want error")
	}
	if _, err := m.RenderCells(nil, 1, 1); err == nil {
		t.Error("RenderCells(nil): want error")
	}
	if _, err := m.Dispatch(nil, nil); err == nil {
		t.Error("Dispatch(nil): want error")
	}
}

// --- package-level operation wrappers --------------------------------------

func TestPackageLevelOps(t *testing.T) {
	root := Container()
	_ = root.SetBody(Label("x"))
	if err := SetSize(root, 10, 4); err != nil {
		t.Fatal(err)
	}
	if _, err := Bounds(root); err != nil {
		t.Fatal(err)
	}
	if _, err := Render(root, 10, 4); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderCells(root, 10, 4); err != nil {
		t.Fatal(err)
	}
	if g := DecodeCells("\x1b[38;2;1;2;3mX", 5, 1); g["cols"] != 5 {
		t.Errorf("DecodeCells cols = %v", g["cols"])
	}
	if _, err := Dispatch(root, map[string]any{"kind": "click"}); err != nil {
		t.Fatal(err)
	}
}

// --- reflective Call / Methods ---------------------------------------------

func TestCallDispatch(t *testing.T) {
	m := NewModule()

	// Constructor via Call returns a *Widget handle.
	v, err := Call(m, "label", "hi")
	if err != nil {
		t.Fatal(err)
	}
	lbl, ok := v.(*Widget)
	if !ok || lbl.Kind() != "label" {
		t.Fatalf("Call label = %#v", v)
	}

	// Getter with (val, error) return.
	got, err := Call(lbl, "text")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hi" {
		t.Errorf("call text = %v", got)
	}

	// Void method (returns error only) yields a nil result.
	if r, err := Call(lbl, "set_text", "bye"); err != nil || r != nil {
		t.Errorf("call set_text = (%v, %v)", r, err)
	}

	// Scalar-only return.
	if r, _ := Call(lbl, "kind"); r != "label" {
		t.Errorf("call kind = %v", r)
	}

	// A trailing method error surfaces through Call.
	if _, err := Call(lbl, "checked"); err == nil {
		t.Error("call checked on label: want error")
	}

	// Double-underscore method name exercises camelize's empty-segment skip.
	if _, err := Call(lbl, "set__text", "z"); err != nil {
		t.Errorf("call set__text: %v", err)
	}
}

func TestCallErrors(t *testing.T) {
	m := NewModule()
	if _, err := Call(nil, "label"); err == nil {
		t.Error("Call(nil recv): want error")
	}
	if _, err := Call(m, "nope"); err == nil {
		t.Error("Call unknown method: want error")
	}
	lbl := m.Label("x")
	if _, err := Call(lbl, "kind", "extra"); err == nil {
		t.Error("Call too many args: want error")
	}
	if _, err := Call(m, "set_active", "x"); err == nil {
		// set_active is not a Module method; unknown method also errors, which
		// is fine — but assert via a real coercion failure below instead.
		_ = err
	}
}

func TestCallCoercion(t *testing.T) {
	m := NewModule()

	// String case: nil -> "", non-string -> Sprint.
	lbl := m.Label("x")
	if _, err := Call(lbl, "set_text", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(lbl, "set_text", 42); err != nil {
		t.Fatal(err)
	}
	if got, _ := lbl.Text(); got != "42" {
		t.Errorf("coerced text = %q", got)
	}

	// Int case: float64 and int64 coerce in; a string fails.
	nb := m.Notebook()
	_ = nb.AddTab("a", m.Label("1"))
	_ = nb.AddTab("b", m.Label("2"))
	if _, err := Call(nb, "set_active", 1.0); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(nb, "set_active", int64(0)); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(nb, "set_active", "x"); err == nil {
		t.Error("set_active with string: want error")
	}

	// Float64 case: int coerces in; a string fails.
	pb := m.ProgressBar()
	if _, err := Call(pb, "set_fraction", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(pb, "set_fraction", "x"); err == nil {
		t.Error("set_fraction with string: want error")
	}

	// Bool case: nil -> false; a non-bool is truthy.
	cb := m.CheckButton("c", false)
	if _, err := Call(cb, "set_checked", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(cb, "set_checked", 1); err != nil {
		t.Fatal(err)
	}
	if v, _ := cb.Checked(); !v {
		t.Error("truthy int should have checked the box")
	}

	// Slice case: nil -> empty; assignable Array passes; a scalar fails.
	lb := m.ListBox([]any{"a"})
	if _, err := Call(lb, "set_items", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(lb, "set_items", []any{"p", "q"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(lb, "set_items", 5); err == nil {
		t.Error("set_items with scalar: want error")
	}

	// Map case: nil -> empty; assignable Hash passes; a scalar fails.
	root := m.Container()
	if _, err := Call(m, "dispatch", root, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(m, "dispatch", root, map[string]any{"kind": "click"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Call(m, "dispatch", root, 5); err == nil {
		t.Error("dispatch with scalar Hash: want error")
	}

	// Pointer/default case: a nil handle where a *Widget is required fails.
	if _, err := Call(root, "set_body", nil); err == nil {
		t.Error("set_body with nil handle: want error")
	}
	// A real handle assigns straight through.
	if _, err := Call(root, "set_body", m.Label("b")); err != nil {
		t.Fatal(err)
	}
}

func TestMethods(t *testing.T) {
	m := NewModule()
	names := Methods(m)
	if !sortedContains(names, "render") || !sortedContains(names, "set_size") {
		t.Errorf("module methods missing entries: %v", names)
	}
	if !isSorted(names) {
		t.Error("module methods not sorted")
	}

	w := m.Label("x")
	wnames := Methods(w)
	if !sortedContains(wnames, "set_text") || !sortedContains(wnames, "on_click") {
		t.Errorf("widget methods missing entries: %v", wnames)
	}
	if !isSorted(wnames) {
		t.Error("widget methods not sorted")
	}
	// Round-trip: every snake name camelizes to a real method.
	for _, n := range wnames {
		if camelize(n) == "" {
			t.Errorf("snake name %q camelized to empty", n)
		}
	}
}

// --- small helper coverage -------------------------------------------------

func TestValueHelpers(t *testing.T) {
	// truthy
	if truthy(nil) || !truthy(true) || truthy(false) || !truthy(7) {
		t.Error("truthy semantics wrong")
	}
	// toInt
	for _, tc := range []struct {
		in any
		ok bool
	}{{1, true}, {int64(2), true}, {3.0, true}, {"x", false}} {
		if _, ok := toInt(tc.in); ok != tc.ok {
			t.Errorf("toInt(%v) ok = %v", tc.in, ok)
		}
	}
	// toIntOr defaults to 0 on a non-number.
	if toIntOr("x") != 0 || toIntOr(5) != 5 {
		t.Error("toIntOr wrong")
	}
	// toFloat
	for _, tc := range []struct {
		in any
		ok bool
	}{{1.5, true}, {2, true}, {int64(3), true}, {"x", false}} {
		if _, ok := toFloat(tc.in); ok != tc.ok {
			t.Errorf("toFloat(%v) ok = %v", tc.in, ok)
		}
	}
	// stringList
	if got := stringList([]any{"a", 2}); len(got) != 2 || got[1] != "2" {
		t.Errorf("stringList = %v", got)
	}
	// childWidget nil path.
	if childWidget(nil) != nil {
		t.Error("childWidget(nil) should be nil")
	}
	// snakeize handles acronym-adjacent capitals without stray underscores.
	if snakeize("HSplit") != "h_split" {
		t.Errorf("snakeize(HSplit) = %q", snakeize("HSplit"))
	}
}

// --- test-local helpers ----------------------------------------------------

func sortedContains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}

func isSorted(s []string) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1] > s[i] {
			return false
		}
	}
	return true
}
