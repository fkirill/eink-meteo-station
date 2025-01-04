package calendar

import (
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/config"
	"fkirill.org/eink-meteo-station/puppettier"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"github.com/rotisserie/eris"
	"html/template"
	"image"
	_ "image/png"
	"strconv"
	"time"
)

type CalendarRenderable interface {
	renderable.Renderable
}

func NewCalendarRenderable(
	rect image.Rectangle,
	provider utils.TimeProvider,
	cfg config.ConfigApi,
) CalendarRenderable {
	currentMonthHtmlTemplate, err := template.New("currentMonthHtml").Parse(currentMonthHtmlTemplateText)
	if err != nil {
		panic(eris.ToString(err, true))
	}
	return &calendarRenderable{
		offset:                   rect.Min,
		size:                     rect.Size(),
		nextRedrawTime:           provider.UtcNow().AddDate(0, 0, -1),
		cachedRaster:             nil,
		timeProvider:             provider,
		currentMonthHtmlTemplate: currentMonthHtmlTemplate,
		cfg:                      cfg,
	}
}

type calendarRenderable struct {
	cfg                      config.ConfigApi
	offset                   image.Point
	size                     image.Point
	nextRedrawTime           time.Time
	cachedRaster             []byte
	timeProvider             utils.TimeProvider
	currentMonthHtmlTemplate *template.Template
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
	html, err := r.renderCurrentMonthHtml(now.Year(), now.Month(), now.Day())
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

func (r *calendarRenderable) renderCurrentMonthHtml(year int, month time.Month, currentDay int) (string, error) {
	return renderCurrentMonth(year, month, currentDay, r.cfg.GetSpecialDays(), r.currentMonthHtmlTemplate)
}
