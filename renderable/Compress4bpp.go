package renderable

import (
	"fmt"
	"image"
	"slices"
)

func CompressRasterTo4bpp(rect image.Rectangle, screenSize image.Point, raster []byte, flipHorizontal bool, fullScreen bool) ([]byte, error) {
	if rect.Dx()%2 != 0 {
		return nil, fmt.Errorf("Width must be even, rect = %v", rect)
	}
	targetImageSize := ((rect.Dx() / 2) + ternary(fullScreen, 0, 1)) * rect.Dy()
	res := make([]byte, targetImageSize, targetImageSize)
	newIndex := 0
	oldIndex := screenSize.X*rect.Min.Y + rect.Min.X
	for y := 0; y < rect.Dy(); y++ {
		row := raster[oldIndex : oldIndex+rect.Dx()]
		if flipHorizontal {
			reverseRow := make([]byte, rect.Dx(), rect.Dx())
			copy(reverseRow, row)
			slices.Reverse(reverseRow)
			row = reverseRow
		}
		rowIndex := 0
		for x := 0; x < rect.Dx()/2; x++ {
			newByte := row[rowIndex]&0xf0 + row[rowIndex+1]&0x0f
			res[newIndex] = newByte
			rowIndex += 2
			newIndex++
		}
		if !fullScreen {
			newIndex++
		}
		oldIndex += rect.Dx()
	}
	return res, nil
}

func ternary[Data any](condition bool, ifTrue Data, ifFalse Data) Data {
	if condition {
		return ifTrue
	} else {
		return ifFalse
	}
}

func CutRectangle(rect image.Rectangle, screenSize image.Point, raster []byte) []byte {
	res := make([]byte, rect.Dx()*rect.Dy(), rect.Dx()*rect.Dy())
	srcIdx := screenSize.X*rect.Min.Y + rect.Min.X
	dstIdx := 0
	for range rect.Dy() {
		srcRow := raster[srcIdx : srcIdx+rect.Dx()-1]
		dstRow := res[dstIdx : dstIdx+rect.Dx()-1]
		copy(dstRow, srcRow)
		srcIdx += screenSize.X
		dstIdx += rect.Dx()
	}
	return res
}
