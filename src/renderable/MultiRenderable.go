package renderable

import (
	"errors"
	"image"
	"renderable/utils"
	"time"
)

func NewMultiRenderable(offset image.Point, size image.Point, renderables []Renderable, startWithBlackScreen bool) (Renderable, error) {
	if offset.X < 0 || offset.Y < 0 || size.X <= 0 || size.Y <= 0 {
		return nil, errors.New("offset coordinates must be positive, size dimentions must be non-negative")
	}
	if len(renderables) == 0 {
		return nil, errors.New("renderables must be non-empty")
	}
	bbox := utils.BoundingBox(offset, size)
	for _, r := range renderables {
		if !r.BoundingBox().In(bbox) {
			return nil, errors.New("all renderables must be contained in the multiRenderable bounds")
		}
	}
	filler := byte(0xff)
	if startWithBlackScreen {
		filler = 0
	}
	raster := make([]byte, size.X*size.Y, size.X*size.Y)
	if filler != 0 {
		for index := range raster {
			raster[index] = filler
		}
	}
	return &multiRenderable{offset: offset, size: size, renderables: renderables, raster: raster, renderCalcPending: true}, nil
}

type multiRenderable struct {
	offset            image.Point
	size              image.Point
	raster            []byte
	renderables       []Renderable
	toRender          []Renderable
	nextRenderTime    time.Time
	renderCalcPending bool
}

func (m *multiRenderable) RedrawNow() {
	for i := range m.raster {
		m.raster[i] = 0xff
	}
	for _, r := range m.renderables {
		r.RedrawNow()
	}
}

func (_ *multiRenderable) String() string{
	return "multi-renderable"
}

func (m *multiRenderable) DisplayMode() int {
	m.maybeCalculateWhatToRerender()
	displayMode := -1
	for _, r := range m.toRender {
		if displayMode < r.DisplayMode() {
			displayMode = r.DisplayMode()
		}
	}
	return displayMode
}

func (m *multiRenderable) Offset() image.Point {
	return m.offset
}

func (m *multiRenderable) BoundingBox() image.Rectangle {
	return utils.BoundingBox(m.offset, m.size)
}

func (m *multiRenderable) Size() image.Point {
	return m.size
}

func (m *multiRenderable) Raster() []byte {
	return m.raster
}

func (m *multiRenderable) maybeCalculateWhatToRerender() {
	if !m.renderCalcPending {
		return
	}
	minTime := m.renderables[0].NextRedrawDateTimeUtc()
	for _, r := range m.renderables {
		redrawTime := r.NextRedrawDateTimeUtc()
		if minTime.After(redrawTime) {
			minTime = redrawTime
		}
	}
	m.nextRenderTime = minTime
	toReRender := make([]Renderable, 0)
	for _, r := range m.renderables {
		if r.NextRedrawDateTimeUtc() == minTime {
			toReRender = append(toReRender, r)
		}
	}
	m.toRender = toReRender
	m.renderCalcPending = false
}

func (m *multiRenderable) NextRedrawDateTimeUtc() time.Time {
	m.maybeCalculateWhatToRerender()
	return m.nextRenderTime
}

func (m *multiRenderable) RedrawFinished() {
	m.maybeCalculateWhatToRerender()
	for _, r := range m.toRender {
		r.RedrawFinished()
	}
	m.renderCalcPending = true
}

func (m *multiRenderable) Render() error {
	m.maybeCalculateWhatToRerender()
	errs := make([]error, 0)
	for _, r := range m.toRender {
		err := r.Render()
		if err != nil {
			errs = append(errs, err)
		} else {
			m.copyRasterFrom(r)
		}
	}
	if len(errs) != 0 {
		errText := "Errors detected during render: "
		for i, err := range errs {
			if i > 0 {
				errText += ", "
			}
			errText += err.Error()
		}
		return errors.New(errText)
	}
	return nil
}

func (m *multiRenderable) copyRasterFrom(r Renderable) {
	offset := r.Offset()
	lineWidth := m.size.X
	initIndex := offset.Y * lineWidth + offset.X
	sizeToCopy := r.Size()
	sourceRaster := r.Raster()
	for i := 0; i < sizeToCopy.Y; i++ {
		srcOffset := i*sizeToCopy.X
		dstOffset := initIndex + (i * lineWidth)
		width := sizeToCopy.X - 1
		// one complete row from source
		src := sourceRaster[srcOffset : srcOffset + width - 1]
		// map onto the part of the row in our raster
		dest := m.raster[dstOffset : dstOffset + width - 1]
		copy(dest, src)
	}
}
