package calendar

import (
	"image"
	_ "image/png"
	"renderable"
	"renderable/puppettier"
	"renderable/utils"
	"strconv"
	"time"
)

func NewCalendarRenderable(offset image.Point, size image.Point, provider utils.TimeProvider) renderable.Renderable {
	return &calendarRenderable{
		offset: offset,
		size: size,
		nextRefrawTime: provider.Now().Truncate(24 * time.Hour),
		timeProvider: provider,
	}
}

type calendarRenderable struct {
	offset         image.Point
	size           image.Point
	nextRefrawTime time.Time
	cachedRaster   []byte
	timeProvider   utils.TimeProvider
}

func (_ *calendarRenderable) String() string {
	return "calendar"
}

func (r *calendarRenderable) DisplayMode() int {
	return 2
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

func (r *calendarRenderable) NextRedrawDateTime() time.Time {
	return r.nextRefrawTime
}

func (r *calendarRenderable) Area() int {
	size := r.Size()
	return size.X * size.Y
}

func (r *calendarRenderable) RedrawFinished() {
	r.nextRefrawTime = r.timeProvider.Now().Truncate(24*time.Hour).Add(24*time.Hour)
}

func (r *calendarRenderable) Raster() []byte {
	return r.cachedRaster
}

func (r *calendarRenderable) Render() error {
	// generate html
	// call puppeteer to render the html into png
	// load png
	// convert to 16 gray-scale image
	// return raster (and cache of course)
	now := r.timeProvider.Now()
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
