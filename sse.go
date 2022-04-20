package htmx

import g "github.com/maragudk/gomponents"

func ServerSentEvents(path string) g.Node {
	return g.Attr("sse-connect", path)
}
