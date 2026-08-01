// Copyright (c) 2026 the go-ruby-widgets/tui authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file at the root of this repository.

package tui

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/go-widgets/toolkit"
	wtui "github.com/go-widgets/tui"
)

// Module is the package-level Ruby receiver: the `Tui` module under rbgo. Its
// methods construct terminal-cell widgets and layout containers, lay them out,
// render them to a frame (an ANSI string and a decoded cell grid), and route
// input events — all returning Ruby-shaped values (a Hash is a
// map[string]any, an Array a []any, and scalars stay scalars).
//
// A Module carries the per-session callback log that [Module.Dispatch] drains,
// so it is stateful and NOT safe for concurrent use: give each terminal
// session its own Module (the package-level convenience functions share one
// default Module and are likewise single-threaded).
type Module struct {
	// fired accumulates the callback ids that fire during a single Dispatch.
	// Dispatch clears it before routing and copies it out afterwards.
	fired []any
}

// NewModule returns a fresh Module with an empty callback log. The
// package-level convenience functions delegate to a shared default Module.
func NewModule() *Module { return &Module{} }

var mod = NewModule()

// fire records that the callback registered under id ran. Wired closures call
// it; Dispatch reads the accumulated ids back out.
func (m *Module) fire(id string) { m.fired = append(m.fired, id) }

// Widget is a Ruby-facing opaque handle over one go-widgets/tui widget or
// layout container. Kind names the concrete widget ("label", "button",
// "container", …) so a caller can branch on it, and every mutator checks it so
// a method that does not apply to the handle's kind returns a clear error
// rather than mis-typing. A Widget belongs to the Module that built it, which
// is where its callbacks log.
type Widget struct {
	w    toolkit.Widget
	kind string
	mod  *Module
}

// wrap builds a handle bound to this Module.
func (m *Module) wrap(kind string, w toolkit.Widget) *Widget {
	return &Widget{w: w, kind: kind, mod: m}
}

// Kind returns the handle's widget-kind name (e.g. "label", "notebook").
func (w *Widget) Kind() string { return w.kind }

// Underlying returns the wrapped go-widgets/tui widget as a toolkit.Widget, for
// Go callers that want to reach past the adapter. Ruby callers never need it.
func (w *Widget) Underlying() toolkit.Widget { return w.w }

// --- construction (Module methods) -----------------------------------------

// Label builds a static, single-line text widget.
func (m *Module) Label(text string) *Widget { return m.wrap("label", wtui.NewLabel(text)) }

// Button builds a clickable action widget. Wire a handler with [Widget.OnClick].
func (m *Module) Button(label string) *Widget {
	return m.wrap("button", wtui.NewButton(label, nil))
}

// Entry builds a single-line text input seeded with initial text.
func (m *Module) Entry(initial string) *Widget { return m.wrap("entry", wtui.NewEntry(initial)) }

// CheckButton builds a boolean toggle with the given label and initial state.
func (m *Module) CheckButton(label string, checked bool) *Widget {
	return m.wrap("check_button", wtui.NewCheckButton(label, checked))
}

// ListBox builds a single-selection list over the given items (a Ruby Array of
// strings).
func (m *Module) ListBox(items []any) *Widget {
	return m.wrap("list_box", wtui.NewListBox(stringList(items)))
}

// ProgressBar builds an empty horizontal progress indicator.
func (m *Module) ProgressBar() *Widget { return m.wrap("progress_bar", wtui.NewProgressBar()) }

// Container builds a border-layout container: a fixed-height header band at the
// top, a fixed-height footer band at the bottom, a body filling the middle, and
// floating overlays. Compose it with SetHeader/SetBody/SetFooter/AddOverlay.
func (m *Module) Container() *Widget { return m.wrap("container", &wtui.VBox{}) }

// HSplit builds a horizontal split container (a left pane and a right pane
// separated by a draggable grip), left pane 50% by default.
func (m *Module) HSplit() *Widget { return m.wrap("h_split", &wtui.HSplit{LeftFrac: 50}) }

