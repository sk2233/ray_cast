/*
@author: sk
@date: 2024/6/9
*/
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// https://lodev.org/cgtutor/raycasting3.html

func main() {
	//fmt.Println(math.Pi / math.Atan2(1, 1))
	ebiten.SetWindowSize(ScreenW, ScreenH)
	err := ebiten.RunGame(NewApp())
	HandleErr(err)
}
