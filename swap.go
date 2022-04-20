package htmx

import g "github.com/maragudk/gomponents"

const (
	SwapInnerHTML   = "innerHTML"   //The default, replace the inner html of the target element
	SwapOuterHTML   = "outerHTML"   //Replace the entire target element with the response
	SwapBeforeBegin = "beforebegin" //Insert the response before the target element
	SwapAfterBegin  = "afterbegin"  //Insert the response before the first child of the target element
	SwapBeforeEnd   = "beforeend"   //Insert the response after the last child of the target element
	SwapAfterEnd    = "afterend"    //Insert the response after the target element
	SwapDelete      = "delete"      //Deletes the target element regardless of the response
	SwapNone        = "none"        //Does not append content from response (out of band items will still be processed)
)

func Swap(value string) g.Node {
	return g.Attr("hx-swap", value)
}