// VSplit builds a vertical split container (a top pane over a bottom pane),
// top pane 50% by default.
func (m *Module) VSplit() *Widget { return m.wrap("v_split", &wtui.VSplit{TopFrac: 50}) }

// Notebook builds a card-layout (tabbed) container: a tab strip selects which
// child page fills the body. Add pages with [Widget.AddTab].
func (m *Module) Notebook() *Widget { return m.wrap("notebook", wtui.NewNotebook()) }

// --- mutation / composition (Widget methods) -------------------------------

// SetText sets the text of a label or entry.
func (w *Widget) SetText(s string) error {
	switch t := w.w.(type) {
	case *wtui.Label:
		t.Text = s
	case *wtui.Entry:
		t.Text = s
		t.Cursor = len([]rune(s))
	default:
		return kindErr(w, "label or entry")
	}
	return nil
}

// Text returns the text of a label or entry.
func (w *Widget) Text() (string, error) {
	switch t := w.w.(type) {
	case *wtui.Label:
		return t.Text, nil
	case *wtui.Entry:
		return t.Text, nil
	default:
		return "", kindErr(w, "label or entry")
	}
}

// SetAlign sets a label's horizontal alignment ("left", "center" or "right";
// any other value is treated as "left").
func (w *Widget) SetAlign(align string) error {
	l, ok := w.w.(*wtui.Label)
	if !ok {
		return kindErr(w, "label")
	}
	switch strings.ToLower(align) {
	case "center", "centre":
		l.Align = wtui.AlignCenter
	case "right":
		l.Align = wtui.AlignRight
	default:
		l.Align = wtui.AlignLeft
	}
	return nil
}

// SetPlaceholder sets an entry's muted placeholder text.
func (w *Widget) SetPlaceholder(s string) error {
	e, ok := w.w.(*wtui.Entry)
	if !ok {
		return kindErr(w, "entry")
	}
	e.Placeholder = s
	return nil
}

// SetLabel sets the caption of a button, check button or progress bar.
func (w *Widget) SetLabel(s string) error {
	switch t := w.w.(type) {
	case *wtui.Button:
		t.Label = s
	case *wtui.CheckButton:
		t.Label = s
	case *wtui.ProgressBar:
		t.Label = s
	default:
		return kindErr(w, "button, check_button or progress_bar")
	}
	return nil
}

// SetChecked sets a check button's state.
func (w *Widget) SetChecked(v bool) error {
	c, ok := w.w.(*wtui.CheckButton)
	if !ok {
		return kindErr(w, "check_button")
	}
	c.Checked = v
	return nil
}

// Checked reports a check button's state.
func (w *Widget) Checked() (bool, error) {
	c, ok := w.w.(*wtui.CheckButton)
	if !ok {
		return false, kindErr(w, "check_button")
	}
	return c.Checked, nil
}

// SetFraction sets a progress bar's fill in [0,1] (clamped).
func (w *Widget) SetFraction(f float64) error {
	p, ok := w.w.(*wtui.ProgressBar)
	if !ok {
		return kindErr(w, "progress_bar")
	}
	p.SetFraction(f)
	return nil
}

// Fraction returns a progress bar's current fill.
func (w *Widget) Fraction() (float64, error) {
	p, ok := w.w.(*wtui.ProgressBar)
	if !ok {
		return 0, kindErr(w, "progress_bar")
	}
	return p.Fraction, nil
}

// SetItems replaces a list box's items (a Ruby Array of strings) and resets the
// selection to the first row.
func (w *Widget) SetItems(items []any) error {
	l, ok := w.w.(*wtui.ListBox)
	if !ok {
		return kindErr(w, "list_box")
	}
	l.Items = stringList(items)
	l.Selected = 0
	return nil
}

// Items returns a list box's items as a Ruby Array of strings.
func (w *Widget) Items() ([]any, error) {
	l, ok := w.w.(*wtui.ListBox)
	if !ok {
		return nil, kindErr(w, "list_box")
	}
	out := make([]any, len(l.Items))
	for i, it := range l.Items {
		out[i] = it
	}
	return out, nil
}

