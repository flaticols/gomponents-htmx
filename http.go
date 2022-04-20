package htmx

import g "github.com/maragudk/gomponents"

func Get(path string) g.Node {
	return g.Attr("hx-get", path)
}

func Post(path string) g.Node {
	return g.Attr("hx-post", path)
}

func Patch(path string) g.Node {
	return g.Attr("hx-patch", path)
}

func Put(path string) g.Node {
	return g.Attr("hx-put", path)
}

func Delete(path string) g.Node {
	return g.Attr("hx-delete", path)
}
