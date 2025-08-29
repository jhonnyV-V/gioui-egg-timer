package main

import (
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var progress float32
var progressIncrementer chan float32
var boiling bool

func main() {
	go func() {
		window := new(app.Window)
		window.Option(app.Title("Egg Timer"))
		window.Option(app.Size(unit.Dp(400), unit.Dp(600)))
		window.Option(app.MinSize(unit.Dp(400), unit.Dp(600)))

		if err := loop(window); err != nil {
			log.Fatal(err)
		}
		//gracefully exit
		os.Exit(0)

	}()

	progressIncrementer = make(chan float32)
	go func() {
		for {
			time.Sleep(time.Second / 25)
			progressIncrementer <- 0.004
		}
	}()

	app.Main()
}

type Context = layout.Context
type Dimensions = layout.Dimensions

func loop(window *app.Window) error {
	var ops op.Ops
	theme := material.NewTheme()

	var startButton widget.Clickable

	go func() {
		for progressUpdate := range progressIncrementer {
			if boiling && progress < 1 {
				progress += progressUpdate
				window.Invalidate()
			}
		}
	}()

	for {
		event := window.Event()

		switch eventType := event.(type) {
		case app.FrameEvent:
			context := app.NewContext(&ops, eventType)

			if startButton.Clicked(context) {
				boiling = !boiling
			}

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(context,

				layout.Rigid(func(context Context) Dimensions {
					bar := material.ProgressBar(theme, progress)
					return bar.Layout(context)
				}),

				layout.Rigid(func(context Context) Dimensions {
					margins := layout.Inset{
						Bottom: unit.Dp(25),
						Top:    unit.Dp(25),
						Left:   unit.Dp(35),
						Right:  unit.Dp(35),
					}
					return margins.Layout(context, func(context Context) Dimensions {
						var text string
						if boiling {
							text = "stop"
						} else {
							text = "start"
						}
						button := material.Button(theme, &startButton, text)
						return button.Layout(context)
					})
				}),
			)

			eventType.Frame(context.Ops)
		case app.DestroyEvent:
			return eventType.Err
		}

	}

}
