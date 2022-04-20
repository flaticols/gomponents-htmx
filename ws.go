package htmx

import (
	g "github.com/maragudk/gomponents"
)

func WebSockets(path string) g.Node {
	return g.Attr("ws-connect", path)
}

//ConnectWebSockets wrap child to "ws-connect" component
func ConnectWebSockets(path string, children ...g.Node) g.Node {
	children = append(children, Ext("ws"))
	children = append(children, WebSockets(path))

	return g.El("ws-connect", g.Group(children))
}
