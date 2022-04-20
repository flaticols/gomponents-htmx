package htmx

import (
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
)

//IncludeHtmx added the script tag with htmx
func IncludeHtmx() g.Node {
	return html.Script(html.Src("https://unpkg.com/htmx.org@1.7.0"),
		g.Attr("integrity", "sha384-EzBXYPt0/T6gxNp0nuPtLkmRpmDBbjg6WmCUZRLXBBwYYmwAUxzlSGej0ARHX0Bo"),
		g.Attr("crossorigin", "anonymous"),
		g.Attr("defer"),
	)
}

//IncludeHyperScript added the script tag with hyperscript
func IncludeHyperScript() g.Node {
	return html.Script(html.Src("https://unpkg.com/hyperscript.org@0.9.5"),
		g.Attr("crossorigin", "anonymous"),
		g.Attr("defer"),
	)
}
