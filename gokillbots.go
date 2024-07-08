package main

import (
	"fmt"
	"image"
	_ "image/png"
	"math/rand"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const (
	winWidth        = 800
	winHeight       = 600
	spriteSize      = 32
	spriteHeight    = 64
	bulletSize      = 2
	bulletSpeed     = 750
	enemySpeed      = 50
	maxEnemySpeed   = 250
	minEnemies      = 2
	maxEnemies      = 10
	playerMoveSpeed = 200
	playerMaxHearts = 3
	heartSpawnRate  = 0.02
)

type Bullet struct {
	pos pixel.Vec
	vel pixel.Vec
}

type Enemy struct {
	sprite  *pixel.Sprite
	pos     pixel.Vec
	vel     pixel.Vec
	frames  []pixel.Rect
	elapsed float64
}

type Heart struct {
	sprite *pixel.Sprite
	pos    pixel.Vec
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open picture: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("could not decode picture: %w", err)
	}

	return pixel.PictureDataFromImage(img), nil
}

func spawnEnemy(spriteSheet pixel.Picture, frames []pixel.Rect, playerPos pixel.Vec, elapsedTime float64) *Enemy {
	var pos pixel.Vec
	side := rand.Intn(4)
	switch side {
	case 0:
		pos = pixel.V(rand.Float64()*winWidth, winHeight)
	case 1:
		pos = pixel.V(rand.Float64()*winWidth, 0)
	case 2:
		pos = pixel.V(winWidth, rand.Float64()*winHeight)
	case 3:
		pos = pixel.V(0, rand.Float64()*winHeight)
	}

	speedIncrease := (maxEnemySpeed - enemySpeed) * (elapsedTime / 60)
	if speedIncrease > maxEnemySpeed-enemySpeed {
		speedIncrease = maxEnemySpeed - enemySpeed
	}
	dir := playerPos.Sub(pos).Unit().Scaled(float64(enemySpeed) + speedIncrease)
	return &Enemy{
		sprite:  pixel.NewSprite(spriteSheet, frames[0]),
		pos:     pos,
		vel:     dir,
		frames:  frames,
		elapsed: 0,
	}
}

