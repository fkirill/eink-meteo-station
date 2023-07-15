package renderable

import (
	"errors"
	"image"
	"renderable/utils"
)

type DiffRenderer interface {
	SingleRenderPass(raster []byte) (image.Rectangle, error)
}

func NewDiffRenderer(size image.Point) DiffRenderer {
	return &renderer{size: size}
}

type renderer struct {
	size           image.Point
	lastKnownImage []byte
}

func (r *renderer) SingleRenderPass(raster []byte) (image.Rectangle, error) {
	// first pass, just accept the whole image as the new one
	buf := make([]byte, len(raster), len(raster))
	copy(buf, raster)
	if r.lastKnownImage == nil {
		r.lastKnownImage = buf
		return utils.BoundingBox(image.Point{}, r.size), nil
	}
	rectangle, err := r.calculateDiffPoints(r.lastKnownImage, raster)
	r.lastKnownImage = buf
	if err != nil {
		return image.Rectangle{}, err
	}
	if rectangle == nil {
		return image.Rectangle{}, nil
	}
	return *rectangle, nil
}

func (r *renderer) calculateDiffPoints(image1, image2 []byte) (*image.Rectangle, error) {
	if len(image1) != len(image2) {
		return nil, errors.New("image sizes don't match")
	}

	minX := r.size.X + 1
	minY := r.size.Y + 1
	maxX := -1
	maxY := -1
	index := 0
	different := false
	for y := 0; y < r.size.Y; y++ {
		for x := 0; x < r.size.X; x++ {
			if image1[index] != image2[index] {
				different = true
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
			index++
		}
	}
	if !different {
		return nil, nil
	}
	return &image.Rectangle{
		Min: image.Point{X: minX, Y: minY},
		Max: image.Point{X: maxX, Y: maxY},
	}, nil
}