// SetSelected sets a list box's selected row index.
func (w *Widget) SetSelected(i int) error {
	l, ok := w.w.(*wtui.ListBox)
	if !ok {
		return kindErr(w, "list_box")
	}
	l.Selected = i
	return nil
}

// Selected returns a list box's selected row index.
func (w *Widget) Selected() (int, error) {
	l, ok := w.w.(*wtui.ListBox)
	if !ok {
		return 0, kindErr(w, "list_box")
	}
	return l.Selected, nil
}

// SetHeader sets a container's fixed header child.
func (w *Widget) SetHeader(child *Widget) error {
	c, ok := w.w.(*wtui.VBox)
	if !ok {
		return kindErr(w, "container")
	}
	c.Header = childWidget(child)
	return nil
}

// SetBody sets a container's body child (the middle, filling area).
func (w *Widget) SetBody(child *Widget) error {
	c, ok := w.w.(*wtui.VBox)
	if !ok {
		return kindErr(w, "container")
	}
	c.Body = childWidget(child)
	return nil
}

// SetFooter sets a container's fixed footer child.
func (w *Widget) SetFooter(child *Widget) error {
	c, ok := w.w.(*wtui.VBox)
	if !ok {
		return kindErr(w, "container")
	}
	c.Footer = childWidget(child)
	return nil
}

// SetHeaderHeight sets a container's header band height in rows.
func (w *Widget) SetHeaderHeight(n int) error {
	c, ok := w.w.(*wtui.VBox)
	if !ok {
		return kindErr(w, "container")
	}
	c.HeaderH = n
	return nil
}

// SetFooterHeight sets a container's footer band height in rows.
func (w *Widget) SetFooterHeight(n int) error {
	c, ok := w.w.(*wtui.VBox)
	if !ok {
		return kindErr(w, "container")
	}
	c.FooterH = n
	return nil
}

// AddOverlay floats a child on top of a container's body.
func (w *Widget) AddOverlay(child *Widget) error {
	c, ok := w.w.(*wtui.VBox)
	if !ok {
		return kindErr(w, "container")
	}
	c.Overlays = append(c.Overlays, childWidget(child))
	return nil
}

// SetLeft sets a horizontal split's left pane.
func (w *Widget) SetLeft(child *Widget) error {
	s, ok := w.w.(*wtui.HSplit)
	if !ok {
		return kindErr(w, "h_split")
	}
	s.Left = childWidget(child)
	return nil
}

// SetRight sets a horizontal split's right pane.
func (w *Widget) SetRight(child *Widget) error {
	s, ok := w.w.(*wtui.HSplit)
	if !ok {
		return kindErr(w, "h_split")
	}
	s.Right = childWidget(child)
	return nil
}

// SetLeftFraction sets a horizontal split's left-pane width percentage.
func (w *Widget) SetLeftFraction(n int) error {
	s, ok := w.w.(*wtui.HSplit)
	if !ok {
		return kindErr(w, "h_split")
	}
	s.LeftFrac = n
	return nil
}

// SetTop sets a vertical split's top pane.
func (w *Widget) SetTop(child *Widget) error {
	s, ok := w.w.(*wtui.VSplit)
	if !ok {
		return kindErr(w, "v_split")
	}
	s.Top = childWidget(child)
	return nil
}

// SetBottom sets a vertical split's bottom pane.
func (w *Widget) SetBottom(child *Widget) error {
	s, ok := w.w.(*wtui.VSplit)
	if !ok {
		return kindErr(w, "v_split")
	}
	s.Bottom = childWidget(child)
	return nil
}

// SetTopFraction sets a vertical split's top-pane height percentage.
func (w *Widget) SetTopFraction(n int) error {
	s, ok := w.w.(*wtui.VSplit)
	if !ok {
		return kindErr(w, "v_split")
	}
	s.TopFrac = n
	return nil
}

// AddTab appends a page to a notebook under the given tab label.
func (w *Widget) AddTab(label string, page *Widget) error {
	n, ok := w.w.(*wtui.Notebook)
	if !ok {
		return kindErr(w, "notebook")
	}
	n.AddTab(label, childWidget(page))
	return nil
}

