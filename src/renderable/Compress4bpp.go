package renderable

import (
	"errors"
	"image"
)

func CompressRasterTo4bpp(size image.Point, raster []byte, flipHorizontal bool) ([]byte, error) {
	if size.X%2 != 0 {
		return nil, errors.New("Width must be even")
	}
	res := make([]byte, len(raster)/2, len(raster)/2)
	w := size.X / 2
	oldIndex := 0
	newIndex := 0
	for y := 0; y < size.Y; y++ {
		for x := 0; x < w; x++ {
			newByte := raster[oldIndex]&0xf0 + raster[oldIndex+1]&0x0f
			res[newIndex] = newByte
			oldIndex += 2
			newIndex++
		}
	}
	if flipHorizontal {
		for y := 0; y < size.Y; y++ {
			reverse4bpp(res[y*size.X/2 : (y+1)*size.X/2-1])
		}
	}
	return res, nil
}

func reverse4bpp(row []byte) {
	length := len(row)
	for i := 0; i < length/2; i++ {
		pixel1 := row[i]
		pixel2 := row[length-1-i]
		row[i] = ((pixel2 & 0x0f) << 4) + ((pixel2 & 0xf0) >> 4)
		row[length-1-i] = ((pixel1 & 0x0f) << 4) + ((pixel1 & 0xf0) >> 4)
	}
}
