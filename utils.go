/*
@author: sk
@date: 2024/6/9
*/
package main

import (
	"image"
	"image/png"
	"os"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Sign(val float64) int {
	if val > 0 {
		return 1
	} else if val < 0 {
		return -1
	} else {
		return 0
	}
}

func Len2(offX float64, offY float64) float64 {
	return offX*offX + offY*offY
}

func OpenImg(path string) image.Image {
	file, err := os.Open(path)
	HandleErr(err)
	res, err := png.Decode(file)
	HandleErr(err)
	return res
}