// SetActive sets a notebook's active tab index.
func (w *Widget) SetActive(i int) error {
	n, ok := w.w.(*wtui.Notebook)
	if !ok {
		return kindErr(w, "notebook")
	}
	n.Active = i
	return nil
}

// Active returns a notebook's active tab index.
func (w *Widget) Active() (int, error) {
	n, ok := w.w.(*wtui.Notebook)
	if !ok {
		return 0, kindErr(w, "notebook")
	}
	return n.Active, nil
}

// --- callbacks (Widget methods) --------------------------------------------

// OnClick wires a button so that a click (or Enter while focused) records id in
// the owning Module's callback log; [Module.Dispatch] returns those ids.
func (w *Widget) OnClick(id string) error {
	b, ok := w.w.(*wtui.Button)
	if !ok {
		return kindErr(w, "button")
	}
	b.OnClick = func() { w.mod.fire(id) }
	return nil
}

// OnToggle wires a check button so a toggle records id.
func (w *Widget) OnToggle(id string) error {
	c, ok := w.w.(*wtui.CheckButton)
	if !ok {
		return kindErr(w, "check_button")
	}
	c.OnToggle = func(bool) { w.mod.fire(id) }
	return nil
}

// OnSelect wires a list box so a selection change records id.
func (w *Widget) OnSelect(id string) error {
	l, ok := w.w.(*wtui.ListBox)
	if !ok {
		return kindErr(w, "list_box")
	}
	l.OnSelect = func(int) { w.mod.fire(id) }
	return nil
}

// OnChange wires an entry so an edit records id.
func (w *Widget) OnChange(id string) error {
	e, ok := w.w.(*wtui.Entry)
	if !ok {
		return kindErr(w, "entry")
	}
	e.OnChange = func(string) { w.mod.fire(id) }
	return nil
}

// OnTabChanged wires a notebook so a tab switch records id.
func (w *Widget) OnTabChanged(id string) error {
	n, ok := w.w.(*wtui.Notebook)
	if !ok {
		return kindErr(w, "notebook")
	}
	n.OnTabChanged = func(int) { w.mod.fire(id) }
	return nil
}

// --- layout / render / events (Module methods) -----------------------------

// SetSize lays root out to fill a (cols, rows) terminal. Non-positive
// dimensions fall back to the classic 80x24 default.
func (m *Module) SetSize(root *Widget, cols, rows int) error {
	if root == nil {
		return errNilWidget
	}
	cols, rows = clamp(cols, rows)
	root.w.SetBounds(toolkit.Rect{X: 0, Y: 0, W: cols, H: rows})
	return nil
}

// Bounds returns a widget's current placement as a Hash with "x", "y", "w" and
// "h".
func (m *Module) Bounds(w *Widget) (map[string]any, error) {
	if w == nil {
		return nil, errNilWidget
	}
	r := w.w.Bounds()
	return map[string]any{"x": r.X, "y": r.Y, "w": r.W, "h": r.H}, nil
}

// Render lays root out to (cols, rows) and returns the frame as a
// self-contained ANSI string a Ruby host can print straight to a terminal.
func (m *Module) Render(root *Widget, cols, rows int) (string, error) {
	if root == nil {
		return "", errNilWidget
	}
	cols, rows = clamp(cols, rows)
	root.w.SetBounds(toolkit.Rect{X: 0, Y: 0, W: cols, H: rows})
	var buf bytes.Buffer
	// RenderToolkitSized writes to a bytes.Buffer, whose Write never fails, so
	// the returned error is always nil here.
	_ = wtui.RenderToolkitSized(&buf, cols, rows, []toolkit.Widget{root.w}, nil)
	return buf.String(), nil
}

// RenderCells lays root out, renders it, and returns the decoded cell grid as a
// Ruby Hash (see [Module.DecodeCells] for the shape). This is the
// DecodeANSI-backed form a test asserts against at cell precision.
func (m *Module) RenderCells(root *Widget, cols, rows int) (map[string]any, error) {
	ansi, err := m.Render(root, cols, rows)
	if err != nil {
		return nil, err
	}
	cols, rows = clamp(cols, rows)
	return gridToRuby(wtui.DecodeANSI([]byte(ansi), cols, rows)), nil
}

