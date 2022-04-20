package ext

import (
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
)

const (
	ExtClientSideTemplates = "client-side-templates" //ExtClientSideTemplates client side templates extension name
)

func ClientSideTemplates() g.Node {
	return html.Script(html.Src("https://unpkg.com/htmx.org/dist/ext/client-side-templates.js"),
		g.Attr("crossorigin", "anonymous"),
		g.Attr("defer"),
	)
}

func MustacheTemplate(elementId string) g.Node {
	return g.Attr("mustache-template", elementId)
}

func HandlebarsTemplate(elementId string) g.Node {
	return g.Attr("handlebars-template", elementId)
}

func NunjucksTemplate(elementId string) g.Node {
	return g.Attr("nunjucks-template", elementId)
}
