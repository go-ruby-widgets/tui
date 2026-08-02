# Copyright (c) 2026 the go-ruby-widgets/tui authors. All rights reserved.
# Use of this source code is governed by a BSD-3-Clause license that can be
# found in the LICENSE file at the root of this repository.
#
# dashboard.rb -- a runnable end-to-end demo of the `require "tui"` adapter.
#
# Run it through rbgo (github.com/go-embedded-ruby/ruby):
#
#     rbgo examples/dashboard.rb
#
# It builds a small terminal UI -- a titled header band, a scrollable mailbox
# list in the body, and a status-line footer bound to the list selection --
# lays it out, drives it with two "Down" key presses, renders it to a decoded
# cell grid, and asserts (at cell precision) that the expected text appears.
# On success it prints a single "OK ..." line; any mismatch raises.

require "tui"

# A tiny terminal dashboard: a titled header band, a scrollable mailbox list in
# the body, and a status-line footer that tracks the current selection.
items  = ["Inbox", "Drafts", "Sent", "Archive", "Trash"]
list   = Tui.list_box(items)
list.on_select("selection_changed")

header = Tui.label("  Mailboxes")
header.set_align(:left)

status = Tui.label("1/#{items.length}: #{items[0]}")

root = Tui.container
root.set_header_height(1)
root.set_footer_height(1)
root.set_header(header)
root.set_body(list)
root.set_footer(status)

Tui.set_size(root, 24, 8)

# Keep the footer bound to the list selection.
refresh = lambda do
  i = list.selected
  status.set_text("#{i + 1}/#{items.length}: #{items[i]}")
end

# Drive the UI with two "Down" key presses, refreshing the status each time.
2.times do
  Tui.dispatch(list, {"kind" => "key_down", "code" => "Down"})
  refresh.call
end

# Assert the rendered cell grid at cell precision.
grid = Tui.render_cells(root, 24, 8)
text = grid["text"].join("\n")

raise "header missing"    unless text.include?("Mailboxes")
raise "body item missing" unless text.include?("Sent")
raise "footer not bound"  unless text.include?("3/5: Sent")

# A single cell carries a char + optional colors.
raise "cell shape" unless grid["cells"][0][0].key?("char")

puts "OK selected=#{list.selected} status=#{status.text}"
