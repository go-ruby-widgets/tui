# Examples

Runnable Ruby programs that drive the `require "tui"` adapter through
[rbgo](https://github.com/go-embedded-ruby/ruby).

## `dashboard.rb`

A tiny terminal dashboard: a titled header band, a scrollable mailbox list in
the body, and a status-line footer bound to the list selection. It lays the UI
out, drives it with two `Down` key presses, renders it to a decoded cell grid,
and asserts (at cell precision) that the expected text appears.

```sh
rbgo examples/dashboard.rb
# => OK selected=2 status=3/5: Sent
```

The same scenario is exercised as a Go acceptance test in
[`dashboard_example_test.go`](../dashboard_example_test.go): `TestDashboardExample`
reproduces it through `tui.Call` (the dynamic-dispatch entry point the rbgo
binding drives) so it runs in CI with no interpreter dependency, and
`TestDashboardExampleThroughRbgo` runs `dashboard.rb` through the real `rbgo`
binary when it is on `PATH`.
