package calendar

import (
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/puppettier"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"image"
	_ "image/png"
	"strconv"
	"time"
)

func NewCalendarRenderable(offset image.Point, size image.Point, provider utils.TimeProvider) renderable.Renderable {
	return &calendarRenderable{
		offset:         offset,
		size:           size,
		nextRedrawTime: provider.UtcNow().AddDate(0, 0, -1),
		timeProvider:   provider,
	}
}

type calendarRenderable struct {
	offset         image.Point
	size           image.Point
	nextRedrawTime time.Time
	cachedRaster   []byte
	timeProvider   utils.TimeProvider
}

func (c *calendarRenderable) RedrawNow() {
	c.nextRedrawTime = c.timeProvider.UtcNow()
}

func (_ *calendarRenderable) String() string {
	return "calendar"
}

func (r *calendarRenderable) DisplayMode() uint8 {
	return clib.GC16_Mode
}

func (r *calendarRenderable) Offset() image.Point {
	return r.offset
}

func (r *calendarRenderable) BoundingBox() image.Rectangle {
	return utils.BoundingBox(r.offset, r.size)
}

func (r *calendarRenderable) Size() image.Point {
	return r.size
}

func (r *calendarRenderable) NextRedrawDateTimeUtc() time.Time {
	return r.nextRedrawTime
}

func (r *calendarRenderable) Area() int {
	size := r.Size()
	return size.X * size.Y
}

func (r *calendarRenderable) RedrawFinished() {
	r.nextRedrawTime = r.timeProvider.LocalNow().Truncate(24*time.Hour).AddDate(0, 0, 1).UTC()
}

func (r *calendarRenderable) Raster() []byte {
	return r.cachedRaster
}

func (r *calendarRenderable) Render() error {
	now := r.timeProvider.LocalNow()
	html, err := RenderCurrentMonthHtml(now.Year(), now.Month(), now.Day())
	if err != nil {
		return err
	}
	filePrefix := "calendar_" + strconv.FormatInt(now.UnixNano(), 16)
	raster, err2 := puppettier.RenderInPuppeteer(html, filePrefix, r.size)
	if err2 != nil {
		return err2
	}
	r.cachedRaster = raster
	return nil
}
