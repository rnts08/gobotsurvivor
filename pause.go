package main

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

func pauseScreen(win *pixelgl.Window, basicAtlas *text.Atlas, kills int, survivalTime time.Duration, enemies []*Enemy) (bool, bool) {
	// Pause enemy movement
	pausedVelocities := make([]pixel.Vec, len(enemies))
	for i, enemy := range enemies {
		pausedVelocities[i] = enemy.vel
		enemy.vel = pixel.ZV
	}

	defer func() {
		// Restore enemy movement
		for i, enemy := range enemies {
			enemy.vel = pausedVelocities[i]
		}
	}()

	for !win.Closed() {
		win.Clear(colornames.Black)

		txt := text.New(pixel.V(winWidth/2-100, winHeight/2), basicAtlas)
		txt.Color = colornames.White
		fmt.Fprintln(txt, "Game Paused")
		txt.Draw(win, pixel.IM.Scaled(txt.Orig, 3))

		txt = text.New(pixel.V(winWidth/2-50, winHeight/2-100), basicAtlas)
		fmt.Fprintf(txt, "Kills: %d \n", kills)
		txt.Draw(win, pixel.IM.Scaled(txt.Orig, 2))

		txt = text.New(pixel.V(winWidth/2-50, winHeight/2-150), basicAtlas)
		fmt.Fprintf(txt, "Survived for: %d minutes %d seconds\n", int(survivalTime.Minutes()), int(survivalTime.Seconds())%60)
		txt.Draw(win, pixel.IM)

		txt = text.New(pixel.V(winWidth/2-50, winHeight/2-200), basicAtlas)
		fmt.Fprintln(txt, "[C] Continue")
		txt.Draw(win, pixel.IM)

		txt = text.New(pixel.V(winWidth/2-50, winHeight/2-230), basicAtlas)
		fmt.Fprintln(txt, "[Q] Quit")
		txt.Draw(win, pixel.IM)

		win.Update()

		if win.JustPressed(pixelgl.KeyC) {
			return false, true
		}
		if win.JustPressed(pixelgl.KeyQ) {
			return true, false
		}
		if win.Closed() {
			return true, false
		}
	}
	return false, false
}