// DecodeCells decodes an ANSI frame into a cell grid, returning a Ruby Hash
// with "cols", "rows", "cells" and "text". "cells" is an Array of rows, each an
// Array of cell Hashes {"char" => String, "fg" => color|nil, "bg" =>
// color|nil}; a color is a Hash {"r","g","b"} or nil for the terminal default.
// "text" is an Array of per-row strings for coarse assertions. It is backed by
// go-widgets/tui's DecodeANSI.
func (m *Module) DecodeCells(ansi string, cols, rows int) map[string]any {
	cols, rows = clamp(cols, rows)
	return gridToRuby(wtui.DecodeANSI([]byte(ansi), cols, rows))
}

// Dispatch routes one input event to root's widget tree and returns a Hash with
// "fired" (an Array of the callback ids that ran) and "repaint" (always true —
// an event may change state, so the caller should re-render). The event Hash
// carries "kind" (e.g. "click", "key_down", "char", "mouse_drag", "mouse_up",
// "tick"), "x", "y", "code"/"key"/"rune", "ctrl" and "shift"; all are optional.
func (m *Module) Dispatch(root *Widget, ev map[string]any) (map[string]any, error) {
	if root == nil {
		return nil, errNilWidget
	}
	m.fired = nil
	root.w.OnEvent(toEvent(ev))
	fired := make([]any, len(m.fired))
	copy(fired, m.fired)
	return map[string]any{"fired": fired, "repaint": true}, nil
}

// --- dynamic dispatch (the uniform surface rbgo binds via method_missing) ---

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// errNilWidget reports an operation attempted on a nil widget handle.
var errNilWidget = errors.New("tui: nil widget")

// Call dispatches a Ruby-style snake_case method name to the matching exported
// method of recv (a *Module or *Widget), coercing each Ruby-supplied argument
// to the Go parameter type. Trailing arguments may be omitted; they default to
// nil. The result is the method's Ruby-shaped return value (or nil for a method
// that returns nothing); a trailing error return is unwrapped into Call's own
// error. This is the single entry point an rbgo binding drives.
func Call(recv any, method string, args ...any) (any, error) {
	rv := reflect.ValueOf(recv)
	if !rv.IsValid() {
		return nil, fmt.Errorf("tui: nil receiver")
	}
	m := rv.MethodByName(camelize(method))
	if !m.IsValid() {
		return nil, fmt.Errorf("tui: unknown method %q", method)
	}
	mt := m.Type()
	if len(args) > mt.NumIn() {
		return nil, fmt.Errorf("tui: %s takes at most %d argument(s), got %d",
			method, mt.NumIn(), len(args))
	}

	in := make([]reflect.Value, mt.NumIn())
	for i := 0; i < mt.NumIn(); i++ {
		var arg any
		if i < len(args) {
			arg = args[i]
		}
		v, err := coerce(arg, mt.In(i))
		if err != nil {
			return nil, fmt.Errorf("tui: %s argument %d: %w", method, i+1, err)
		}
		in[i] = v
	}

	outs := m.Call(in)
	if n := len(outs); n > 0 && mt.Out(n-1) == errorType {
		if e := outs[n-1]; !e.IsNil() {
			return nil, e.Interface().(error)
		}
		outs = outs[:n-1]
	}
	if len(outs) == 0 {
		return nil, nil
	}
	return outs[0].Interface(), nil
}

// Methods lists, sorted, the Ruby-style snake_case names Call accepts for recv.
func Methods(recv any) []string {
	t := reflect.TypeOf(recv)
	names := make([]string, 0, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		names = append(names, snakeize(t.Method(i).Name))
	}
	sort.Strings(names)
	return names
}

