package renderable

import (
	"errors"
	"image"
)

func CompressRasterTo4bpp(rect image.Rectangle, screenSize image.Point, raster []byte, flipHorizontal bool) ([]byte, error) {
	if rect.Dx()%2 != 0 {
		return nil, errors.New("Width must be even")
	}
	res := make([]byte, rect.Dx()*rect.Dy()/2, rect.Dx()*rect.Dy()/2)
	w := rect.Dx() / 2
	newIndex := 0
	for y := 0; y < rect.Dy(); y++ {
		oldIndex := screenSize.X*(rect.Min.Y+y)/2 + rect.Min.X/2
		for x := 0; x < w; x++ {
			newByte := raster[oldIndex]&0xf0 + raster[oldIndex+1]&0x0f
			res[newIndex] = newByte
			oldIndex += 2
			newIndex++
		}
	}
	if flipHorizontal {
		for y := 0; y < rect.Dy(); y++ {
			reverse4bpp(res[y*rect.Dx()/2 : (y+1)*rect.Dx()/2-1])
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
