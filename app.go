/*
@author: sk
@date: 2024/6/9
*/
package main

import (
	"image"
	"image/color"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

type App struct {
	PosX, PosY float64 // 玩家位置
	Dir        float64 // 玩家朝向
	Fov        float64 // 半视野大小
	Dep        float64 // 成像平面距离相机的距离要保证其正好等于视频宽度
	Imgs       []image.Image
	Depths     []float64 // 深度值
}

func NewApp() *App {
	imgs := make([]image.Image, 0)
	imgs = append(imgs, OpenImg("res/wall/redbrick.png"))
	imgs = append(imgs, OpenImg("res/wall/eagle.png"))
	imgs = append(imgs, OpenImg("res/wall/greystone.png"))
	imgs = append(imgs, OpenImg("res/wall/colorstone.png"))
	imgs = append(imgs, OpenImg("res/wall/mossy.png"))
	imgs = append(imgs, OpenImg("res/wall/bluestone.png"))
	//imgs = append(imgs, OpenImg("res/wall/purplestone.png"))
	imgs = append(imgs, OpenImg("res/wall/wood.png"))
	imgs = append(imgs, OpenImg("res/item/barrel.png"))
	imgs = append(imgs, OpenImg("res/item/pillar.png"))
	imgs = append(imgs, OpenImg("res/item/greenlight.png"))
	return &App{PosX: 12, PosY: 12, Dir: 0, Fov: math.Pi * 33 / 180, Imgs: imgs,
		Depths: make([]float64, ScreenW), Dep: ScreenW / 2 / math.Tan(math.Pi*33/180)}
}

func (a *App) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		offX, offY := math.Cos(a.Dir)*0.01, math.Sin(a.Dir)*0.01
		if Map[int(a.PosY+offY)][int(a.PosX+offX)] == 0 { // 不能穿墙
			a.PosX += offX
			a.PosY += offY
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		offX, offY := math.Cos(a.Dir)*0.01, math.Sin(a.Dir)*0.01
		if Map[int(a.PosY-offY)][int(a.PosX-offX)] == 0 { // 不能穿墙
			a.PosX -= offX
			a.PosY -= offY
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyA) {
		a.Dir += math.Pi / 180
		if a.Dir >= math.Pi*2 {
			a.Dir -= math.Pi * 2
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		a.Dir -= math.Pi / 180
		if a.Dir < 0 {
			a.Dir += math.Pi * 2
		}
	}
	return nil
}

func (a *App) Debug(screen *ebiten.Image) {
	// 20 绘制地图
	for y := 0; y < len(Map); y++ {
		for x := 0; x < len(Map[y]); x++ {
			if Map[y][x] == 0 {
				continue
			}
			clr := a.GetClr(Map[y][x])
			vector.DrawFilledRect(screen, float32(x*20), float32(y*20), 20, 20, clr, false)
		}
	}
	for i := 0; i < 24; i++ { // 绘制坐标轴
		ebitenutil.DebugPrintAt(screen, strconv.Itoa(i), 0, i*20)
		ebitenutil.DebugPrintAt(screen, strconv.Itoa(i), i*20, 0)
	}
	// 绘制所有射线
	startX, startY := math.Cos(a.Dir+a.Fov), math.Sin(a.Dir+a.Fov)
	endX, endY := math.Cos(a.Dir-a.Fov), math.Sin(a.Dir-a.Fov)
	for x := 0; x < ScreenW; x++ {
		rate := float64(x) / ScreenW
		dirX := (1-rate)*startX + rate*endX
		dirY := (1-rate)*startY + rate*endY
		cx, cy, type0 := a.Collision(dirX, dirY)
		if type0 == 0 {
			continue
		}
		clr := a.GetClr(type0)
		vector.StrokeLine(screen, float32(a.PosX*20), float32(a.PosY*20), float32(cx*20), float32(cy*20), 1, clr, false)
	}
	vector.StrokeLine(screen, float32(a.PosX*20), float32(a.PosY*20), float32((a.PosX+math.Cos(a.Dir)*3)*20), float32((a.PosY+math.Sin(a.Dir)*3)*20), 1, colornames.White, false)
}

func (a *App) Draw(screen *ebiten.Image) {
	//a.Debug(screen)
	//return
	// 绘制天空与地面
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH/2, colornames.Saddlebrown, false)
	vector.DrawFilledRect(screen, 0, ScreenH/2, ScreenW, ScreenH/2, colornames.Gray, false)
	// 主要绘制
	startX, startY := math.Cos(a.Dir+a.Fov), math.Sin(a.Dir+a.Fov)
	endX, endY := math.Cos(a.Dir-a.Fov), math.Sin(a.Dir-a.Fov)
	forwardX, forwardY := math.Cos(a.Dir), math.Sin(a.Dir)
	for x := 0; x < ScreenW; x++ {
		rate := float64(x) / ScreenW
		dirX := (1-rate)*startX + rate*endX
		dirY := (1-rate)*startY + rate*endY
		cx, cy, type0 := a.Collision(dirX, dirY)
		if type0 == 0 {
			a.Depths[x] = math.MaxFloat64
			continue
		}
		l := a.GetLen(cx-a.PosX, cy-a.PosY, forwardX, forwardY) // 获取碰撞方向沿玩家向前方向的投影长度
		a.Depths[x] = l                                         // 更新深度值
		realH := ScreenH / l
		realY := (ScreenH - realH) / 2
		h := int(math.Min(realH, ScreenH)) // 要绘制线的高度
		y := (ScreenH - h) / 2
		xOff := math.Max(cx-float64(int(cx)), cy-float64(int(cy))) // 选最大的用于纹理
		for i := 0; i < h; i++ {
			clr := a.TextureClr(type0, xOff, (float64(y+i)-realY)/realH)
			screen.Set(x, y+i, clr)
		}
	}
	// 绘制精灵
	for _, sprite := range Sprites {
		dirX, dirY := sprite.X-a.PosX, sprite.Y-a.PosY
		dir, ok := a.GetDir(dirX, dirY)
		if !ok {
			continue
		}
		l := a.GetLen(dirX, dirY, forwardX, forwardY)
		size := int(ScreenH / l) // 方形的
		minX := a.GetScreenPos(dirX, dirY, l, dir) - size/2
		minY := (ScreenH - size) / 2
		for i := 0; i < size; i++ {
			x := minX + i
			if x < 0 || x >= ScreenW || l > a.Depths[x] {
				continue
			}
			a.Depths[x] = l // 更新深度值
			for j := 0; j < size; j++ {
				y := minY + j
				if y < 0 || y >= ScreenH {
					continue
				}
				clr := a.TextureClr(sprite.Index, float64(i)/float64(size), float64(j)/float64(size))
				if _, _, _, a0 := clr.RGBA(); a0 > 0 {
					screen.Set(x, y, clr) // 精灵要处理透明色
				}
			}
		}
	}
	// 绘制小地图
	vector.DrawFilledRect(screen, 0, 0, MapW*6, MapH*6, colornames.Black, false)
	for y := 0; y < len(Map); y++ {
		for x := 0; x < len(Map[y]); x++ {
			if Map[y][x] == 0 {
				continue
			}
			clr := a.GetClr(Map[y][x])
			vector.DrawFilledRect(screen, float32(x*6), float32(y*6), 6, 6, clr, false)
		}
	} // 绘制玩家位置
	x := a.PosX * 6
	y := a.PosY * 6
	vector.StrokeLine(screen, float32(x), float32(y-3), float32(x), float32(y+3), 1, colornames.White, false)
	vector.StrokeLine(screen, float32(x-3), float32(y), float32(x+3), float32(y), 1, colornames.White, false)
	vector.StrokeLine(screen, float32(x), float32(y), float32(x+math.Cos(a.Dir)*5), float32(y+math.Sin(a.Dir)*5), 1, colornames.Pink, false)
}

func (a *App) GetDir(dirX float64, dirY float64) (int, bool) {
	dir := math.Atan2(dirY, dirX)
	if dir < 0 {
		dir += math.Pi * 2
	}
	if math.Abs(dir-a.Dir) > a.Fov {
		return 0, false
	}
	if dir < a.Dir {
		return 1, true
	}
	return -1, true
}

func (a *App) GetClr(type0 int) color.Color {
	switch type0 {
	case 1:
		return colornames.Blue
	case 2:
		return colornames.Green
	case 3:
		return colornames.Red
	case 4:
		return colornames.Aqua
	case 5:
		return colornames.Yellow
	}
	return colornames.White
}

func (a *App) GetLen(x float64, y float64, dirX float64, dirY float64) float64 {
	return x*dirX + y*dirY // 不用除 dir的长度了，他的长度就是 1
}

func (a *App) Collision(dirX float64, dirY float64) (float64, float64, int) {
	// 两个方向的结果都要取
	x1, y1, r1 := a.collisionDirX(dirX, dirY)
	x2, y2, r2 := a.collisionDirY(dirX, dirY)
	// 有任意一方放弃取对面的
	if r1 == 0 {
		return x2, y2, r2
	}
	if r2 == 0 {
		return x1, y1, r1
	}
	// 否则取距离短的
	if Len2(x1-a.PosX, y1-a.PosY) < Len2(x2-a.PosX, y2-a.PosY) {
		return x1, y1, r1
	}
	return x2, y2, r2
}

func (a *App) Layout(_, _ int) (int, int) {
	return ScreenW, ScreenH
}

func (a *App) collisionDirX(dirX float64, dirY float64) (float64, float64, int) {
	// 先保证 x 是整数进行偏移
	if dirX != 0 {
		offX := Sign(dirX)
		x := int(a.PosX) // 默认 x 就是其位置，若不一致，就进行纠正
		if float64(x) != a.PosX && offX > 0 {
			x++
		}
		for {
			scale := (float64(x) - a.PosX) / dirX
			y := int(a.PosY + scale*dirY)
			xi := x
			if offX < 0 { // 沿x 对齐，若是向负方向索引-1
				xi--
			}
			if y >= 0 && y < len(Map) && xi >= 0 && xi < len(Map[y]) {
				if Map[y][xi] != 0 {
					return float64(x), a.PosY + scale*dirY, Map[y][xi]
				} else {
					x += offX
				}
			} else { // 出界了
				break
			}
		}
	}
	return 0, 0, 0
}

func (a *App) collisionDirY(dirX float64, dirY float64) (float64, float64, int) {
	// 再判断 y 与 x类似
	if dirY != 0 {
		offY := Sign(dirY)
		y := int(a.PosY) // 默认 y 就是其位置，若不一致，就进行纠正
		if float64(y) != a.PosY && offY > 0 {
			y++
		}
		for {
			scale := (float64(y) - a.PosY) / dirY
			x := int(a.PosX + scale*dirX)
			yi := y
			if offY < 0 { // 沿y对齐，若是向负方向索引-1
				yi--
			}
			if yi >= 0 && yi < len(Map) && x >= 0 && x < len(Map[yi]) {
				if Map[yi][x] != 0 {
					return a.PosX + scale*dirX, float64(y), Map[yi][x]
				} else {
					y += offY
				}
			} else { // 出界了
				break
			}
		}
	}
	return 0, 0, 0
}

func (a *App) TextureClr(type0 int, x float64, y float64) color.Color {
	img := a.Imgs[type0-1]
	return img.At(int(x*ImgW), int(y*ImgH))
}

func (a *App) GetScreenPos(dirX float64, dirY float64, l float64, dir int) int {
	l2 := dirX*dirX + dirY*dirY
	w := math.Sqrt(l2 - l*l)
	w /= l / a.Dep
	return ScreenW/2 + dir*int(w)
}
