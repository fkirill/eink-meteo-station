package renderable

import (
	"fmt"
	"image"
)

func CompressRasterTo4bpp(rect image.Rectangle, screenSize image.Point, raster []byte, flipHorizontal bool) ([]byte, error) {
	if rect.Dx()%2 != 0 {
		return nil, fmt.Errorf("Width must be even, rect = %i", rect)
	}
	targetWidthBytes := rect.Dx() / 2
	targetImageSize := targetWidthBytes * rect.Dy()
	res := make([]byte, targetImageSize, targetImageSize)
	newIndex := 0
	screenWidthBytes := screenSize.X
	oldIndex := screenWidthBytes*rect.Min.Y + rect.Min.X
	deltaBytes := screenWidthBytes - rect.Dx()
	for y := 0; y < rect.Dy(); y++ {
		for x := 0; x < rect.Dx()/2; x++ {
			newByte := raster[oldIndex]&0xf0 + raster[oldIndex+1]&0x0f
			res[newIndex] = newByte
			oldIndex += 2
			newIndex++
		}
		oldIndex += deltaBytes
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
