package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"reflect"
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
			c := color.NRGBA{R: grayColor, G: grayColor, B: grayColor, A: 255}
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
	rgbaOrNrgba := false
	var source []uint8
	if nrgba, nrgbaSuccess := img.(*image.NRGBA); nrgbaSuccess {
		rgbaOrNrgba = true
		source = nrgba.Pix
	} else if rgba, rgbaSuccess := img.(*image.RGBA); rgbaSuccess {
		rgbaOrNrgba = true
		source = rgba.Pix
	}
	if rgbaOrNrgba {
		raster = make([]byte, size.X*size.Y)
		srcIndex := 0
		grayIndex := 0
		for y := 0; y < size.Y; y++ {
			for x := 0; x < size.X; x++ {
				r := int(source[srcIndex])
				srcIndex++
				g := int(source[srcIndex])
				srcIndex++
				b := int(source[srcIndex])
				srcIndex++
				// we don't use alpha
				// a := source[srcIndex]
				srcIndex++
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
			return nil, fmt.Errorf("expected NRGBA or Gray, but was %v", reflect.TypeOf(img))
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
	LocalNow() time.Time
	UtcNow() time.Time
}

type locationTimeProvider struct {
	location *time.Location
}

func (o *locationTimeProvider) LocalNow() time.Time {
	return time.Now().In(o.location)
}

func (o *locationTimeProvider) UtcNow() time.Time {
	return time.Now().UTC()
}

type testTimeProvider struct {
	offset time.Duration
}

func (o *testTimeProvider) LocalNow() time.Time {
	return time.Now().UTC().Add(o.offset)
}

func (o *testTimeProvider) UtcNow() time.Time {
	return time.Now().UTC()
}

func NewTestTimeProvider(startTime time.Time) TimeProvider {
	now := time.Now()
	_, tzOffset := now.Zone()
	testOffset := int(startTime.Sub(now).Seconds())
	return &testTimeProvider{time.Duration(tzOffset+testOffset) * time.Second}
}

func NewTimeProvider() TimeProvider {
	return &locationTimeProvider{time.Local}
}

func DrawImage(targetImage []byte, targetImageSize image.Point, targetOffset image.Point, sourceImage []byte, sourceImageSize image.Point) {
	if targetOffset.X+sourceImageSize.X > targetImageSize.X || targetOffset.Y+sourceImageSize.Y > targetImageSize.Y {
		panic("images don't overlap fully")
	}
	if len(targetImage) != targetImageSize.X*targetImageSize.Y || len(sourceImage) != sourceImageSize.X*sourceImageSize.Y {
		fmt.Printf("sourceImageSize %v, expectedLen %d, actualLen %d\n", sourceImageSize, len(sourceImage), sourceImageSize.X*sourceImageSize.Y)
		fmt.Printf("targetImageSize %v, expectedLen %d, actualLen %d\n", targetImageSize, len(targetImage), targetImageSize.X*targetImageSize.Y)
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
