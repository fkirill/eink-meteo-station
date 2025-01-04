package clock

import (
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"github.com/rotisserie/eris"
	"image"
	"path"
	"strconv"
	"time"
)

type ClockRenderable interface {
	renderable.Renderable
}

func NewClockRenderable(rect image.Rectangle, provider utils.TimeProvider) (ClockRenderable, error) {
	rasterSize := rect.Dx() * rect.Dy()
	raster := make([]byte, rasterSize, rasterSize)
	res := &clockRenderable{
		offset:         rect.Min,
		size:           rect.Size(),
		raster:         raster,
		nextRedrawTime: provider.UtcNow(),
		// unrealistic values, will be reset upon first render
		hour:         70,
		minute:       70,
		second:       70,
		timeProvider: provider,
	}
	err := res.loadNumbersAndColon()
	if err != nil {
		return nil, eris.Wrapf(err, "Error loading images")
	}
	res.drawColons()
	return res, nil
}

type clockRenderable struct {
	// clockWidgetSize: 963x237
	digitImages [][]byte

	// digitsSize: 241x237
	digitsImageSize []image.Point

	// colonSize: 120x237
	colonSize            image.Point
	colonImage           []byte
	offset               image.Point
	size                 image.Point
	raster               []byte
	nextRedrawTime       time.Time
	hour, minute, second int
	timeProvider         utils.TimeProvider
}

func (c *clockRenderable) loadNumbersAndColon() error {
	if c.digitImages != nil {
		return nil
	}
	c.digitImages = make([][]byte, 60, 60)
	c.digitsImageSize = make([]image.Point, 60, 60)
	for i := 0; i < 60; i++ {
		imageNamePart := strconv.Itoa(i)
		if len(imageNamePart) == 1 {
			imageNamePart = "0" + imageNamePart
		}
		img, err := utils.LoadImage(path.Join(utils.GetRootDir(), "numbers/num_"+imageNamePart+".png"))
		if err != nil {
			return err
		}
		c.digitsImageSize[i] = img.Bounds().Max
		c.digitImages[i], err = utils.ConvertToGrayScale(img)
		if err != nil {
			return err
		}
	}
	colonImg, err := utils.LoadImage(path.Join(utils.GetRootDir(), "numbers/colon.png"))
	if err != nil {
		return err
	}
	c.colonSize = colonImg.Bounds().Size()
	c.colonImage, err = utils.ConvertToGrayScale(colonImg)
	if err != nil {
		return err
	}
	return nil
}

func (c *clockRenderable) drawColons() {
	utils.DrawImage(c.raster, c.size, image.Point{X: c.digitsImageSize[0].X}, c.colonImage, c.colonSize)
	utils.DrawImage(c.raster, c.size, image.Point{X: c.digitsImageSize[0].X*2 + c.colonSize.X}, c.colonImage, c.colonSize)
}

func (c *clockRenderable) RedrawNow() {
	c.nextRedrawTime = c.timeProvider.UtcNow()
}

func (_ *clockRenderable) String() string {
	return "clock"
}

func (c *clockRenderable) DisplayMode() uint8 {
	return clib.A2_Mode
}

func (c *clockRenderable) BoundingBox() image.Rectangle {
	return image.Rectangle{Min: c.offset, Max: image.Point{X: c.offset.X + c.size.X, Y: c.offset.Y + c.size.Y}}
}

func (c *clockRenderable) Offset() image.Point {
	return c.offset
}

func (c *clockRenderable) Size() image.Point {
	return c.size
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
		c.drawNumber(c.digitsImageSize[0].X+c.colonSize.X, nextMinute)
	}
	if c.second != nextSecond {
		c.drawNumber(c.digitsImageSize[0].X*2+c.colonSize.X*2, nextSecond)
	}
	c.hour = nextHour
	c.minute = nextMinute
	c.second = nextSecond
	return nil
}

func (c *clockRenderable) drawNumber(xOffset int, number int) {
	utils.DrawImage(c.raster, c.Size(), image.Point{X: xOffset, Y: 0}, c.digitImages[number], c.digitsImageSize[number])
}
