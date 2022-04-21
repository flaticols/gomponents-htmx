package ext

import (
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
)

const (
	ExtJsonEnc = "json-enc" //ExtClientSideTemplates client side templates extension name
)

func JsonEnc() g.Node {
	return html.Script(html.Src("https://unpkg.com/htmx.org/dist/ext/json-enc.js"),
		g.Attr("crossorigin", "anonymous"),
		g.Attr("defer"),
	)
}

func ClassTools() g.Node {
	return html.Script(html.Src("https://unpkg.com/htmx.org/dist/ext/class-tools.js"),
		g.Attr("crossorigin", "anonymous"),
		g.Attr("defer"),
	)
}

func ServerSentEvents() g.Node {
	return html.Script(html.Src("https://unpkg.com/htmx.org/dist/ext/sse.js"),
		g.Attr("crossorigin", "anonymous"),
		g.Attr("defer"),
	)
}

func AlpineMorph() g.Node {
	return html.Script(html.Src("https://unpkg.com/htmx.org/dist/ext/alpine-morph.js"),
		g.Attr("crossorigin", "anonymous"),
		g.Attr("defer"),
	)
}
