package renderable

import (
	"image"
	"time"
)

type Renderable interface {
	BoundingBox() image.Rectangle
	Offset() image.Point
	Size() image.Point
	Raster() []byte
	NextRedrawDateTimeUtc() time.Time
	RedrawFinished()
	Render() error
	DisplayMode() uint8
	String() string
	RedrawNow()
}
