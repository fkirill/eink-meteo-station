package eink

type EinkScreenProvider interface {
	NewEinkScreen(vcom float64) (EInkScreen, error)
}

type einkScreenProvider struct {
}

func (e einkScreenProvider) NewEinkScreen(vcom float64) (EInkScreen, error) {
	return NewEInkScreen(vcom)
}

func NewEinkScreenProvider() EinkScreenProvider {
	return &einkScreenProvider{}
}