// coerce converts a Ruby-supplied argument into the Go type target expects.
func coerce(val any, target reflect.Type) (reflect.Value, error) {
	if val != nil {
		if vv := reflect.ValueOf(val); vv.Type().AssignableTo(target) {
			return vv, nil
		}
	}
	switch target.Kind() {
	case reflect.String:
		if val == nil {
			return reflect.ValueOf("").Convert(target), nil
		}
		return reflect.ValueOf(fmt.Sprint(val)).Convert(target), nil
	case reflect.Int:
		n, ok := toInt(val)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot use %T as an integer", val)
		}
		return reflect.ValueOf(n).Convert(target), nil
	case reflect.Float64:
		f, ok := toFloat(val)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot use %T as a float", val)
		}
		return reflect.ValueOf(f).Convert(target), nil
	case reflect.Bool:
		return reflect.ValueOf(truthy(val)), nil
	case reflect.Slice:
		if val == nil { // an omitted Array becomes a nil slice
			return reflect.Zero(target), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot use %T as %s", val, target)
	case reflect.Map:
		if val == nil { // an omitted Hash becomes a nil map
			return reflect.Zero(target), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot use %T as %s", val, target)
	default: // pointers (a handle) must be non-nil + assignable; nothing else
		return reflect.Value{}, fmt.Errorf("cannot use %T as %s", val, target)
	}
}

// --- Ruby <-> Go value marshalling helpers ---------------------------------

// stringList coerces a Ruby Array (or a single value) into a []string.
func stringList(v []any) []string {
	out := make([]string, len(v))
	for i, e := range v {
		out[i] = fmt.Sprint(e)
	}
	return out
}

// childWidget unwraps a handle to its underlying widget, mapping a nil handle to
// a nil widget (an unset pane).
func childWidget(c *Widget) toolkit.Widget {
	if c == nil {
		return nil
	}
	return c.w
}

// clamp substitutes the classic 80x24 default for any non-positive dimension.
func clamp(cols, rows int) (int, int) {
	if cols <= 0 {
		cols = wtui.DefaultCols
	}
	if rows <= 0 {
		rows = wtui.DefaultRows
	}
	return cols, rows
}

// toEvent builds a toolkit.Event from a Ruby event Hash. A nil Hash yields a
// zero-position key event (reads from a nil map are safe).
func toEvent(ev map[string]any) toolkit.Event {
	return toolkit.Event{
		Kind:  eventKind(fmt.Sprint(ev["kind"])),
		X:     toIntOr(ev["x"]),
		Y:     toIntOr(ev["y"]),
		Code:  eventCode(ev),
		Ctrl:  truthy(ev["ctrl"]),
		Shift: truthy(ev["shift"]),
	}
}

// eventCode picks the key/character string from the first of "code", "key" or
// "rune" that is present and non-nil.
func eventCode(ev map[string]any) string {
	for _, k := range []string{"code", "key", "rune"} {
		if v, ok := ev[k]; ok && v != nil {
			return fmt.Sprint(v)
		}
	}
	return ""
}

// eventKind maps a Ruby event-kind name to a toolkit.EventKind. An unknown name
// is treated as a key-down.
func eventKind(name string) toolkit.EventKind {
	switch name {
	case "click":
		return toolkit.EventClick
	case "key_down", "keydown":
		return toolkit.EventKeyDown
	case "key_up", "keyup":
		return toolkit.EventKeyUp
	case "char":
		return toolkit.EventChar
	case "mouse_drag", "drag":
		return toolkit.EventMouseDrag
	case "mouse_up", "mouseup":
		return toolkit.EventMouseUp
	case "tick":
		return wtui.EventTick
	default:
		return toolkit.EventKeyDown
	}
}

// gridToRuby converts a decoded TermGrid to a Ruby Hash.
func gridToRuby(g *wtui.TermGrid) map[string]any {
	cells := make([]any, g.Rows)
	text := make([]any, g.Rows)
	for y := 0; y < g.Rows; y++ {
		row := make([]any, g.Cols)
		for x := 0; x < g.Cols; x++ {
			c := g.At(x, y)
			row[x] = map[string]any{
				"char": string(c.Rune),
				"fg":   colorToRuby(c.Fg),
				"bg":   colorToRuby(c.Bg),
			}
		}
		cells[y] = row
		text[y] = g.RowText(y)
	}
	return map[string]any{"cols": g.Cols, "rows": g.Rows, "cells": cells, "text": text}
}

