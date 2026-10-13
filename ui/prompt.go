package ui

import (
	"webtyp.com/components/actionbutton"
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/widget"
)

const nameUpdatePrompt = widget.Name("updateprompt")

var clsRoot = nameUpdatePrompt.Root()

// UpdatePrompt is a "New version · Reload" button that stays hidden until a new version of the app
// has been downloaded and waits (update.OnReady); clicking it lets that version take over
// (update.Apply) and the page reloads. Outside the browser it renders hidden and does nothing.
func UpdatePrompt() Component {
	p := &prompt{
		visible: NewBool(false),
	}
	p.Element = *Span().Set(clsRoot.AsAttr()).Child(Show(p.visible, func() *Element {
		btn := &actionbutton.ActionButton{
			Text:    "New version · Reload",
			Variant: "primary",
			OnClick: func(Event) {
				p.apply()
			},
		}
		return btn.Render()
	}))
	arm(p)
	return p
}

type prompt struct {
	Element
	visible *SignalBool
}

func (p *prompt) WidgetName() widget.Name { return nameUpdatePrompt }
func (p *prompt) WidgetKind() widget.Kind { return widget.Region }
