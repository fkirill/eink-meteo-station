package utils

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"time"
)

func BoundingBox(offset, size image.Point) image.Rectangle {
	return image.Rectangle{Min: offset, Max: image.Point{X: offset.X + size.X, Y: offset.Y + size.Y}}
}

func WriteImageFromRaster(size image.Point, raster []byte, filename string) error {
	i := image.NewNRGBA(image.Rectangle{Min: image.Point{}, Max: size})
	index := 0
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			grayColor := raster[index]
			c := color.NRGBA{grayColor, grayColor, grayColor, 255}
			i.Set(x, y, c)
			index++
		}
	}
	pngBuffer := bytes.Buffer{}
	err := png.Encode(&pngBuffer, i)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, pngBuffer.Bytes(), 0755)
	if err != nil {
		return err
	}
	return nil
}

func ConvertToGrayScale(img image.Image) ([]byte, error) {
	size := img.Bounds().Size()
	var raster []byte = nil
	nrgba, success := img.(*image.NRGBA)
	if success {
		raster = make([]byte, size.X*size.Y)
		source := nrgba.Pix
		nrgbaIndex := 0
		grayIndex := 0
		for y := 0; y < size.Y; y++ {
			for x := 0; x < size.X; x++ {
				r := int(source[nrgbaIndex])
				nrgbaIndex++
				g := int(source[nrgbaIndex])
				nrgbaIndex++
				b := int(source[nrgbaIndex])
				nrgbaIndex++
				// we don't use alpha
				// a := source[nrgbaIndex]
				nrgbaIndex++
				// https://www.tutorialspoint.com/dip/grayscale_to_rgb_conversion.htm
				gray := (299*r + 587*g + 114*b) / 1000
				// coarse it down to 16 colours
				grayShade := 0
				if gray >= 255-8 {
					grayShade = 15
				} else {
					grayShade = (gray + 8) >> 4
				}
				if grayShade >= 16 {
					panic("unexpected gray color")
				}
				// making it a byte: 0x00, 0x11, 0x22, 0x33 ... 0xff
				grayShade = grayShade<<4 + grayShade
				raster[grayIndex] = byte(grayShade)
				grayIndex++
			}
		}
	} else {
		grey, success := img.(*image.Gray)
		if !success {
			return nil, errors.New("expected NRGBA or Gray")
		}
		raster = grey.Pix
	}
	return raster, nil
}

func LoadImage(fileName string) (image.Image, error) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}
	return img, nil
}

type TimeProvider interface {
	Now() time.Time
}

type offsetTimeProvider struct {
	offset time.Duration
}

func (o *offsetTimeProvider) Now() time.Time {
	return time.Now().UTC().Add(o.offset)
}

func NewTestTimeProvider(startTime time.Time) TimeProvider {
	now := time.Now()
	_, tzOffset := now.Zone()
	testOffset := int(startTime.Sub(now).Seconds())
	return &offsetTimeProvider{time.Duration(tzOffset+testOffset) * time.Second}
}

func NewTimeProvider() TimeProvider {
	now := time.Now()
	_, tzOffset := now.Zone()
	return &offsetTimeProvider{time.Duration(tzOffset) * time.Second}
}

func DrawImage(targetImage []byte, targetImageSize image.Point, targetOffset image.Point, sourceImage []byte, sourceImageSize image.Point) {
	if targetOffset.X + sourceImageSize.X > targetImageSize.X || targetOffset.Y + sourceImageSize.Y > targetImageSize.Y {
		panic("images don't overlap fully")
	}
	if len(targetImage) != targetImageSize.X * targetImageSize.Y || len(sourceImage) != sourceImageSize.X * sourceImageSize.Y {
		fmt.Printf("sourceImageSize %v, expectedLen %d, actualLen %d\n", sourceImageSize, len(sourceImage), sourceImageSize.X * sourceImageSize.Y)
		fmt.Printf("targetImageSize %v, expectedLen %d, actualLen %d\n", targetImageSize, len(targetImage), targetImageSize.X * targetImageSize.Y)
		panic("wrong image sizes")
	}
	dstIndex := targetOffset.Y*targetImageSize.X + targetOffset.X
	srcIndex := 0
	for y := 0; y < sourceImageSize.Y; y++ {
		src := sourceImage[srcIndex : srcIndex+sourceImageSize.X-1]
		dst := targetImage[dstIndex : dstIndex+sourceImageSize.X-1]
		copy(dst, src)
		srcIndex += sourceImageSize.X
		dstIndex += targetImageSize.X
	}
}
