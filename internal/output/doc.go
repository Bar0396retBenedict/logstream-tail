// Package output provides the terminal-writing layer for logstream-tail.
//
// A [Writer] bridges the fan-in event channel produced by [source.FanIn] and
// a destination [io.Writer] (typically os.Stdout).  It delegates all
// formatting decisions to a [formatter.Formatter], keeping concerns cleanly
// separated.
//
// Typical usage:
//
//	f := formatter.New(formatter.StyleColored, false)
//	w := output.New(os.Stdout, f)
//	if err := w.Run(ctx, eventCh); err != nil && !errors.Is(err, context.Canceled) {
//		log.Fatal(err)
//	}
package output
