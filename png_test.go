/*
@author: sk
@date: 2024/6/10
*/
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestRemoveBlock(t *testing.T) {
	fileName := "res/item/pillar.png"
	file, err := os.Open(fileName)
	HandleErr(err)
	temp, err := png.Decode(file)
	HandleErr(err)
	res := image.NewRGBA(temp.Bounds())
	bound := temp.Bounds()
	for y := 0; y < bound.Dy(); y++ {
		for x := 0; x < bound.Dx(); x++ {
			clr := temp.At(x, y)
			r, g, b, _ := clr.RGBA()
			if r == 0 && g == 0 && b == 0 {
				clr = color.RGBA{}
			}
			res.Set(x, y, clr)
		}
	}
	file.Close()

	file, err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
	err = png.Encode(file, res)
	HandleErr(err)
	file.Close()
}
