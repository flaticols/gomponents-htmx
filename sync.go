package htmx

import g "maragu.dev/gomponents"

const (
	SyncDrop       = "drop"
	SyncAbort      = "abort"
	SyncReplace    = "replace"
	SyncQueue      = "queue"
	SyncQueueFirst = "queue first"
	SyncQueueLast  = "queue last"
	SyncQueueAll   = "queue all"
)

func Sync(val string) g.Node {
	return g.Attr("hx-sync", val)
}
