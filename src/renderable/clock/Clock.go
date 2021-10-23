package clock

import (
	"image"
	"image/color"
	"image/draw"
	"renderable"
	"renderable/utils"
	"strconv"
	"time"
)

// clockWidgetSize: 963x237
var clockWidgetSize = image.Point{}
var digitImages [][]byte = nil

// digitsSize: 241x237
var digitsImageSize = image.Point{}

// colonSize: 120x237
var colonSize = image.Point{}
var initialRaster []byte = nil

func loadImages() error {
	if digitImages != nil {
		return nil
	}
	digitImages = make([][]byte, 60, 60)
	for i := 0; i < 60; i++ {
		imageNamePart := strconv.Itoa(i)
		if len(imageNamePart) == 1 {
			imageNamePart = "0" + imageNamePart
		}
		img, err := utils.LoadImage("numbers/num_" + imageNamePart + ".png")
		if err != nil {
			return err
		}
		digitsImageSize = img.Bounds().Max
		digitImages[i], err = utils.ConvertToGrayScale(img)
		if err != nil {
			return err
		}
	}
	colonImg, err := utils.LoadImage("numbers/colon.png")
	if err != nil {
		return err
	}
	colonSize = colonImg.Bounds().Size()
	clockWidgetSize = image.Point{X: digitsImageSize.X*3 + colonSize.X*2, Y: colonSize.Y}

	// create raster with two colons on
	rasterImage := image.NewGray(image.Rectangle{Max: clockWidgetSize})
	// fill in white
	draw.Draw(rasterImage, rasterImage.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	initialRaster, err = utils.ConvertToGrayScale(rasterImage)
	if err != nil {
		return err
	}

	colonImage, err := utils.ConvertToGrayScale(colonImg)
	if err != nil {
		return err
	}

	utils.DrawImage(initialRaster, clockWidgetSize, image.Point{X: digitsImageSize.X}, colonImage, colonSize)
	utils.DrawImage(initialRaster, clockWidgetSize, image.Point{X: digitsImageSize.X*2+ colonSize.X}, colonImage, colonSize)

	return nil
}

func NewClockRenderable(offset image.Point, provider utils.TimeProvider) (renderable.Renderable, error) {
	err := loadImages()
	if err != nil {
		return nil, err
	}
	raster := make([]byte, len(initialRaster), len(initialRaster))
	copy(raster, initialRaster)
	return &clockRenderable{
		offset:         offset,
		raster:         raster,
		nextRedrawTime: provider.UtcNow(),
		// unrealistic values, will be reset upon first render
		hour:         70,
		minute:       70,
		second:       70,
		timeProvider: provider,
	}, nil
}

type clockRenderable struct {
	offset               image.Point
	raster               []byte
	nextRedrawTime       time.Time
	hour, minute, second int
	timeProvider         utils.TimeProvider
}

func (c *clockRenderable) RedrawNow() {
	c.nextRedrawTime = c.timeProvider.UtcNow()
}

func (_ *clockRenderable) String() string {
	return "clock"
}

func (c *clockRenderable) DisplayMode() int {
	return 1
}

func (c *clockRenderable) BoundingBox() image.Rectangle {
	return image.Rectangle{Min: c.offset, Max: image.Point{X: c.offset.X + clockWidgetSize.X, Y: c.offset.Y + clockWidgetSize.Y}}
}

func (c *clockRenderable) Offset() image.Point {
	return c.offset
}

func (c *clockRenderable) Size() image.Point {
	return clockWidgetSize
}

func (c *clockRenderable) Raster() []byte {
	return c.raster
}

func (c *clockRenderable) NextRedrawDateTimeUtc() time.Time {
	return c.nextRedrawTime
}

func (c *clockRenderable) RedrawFinished() {
	now := c.timeProvider.UtcNow().Truncate(time.Second)
	c.nextRedrawTime = now.Add(time.Second)
}

func (c *clockRenderable) Render() error {
	now := c.timeProvider.LocalNow()
	nextHour := now.Hour()
	nextMinute := now.Minute()
	nextSecond := now.Second()
	if c.hour != nextHour {
		c.drawNumber(0, nextHour)
	}
	if c.minute != nextMinute {
		c.drawNumber(digitsImageSize.X+colonSize.X, nextMinute)
	}
	if c.second != nextSecond {
		c.drawNumber(digitsImageSize.X*2+colonSize.X*2, nextSecond)
	}
	c.hour = nextHour
	c.minute = nextMinute
	c.second = nextSecond
	return nil
}

func (c *clockRenderable) drawNumber(xOffset int, number int) {
	utils.DrawImage(c.raster, c.Size(), image.Point{X: xOffset, Y:0}, digitImages[number], digitsImageSize)
}