// colorToRuby renders a decoded cell color as a Ruby Hash, or nil for the
// terminal default (an unset color).
func colorToRuby(c wtui.Color) any {
	if !c.Set {
		return nil
	}
	return map[string]any{"r": int(c.R), "g": int(c.G), "b": int(c.B)}
}

// toInt coerces a Ruby number to an int.
func toInt(val any) (int, bool) {
	switch x := val.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
}

// toIntOr coerces a Ruby number to an int, defaulting to 0.
func toIntOr(val any) int {
	if n, ok := toInt(val); ok {
		return n
	}
	return 0
}

// toFloat coerces a Ruby number to a float64.
func toFloat(val any) (float64, bool) {
	switch x := val.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	default:
		return 0, false
	}
}

// truthy applies Ruby truthiness: only false and nil are falsey.
func truthy(val any) bool {
	switch x := val.(type) {
	case nil:
		return false
	case bool:
		return x
	default:
		return true
	}
}

// kindErr reports a method invoked on a handle of the wrong widget kind.
func kindErr(w *Widget, want string) error {
	return fmt.Errorf("tui: %s handle is not a %s", w.kind, want)
}

// camelize turns a snake_case method name into its exported Go method name.
func camelize(s string) string {
	var b strings.Builder
	for _, w := range strings.Split(s, "_") {
		if w == "" {
			continue
		}
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	return b.String()
}

// snakeize is the inverse: an exported Go method name to snake_case.
func snakeize(s string) string {
	var b []rune
	rs := []rune(s)
	for i, r := range rs {
		if unicode.IsUpper(r) {
			prevLower := i > 0 && unicode.IsLower(rs[i-1])
			nextLower := i+1 < len(rs) && unicode.IsLower(rs[i+1])
			if i > 0 && (prevLower || nextLower) {
				b = append(b, '_')
			}
			r = unicode.ToLower(r)
		}
		b = append(b, r)
	}
	return string(b)
}

// --- package-level convenience wrappers over the default Module ------------

// Label builds a label on the default Module.
func Label(text string) *Widget { return mod.Label(text) }

// Button builds a button on the default Module.
func Button(label string) *Widget { return mod.Button(label) }

// Entry builds an entry on the default Module.
func Entry(initial string) *Widget { return mod.Entry(initial) }

// CheckButton builds a check button on the default Module.
func CheckButton(label string, checked bool) *Widget { return mod.CheckButton(label, checked) }

// ListBox builds a list box on the default Module.
func ListBox(items []any) *Widget { return mod.ListBox(items) }

// ProgressBar builds a progress bar on the default Module.
func ProgressBar() *Widget { return mod.ProgressBar() }

// Container builds a border-layout container on the default Module.
func Container() *Widget { return mod.Container() }

// HSplit builds a horizontal split on the default Module.
func HSplit() *Widget { return mod.HSplit() }

// VSplit builds a vertical split on the default Module.
func VSplit() *Widget { return mod.VSplit() }

// Notebook builds a notebook on the default Module.
func Notebook() *Widget { return mod.Notebook() }

// SetSize lays root out on the default Module.
func SetSize(root *Widget, cols, rows int) error { return mod.SetSize(root, cols, rows) }

// Bounds returns a widget's placement via the default Module.
func Bounds(w *Widget) (map[string]any, error) { return mod.Bounds(w) }

// Render renders root to an ANSI string via the default Module.
func Render(root *Widget, cols, rows int) (string, error) { return mod.Render(root, cols, rows) }

// RenderCells renders root to a cell grid via the default Module.
func RenderCells(root *Widget, cols, rows int) (map[string]any, error) {
	return mod.RenderCells(root, cols, rows)
}

// DecodeCells decodes an ANSI frame to a cell grid via the default Module.
func DecodeCells(ansi string, cols, rows int) map[string]any {
	return mod.DecodeCells(ansi, cols, rows)
}

// Dispatch routes an event to root via the default Module.
func Dispatch(root *Widget, ev map[string]any) (map[string]any, error) {
	return mod.Dispatch(root, ev)
}
