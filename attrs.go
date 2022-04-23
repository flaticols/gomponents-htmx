package htmx

import (
	"encoding/json"
	"strconv"

	g "github.com/maragudk/gomponents"
)

func Boost(val bool) g.Node {
	return g.Attr("hx-boost", strconv.FormatBool(val))
}

func Confirm(msg string) g.Node {
	return g.Attr("hx-confirm", msg)
}

func Disable() g.Node {
	return g.Attr("hx-disable")
}

func Disinherit(val string) g.Node {
	return g.Attr("hx-disinherit", val)
}

func Encoding() g.Node {
	return g.Attr("hx-encoding", "multipart/form-data")
}

func Ext(ext string) g.Node {
	return g.Attr("hx-ext", ext)
}

func Headers(headers map[string]string) g.Node {
	buf, err := json.Marshal(headers)
	if err != nil {
		return g.Attr("hx-headers", err.Error())
	}

	return g.Attr("hx-headers", string(buf))
}

func HistoryElt() g.Node {
	return g.Attr("hx-history-elt")
}

func Include(val string) g.Node {
	return g.Attr("hx-include", val)
}

func Indicator(val string) g.Node {
	return g.Attr("hx-indicator", val)
}

func Params(val string) g.Node {
	return g.Attr("hx-params", val)
}

func Preserve(val string) g.Node {
	return g.Attr("hx-preserve", val)
}

func Prompt(msg string) g.Node {
	return g.Attr("hx-prompt", msg)
}

func PushUrl(val bool) g.Node {
	return g.Attr("hx-push-url", strconv.FormatBool(val))
}

func PushUrlValue(val string) g.Node {
	return g.Attr("hx-push-url", val)
}

func Request(val string) g.Node {
	return g.Attr("hx-request", val)
}

func Select(val string) g.Node {
	return g.Attr("hx-select", val)
}

func Target(val string) g.Node {
	return g.Attr("hx-target", val)
}

func Trigger(val string) g.Node {
	return g.Attr("hx-trigger", val)
}

func Vals(vals map[string]string) g.Node {
	buf, err := json.Marshal(vals)
	if err != nil {
		return g.Attr("hx-vals", err.Error())
	}

	return g.Attr("hx-vals", string(buf))
}
