package htmx

import g "maragu.dev/gomponents"

func ServerSentEvents(path string) g.Node {
	return g.Attr("sse-connect", path)
}
