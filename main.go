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
	Expanded       bool
	ExternalInset  layout.Inset
	InternalInset  layout.Inset
	ClosedHeight   unit.Value
	ExpandedHeight unit.Value
	CornerRadius   unit.Value
	Color          color.NRGBA
	clickable      widget.Clickable
}

func NewCard(
	closedHeight unit.Value,
	expandedHeight unit.Value,
	color color.NRGBA,
	externalInset layout.Inset,
	internalInset layout.Inset,
	cornerRadius unit.Value) Card {
	return Card{
		ExternalInset:  externalInset,
		InternalInset:  internalInset,
		ClosedHeight:   closedHeight,
		ExpandedHeight: expandedHeight,
		CornerRadius:   cornerRadius,
		Color:          color,
		clickable:      widget.Clickable{},
	}
}

func (c *Card) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	if c.Clicked() {
		c.Expanded = !c.Expanded
	}

	return c.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return c.ExternalInset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			height := float32(gtx.Px(c.ClosedHeight))
			if c.Expanded {
				height = float32(gtx.Px(c.ExpandedHeight))
			}
			rr := float32(gtx.Px(c.CornerRadius))
			sz := f32.Point{
				X: float32(gtx.Constraints.Max.X),
				Y: height,
			}
			r := f32.Rectangle{Max: sz}
			rrect := clip.UniformRRect(r, rr)
			paint.FillShape(gtx.Ops, c.Color, rrect.Op(gtx.Ops))
			gtx.Constraints = layout.Exact(image.Point{X: int(rrect.Rect.Max.X), Y: int(rrect.Rect.Max.Y)})
			return c.InternalInset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				d := w(gtx)
				d.Size = gtx.Constraints.Max
				return d
			})
		})
	})
}

func (c *Card) Clicked() bool {
	return c.clickable.Clicked()
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
						cds[i] = NewCard(unit.Dp(100),
							unit.Dp(200),
							th.ContrastBg,
							layout.Inset{
								Top:  unit.Dp(5),
								Left: unit.Dp(5),
							},
							layout.UniformInset(unit.Dp(5)),
							unit.Dp(30))
					}
				}
				return func(gtx layout.Context) layout.Dimensions {
					children := make([]layout.FlexChild, 0, count)
					for i := 0; i < count; i++ {
						idx := i
						children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return cds[idx].Layout(gtx, material.H2(th, fmt.Sprintf("Child #%d", idx+1)).Layout)
						}))
					}

					return material.List(th, &list).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
						return layout.Flex{
							Axis:      layout.Vertical,
							Alignment: layout.Middle,
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
