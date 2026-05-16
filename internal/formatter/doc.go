// Package formatter provides text rendering for log events.
//
// Three output styles are supported:
//
//   - StylePlain   — human-readable, no colour codes, suitable for piping.
//   - StyleColored — ANSI-coloured output for interactive terminals; severity
//     level drives the colour (debug=white, info=cyan, warn=yellow,
//     error=red, critical=magenta).
//   - StyleJSON    — single-line JSON object per event, useful for downstream
//     log processors.
//
// Usage:
//
//	f := formatter.New()          // defaults: colored, UTC, show source
//	f.Style = formatter.StyleJSON
//	fmt.Println(f.Format(event))
package formatter
