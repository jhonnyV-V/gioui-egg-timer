package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
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
	var boilDurationInput widget.Editor
	var boilDuration float32

	go func() {
		for range progressIncrementer {
			if boiling && progress < 1 {
				progress += 1.0 / 25.0 / boilDuration
				if progress >= 1 {
					progress = 1
				}
				// Force a redraw by invalidating the frame
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
				if progress >= 1 {
					progress = 0
				}

				inputFloat, err := strconv.ParseFloat(
					strings.TrimSpace(boilDurationInput.Text()),
					32,
				)

				if err != nil {
					fmt.Printf("error parsing boiling value: %#v\n", err)
				} else {
					boilDuration = float32(inputFloat)
					boilDuration = boilDuration / (1 - progress)
				}

			}

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(context,

				layout.Rigid(func(context Context) Dimensions {
					var eggPath clip.Path
					op.Offset(image.Pt(context.Dp(200), context.Dp(125))).Add(context.Ops)
					eggPath.Begin(context.Ops)
					// Rotate from 0 to 360 degrees
					for deg := 0.0; deg <= 360; deg++ {
						// Egg math (really) at this brilliant site. Thanks!
						// https://observablehq.com/@toja/egg-curve
						// Convert degrees to radians
						rad := deg * math.Pi / 180
						cosT := math.Cos(rad)
						sinT := math.Sin(rad)
						// Constants to define the eggshape
						a := 110.0
						b := 150.0
						d := 20.0

						x := a * cosT
						y := -(math.Sqrt(b*b-d*d*cosT*cosT) + d*sinT) * sinT

						point := f32.Pt(float32(x), float32(y))
						eggPath.LineTo(point)
					}
					eggPath.Close()

					eggArea := clip.Outline{Path: eggPath.End()}.Op()

					color := color.NRGBA{R: 255, G: uint8(239 * (1 - progress)), B: uint8(174 * (1 - progress)), A: 255}
					paint.FillShape(context.Ops, color, eggArea)

					d := image.Point{Y: 375}
					return layout.Dimensions{Size: d}
				}),

				layout.Rigid(func(context Context) Dimensions {
					editor := material.Editor(theme, &boilDurationInput, "sec")
					boilDurationInput.SingleLine = true
					boilDurationInput.Alignment = text.Middle
					if boiling && progress < 1 {
						remaining := (1 - progress) * boilDuration
						inputString := fmt.Sprintf("%.1f", math.Round(float64(remaining)*10)/10)
						boilDurationInput.SetText(inputString)
					}

					margins := layout.Inset{
						Top:    unit.Dp(0),
						Right:  unit.Dp(170),
						Bottom: unit.Dp(40),
						Left:   unit.Dp(170),
					}
					// ... and borders ...
					border := widget.Border{
						Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
						CornerRadius: unit.Dp(3),
						Width:        unit.Dp(2),
					}
					// ... before laying it out, one inside the other
					return margins.Layout(context,
						func(context Context) Dimensions {
							return border.Layout(context, editor.Layout)
						},
					)
				}),

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
						if boiling && progress < 1 {
							text = "stop"
						} else if boiling && progress >= 1 {
							text = "finished"
						} else if !boiling {
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
