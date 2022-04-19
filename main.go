package main

// A simple Gio program. See https://gioui.org for more information.

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"gioui.org/x/component"
)

func main() {
	go func() {
		w := app.NewWindow()
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

var contentInset = layout.Inset{
	Top:    unit.Dp(10),
	Bottom: unit.Dp(10),
	Left:   unit.Dp(10),
	Right:  unit.Dp(10),
}

type Card struct {
	content   string
	expanded  bool
	theme     *material.Theme
	clickable widget.Clickable
}

func NewCard(content string, theme *material.Theme) Card {
	return Card{
		content:   content,
		theme:     theme,
		clickable: widget.Clickable{},
	}
}

func (c *Card) layout(gtx layout.Context) layout.Dimensions {
	l := material.H2(c.theme, c.content)
	maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
	l.Color = maroon
	l.Alignment = text.Middle
	macro := op.Record(gtx.Ops)
	gtx.Constraints.Max.Y = gtx.Px(unit.Dp(100))
	contentDims := contentInset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return l.Layout(gtx)
	})
	contentOp := macro.Stop()

	height := contentDims.Size.Y
	if c.expanded {
		height *= 2
	}

	rrect := clip.UniformRRect(f32.Rect(
		float32(0),
		float32(0),
		float32(gtx.Constraints.Max.X),
		float32(height),
	), unit.Dp(30).V)
	paint.FillShape(gtx.Ops, c.theme.ContrastBg, rrect.Op(gtx.Ops))
	contentOp.Add(gtx.Ops)
	return layout.Dimensions{Size: image.Point{
		X: gtx.Constraints.Max.X,
		Y: height,
	}}

}

func (c *Card) Clicked() bool {
	return c.clickable.Clicked()
}

func (c *Card) Layout(gtx layout.Context) layout.Dimensions {
	if c.Clicked() {
		fmt.Printf("Clicked: %s\n", c.content)
		c.expanded = !c.expanded
	}
	return c.clickable.Layout(gtx, c.layout)
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops

	menuBtn, _ := widget.NewIcon(icons.NavigationMenu)

	modal := component.NewModal()
	nav := component.NewNav("Navigation Drawer", "This is an example.")
	modalNav := component.ModalNavFrom(&nav, modal)
	appBar := component.NewAppBar(modal)
	appBar.Title = "Hello there"
	appBar.NavigationIcon = menuBtn

	var list widget.List
	list.Axis = layout.Vertical
	list.Alignment = layout.Middle
	var cds []Card

	a := component.NavItem{
		Tag:  nil,
		Name: "Hello",
		Icon: nil,
	}
	modalNav.AddNavItem(a)

	na := component.VisibilityAnimation{
		State:    component.Invisible,
		Duration: time.Millisecond * 250,
	}

	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			for _, event := range appBar.Events(gtx) {
				switch event := event.(type) {
				case component.AppBarNavigationClicked:
					modalNav.Appear(gtx.Now)
					na.Disappear(gtx.Now)
				case component.AppBarContextMenuDismissed:
					log.Printf("Context menu dismissed: %v", event)
				case component.AppBarOverflowActionClicked:
					log.Printf("Overflow action selected: %v", event)
				}
			}

			cards := func(count int) func(layout.Context) layout.Dimensions {
				if cds == nil {
					cds = make([]Card, count)
					for i := 0; i < count; i++ {
						cds[i] = NewCard(fmt.Sprintf("Child #%d", i+1), th)
					}
				}
				return func(gtx layout.Context) layout.Dimensions {
					children := make([]layout.FlexChild, 0, count)
					for i := 0; i < count; i++ {
						idx := i
						children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return cds[idx].Layout(gtx)
						}))
					}

					return material.List(th, &list).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
						return layout.Flex{
							Axis:      layout.Vertical,
							Alignment: layout.Middle,
							Spacing:   layout.SpaceEvenly,
						}.Layout(gtx, children...)
					})
				}
			}

			content := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.X /= 3
						return modalNav.NavDrawer.Layout(gtx, th, &na)
					}),
					layout.Rigid(cards(100)),
				)
			})
			bar := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return appBar.Layout(gtx, th, "Menu", "Actions")
			})
			flex := layout.Flex{
				Axis: layout.Vertical,
			}
			flex.Layout(gtx, bar, content)
			modal.Layout(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}
