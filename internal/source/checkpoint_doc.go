// Package source — checkpoint
//
// # Overview
//
// The Checkpoint type provides lightweight cursor persistence for log sources.
// When logstream-tail is restarted it can resume from the last successfully
// processed event rather than replaying the entire look-back window.
//
// # Usage
//
//	cp, err := source.NewCheckpoint("/var/lib/logstream-tail/state.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Restore cursor before the first poll.
//	lastToken := cp.Get("my-log-group")
//
//	// After a successful poll, advance the cursor.
//	if err := cp.Set("my-log-group", nextToken); err != nil {
//		log.Printf("checkpoint write failed: %v", err)
//	}
//
// # File format
//
// State is stored as indented JSON so it is human-readable and easy to
// inspect or reset manually:
//
//	{
//	  "cursors": {
//	    "my-log-group": "f/12345/0000"
//	  },
//	  "updated_at": "2024-01-15T10:30:00Z"
//	}
package source
