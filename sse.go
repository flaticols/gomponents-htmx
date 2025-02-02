package htmx

import g "maragu.dev/gomponents"

// ServerSentEvents creates a group node with attributes to support Server-Sent Events for the specified connection path.
// HTMX Docs: https://htmx.org/extensions/sse/
func ServerSentEvents(path string) g.Node {
	return g.Group{
		g.Attr("hx-ext", "sse"),
		g.Attr("sse-connect", path),
	}
}

// ServerSentSwap sets up a server-sent event attribute with the provided message to enable SSE-based content swapping in the DOM.
// HTMX Docs: https://htmx.org/extensions/sse/
func ServerSentSwap(message string) g.Node {
	return g.Attr("sse-swap", message)
}

// ServerSentConnect sets up a Server-Sent Events connection by specifying the URL path as an attribute. Returns a g.Node with the configured attribute.
// HTMX Docs: https://htmx.org/extensions/sse/
func ServerSentConnect(path string) g.Node {
	return g.Attr("sse-connect", path)
}