func spawnHeart(heartSheet pixel.Picture) *Heart {
	pos := pixel.V(rand.Float64()*winWidth, rand.Float64()*winHeight)
	return &Heart{
		sprite: pixel.NewSprite(heartSheet, pixel.R(0, 0, spriteSize, spriteSize)),
		pos:    pos,
	}
}
func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "GoKillBots",
		Bounds: pixel.R(0, 0, winWidth, winHeight),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	playerSheet, err := loadPicture("player.png")
	if err != nil {
		panic(err)
	}

	enemySheet, err := loadPicture("enemy.png")
	if err != nil {
		panic(err)
	}

	heartSheet, err := loadPicture("heart.png")
	if err != nil {
		panic(err)
	}

	// Create frames for player and enemy animation
	playerFrames := []pixel.Rect{
		pixel.R(0, 0, spriteSize, spriteHeight),
		pixel.R(spriteSize, 0, spriteSize*2, spriteHeight),
	}

	enemyFrames := []pixel.Rect{
		pixel.R(0, 0, spriteSize, spriteHeight),
		pixel.R(spriteSize, 0, spriteSize*2, spriteHeight),
	}

	// Import base font
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	// Game loop
	restart := true
	for restart {
		playerSprite := pixel.NewSprite(playerSheet, playerFrames[0])
		playerPos := pixel.V(winWidth/2, winHeight/2)
		playerFrameIndex := 0
		playerFrameTime := 0.1 // time per frame in seconds
		playerElapsed := 0.0
		playerHealth := playerMaxHearts

		bullets := []*Bullet{}
		enemies := []*Enemy{}
		hearts := []*Heart{}
		kills := 0

		startTime := time.Now()

		last := time.Now()
		paused := false

		for !win.Closed() {
			if win.JustPressed(pixelgl.KeyEscape) {
				paused = !paused
				if paused {
					last = time.Now() // Reset the last update time to prevent time jump after pause
				}
			}

			if paused {
				survivalTime := time.Since(startTime)
				quit, cont := pauseScreen(win, basicAtlas, kills, survivalTime, enemies)
				if quit {
					return
				}
				if !cont {
					break
				}
				paused = false
				last = time.Now() // Reset the last update time to prevent time jump after pause
				continue
			}

			// Only update the game state when not paused
			dt := time.Since(last).Seconds()
			last = time.Now()
			playerElapsed += dt

			elapsedTime := time.Since(startTime).Seconds()
			win.SetTitle(fmt.Sprintf("GoKillBots - Time: %d:%02d", int(elapsedTime)/60, int(elapsedTime)%60))

			if win.Closed() {
				return
			}

			// Handle player movement
			if win.Pressed(pixelgl.KeyA) {
				playerPos.X -= playerMoveSpeed * dt
				if playerPos.X < spriteSize/2 {
					playerPos.X = spriteSize / 2
				}
			}
			if win.Pressed(pixelgl.KeyD) {
				playerPos.X += playerMoveSpeed * dt
				if playerPos.X > winWidth-spriteSize/2 {
					playerPos.X = winWidth - spriteSize/2
				}
			}
			if win.Pressed(pixelgl.KeyW) {
				playerPos.Y += playerMoveSpeed * dt
				if playerPos.Y > winHeight-spriteHeight/2 {
					playerPos.Y = winHeight - spriteHeight/2
				}
			}
			if win.Pressed(pixelgl.KeyS) {
				playerPos.Y -= playerMoveSpeed * dt
				if playerPos.Y < spriteHeight/2 {
					playerPos.Y = spriteHeight / 2
				}
			}

			// Shoot bullets
			if win.JustPressed(pixelgl.KeySpace) {
				bulletDir := pixel.V(0, bulletSpeed)
				if len(enemies) > 0 {
					closestEnemy := enemies[0]
					for _, enemy := range enemies {
						if playerPos.To(enemy.pos).Len() < playerPos.To(closestEnemy.pos).Len() {
							closestEnemy = enemy
						}
					}
					bulletDir = closestEnemy.pos.Sub(playerPos).Unit().Scaled(bulletSpeed)
				}
				bullet := &Bullet{
					pos: playerPos,
					vel: bulletDir,
				}
				bullets = append(bullets, bullet)
			}

			// Update bullets
			for i := 0; i < len(bullets); i++ {
				bullets[i].pos = bullets[i].pos.Add(bullets[i].vel.Scaled(dt))
				if bullets[i].pos.X < 0 || bullets[i].pos.X > winWidth || bullets[i].pos.Y < 0 || bullets[i].pos.Y > winHeight {
					bullets = append(bullets[:i], bullets[i+1:]...)
					i--
				}
			}

			// Gradually increase the number of enemies and their speed
			if len(enemies) < maxEnemies {
				enemiesToSpawn := minEnemies + int(elapsedTime/10)
				if enemiesToSpawn > maxEnemies {
					enemiesToSpawn = maxEnemies
				}
				for len(enemies) < enemiesToSpawn {
					enemies = append(enemies, spawnEnemy(enemySheet, enemyFrames, playerPos, elapsedTime))
				}
			}

			// Chance to spawn hearts on the map if the players health is less than 3, only allow 1 heart on the map at the time.
			if playerHealth < 3 && rand.Float64() < heartSpawnRate && len(hearts) < 1 {
				hearts = append(hearts, spawnHeart(heartSheet))
			}

			// Update enemies
			for _, enemy := range enemies {
				speedIncrease := (maxEnemySpeed - enemySpeed) * (elapsedTime / 60)
				if speedIncrease > maxEnemySpeed-enemySpeed {
					speedIncrease = maxEnemySpeed - enemySpeed
				}
				enemy.vel = playerPos.Sub(enemy.pos).Unit().Scaled(float64(enemySpeed) + speedIncrease)
				enemy.pos = enemy.pos.Add(enemy.vel.Scaled(dt))
				enemy.elapsed += dt
				if enemy.elapsed >= playerFrameTime {
					enemy.elapsed -= playerFrameTime
					enemy.sprite.Set(enemySheet, enemy.frames[(int(enemy.elapsed/playerFrameTime)+1)%len(enemy.frames)])
				}
			}

			// Check for bullet-enemy collisions
			for i := 0; i < len(bullets); i++ {
				for j := 0; j < len(enemies); j++ {
					if bullets[i].pos.To(enemies[j].pos).Len() < spriteSize/2 {
						bullets = append(bullets[:i], bullets[i+1:]...)
						enemies = append(enemies[:j], enemies[j+1:]...)
						i--
						kills++
						break
					}
				}
			}

			// Check for player-enemy collisions
			for i := 0; i < len(enemies); i++ {
				if playerPos.To(enemies[i].pos).Len() < spriteSize/2 {
					enemies = append(enemies[:i], enemies[i+1:]...)
					i--
					//TODO:
					// - check if player was hit recently, and is in immune state, otherwise take health
					playerHealth--

					if playerHealth == 0 {
						survivalTime := time.Since(startTime)
						quit, retry := gameOverScreen(win, basicAtlas, kills, survivalTime)
						if quit {
							return
						}
						if retry {
							restart = true
						}
						goto endGameLoop // Exit current game loop
					}
				}
			}

			// Check for player-heart collisions
			for i := 0; i < len(hearts); i++ {
				if playerPos.To(hearts[i].pos).Len() < spriteSize/2 {
					if playerHealth < playerMaxHearts {
						playerHealth++
					}
					hearts = append(hearts[:i], hearts[i+1:]...)
					i--
				}
			}

			// Update player animation frame
			if playerElapsed >= playerFrameTime {
				playerElapsed -= playerFrameTime
				playerFrameIndex = (playerFrameIndex + 1) % len(playerFrames)
				playerSprite.Set(playerSheet, playerFrames[playerFrameIndex])
			}

			win.Clear(colornames.Black)

			// Draw bullets
			for _, bullet := range bullets {
				imd := imdraw.New(nil)
				imd.Color = pixel.RGB(1, 1, 1)
				imd.Push(bullet.pos)
				imd.Circle(bulletSize, 0)
				imd.Draw(win)
			}

			// Draw enemies
			for _, enemy := range enemies {
				enemy.sprite.Draw(win, pixel.IM.Moved(enemy.pos))
			}

			// Draw hearts
			for _, heart := range hearts {
				heart.sprite.Draw(win, pixel.IM.Moved(heart.pos))
			}

			// Draw player
			playerSprite.Draw(win, pixel.IM.Moved(playerPos))

			// Draw player hearts
			for i := 0; i < playerHealth; i++ {
				heart := pixel.NewSprite(heartSheet, pixel.R(0, 0, spriteSize, spriteSize))
				heart.Draw(win, pixel.IM.Moved(pixel.V(32+float64(i*40), winHeight-32)))
			}

			// Draw kills count
			txt := text.New(pixel.V(winWidth-100, winHeight-30), basicAtlas)
			txt.Color = colornames.White
			fmt.Fprintf(txt, "Kills: %d", kills)
			txt.Draw(win, pixel.IM)

			win.Update()
		}
	endGameLoop:
	}
}

func main() {
	pixelgl.Run(run)
}
