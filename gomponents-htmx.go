package htmx

import (
	g "maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

// IncludeHtmx added the script tag with htmx.
// Note: Using a CDN in production may not be ideal due to potential issues with
// reliability, security, and performance. It might be preferable to self-host or bundle
// the script for production deployments.
func IncludeHtmx() g.Node {
	return html.Script(
		html.Src("https://unpkg.com/htmx.org@2.0.4"),
		g.Attr("integrity", "sha384-HGfztofotfshcF7+8n44JQL2oJmowVChPTg48S+jvZoztPfvwD79OC/LTtG6dMp+"),
		g.Attr("crossorigin", "anonymous"),
		g.Attr("defer"),
	)
}
