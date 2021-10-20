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
	NextRedrawDateTime() time.Time
	RedrawFinished()
	Render() error
	DisplayMode() int
	String() string
	RedrawNow()
}
