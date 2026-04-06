//go:build sdl

package sdlrender

import (
	"context"
	"runtime"
	"time"

	"github.com/rensjaspers/bookzo-vacancy-sign/internal/render"
	"github.com/rensjaspers/bookzo-vacancy-sign/internal/vacancy"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Config struct {
	FontPath   string
	Fullscreen bool
	Title      string
	Width      int
	Height     int
}

type Renderer struct {
	config Config
	fonts  map[int]*ttf.Font
}

type layout struct {
	width        int32
	height       int32
	hotelSize    int
	headlineSize int
	sublineSize  int
	hintY        int32
}

func New(config Config) *Renderer {
	return &Renderer{config: config, fonts: map[int]*ttf.Font{}}
}

func (r *Renderer) Run(ctx context.Context, source render.SnapshotSource) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return r.runLocked(ctx, source)
}

func (r *Renderer) runLocked(ctx context.Context, source render.SnapshotSource) error {
	if err := initSDL(); err != nil {
		return err
	}
	defer quitSDL()
	window, canvas, err := r.newWindowAndCanvas()
	if err != nil {
		return err
	}
	defer r.closeWindow(window, canvas)
	return r.loop(ctx, canvas, source)
}

func initSDL() error {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return err
	}
	return ttf.Init()
}

func quitSDL() {
	ttf.Quit()
	sdl.Quit()
}

func (r *Renderer) newWindowAndCanvas() (*sdl.Window, *sdl.Renderer, error) {
	window, err := sdl.CreateWindow(r.config.Title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(r.config.Width), int32(r.config.Height), r.windowFlags())
	if err != nil {
		return nil, nil, err
	}
	canvas, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, nil, err
	}
	return window, canvas, nil
}

func (r *Renderer) windowFlags() uint32 {
	if r.config.Fullscreen {
		return sdl.WINDOW_SHOWN | sdl.WINDOW_FULLSCREEN_DESKTOP
	}
	return sdl.WINDOW_SHOWN
}

func (r *Renderer) closeWindow(window *sdl.Window, canvas *sdl.Renderer) {
	r.closeFonts()
	canvas.Destroy()
	window.Destroy()
}

func (r *Renderer) closeFonts() {
	for _, font := range r.fonts {
		font.Close()
	}
}

func (r *Renderer) loop(
	ctx context.Context,
	canvas *sdl.Renderer,
	source render.SnapshotSource,
) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	last := vacancy.ViewModel{}
	seen := false
	for {
		if quit := pollQuit(); quit {
			return nil
		}
		current := source.Snapshot(time.Now())
		if err := r.drawIfNeeded(canvas, current, last, seen); err != nil {
			return err
		}
		last, seen = current, true
		if done := waitFrame(ctx, ticker); done {
			return nil
		}
	}
}

func pollQuit() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		if isQuitEvent(event) {
			return true
		}
	}
	return false
}

func isQuitEvent(event sdl.Event) bool {
	switch value := event.(type) {
	case *sdl.QuitEvent:
		return true
	case *sdl.KeyboardEvent:
		return value.Type == sdl.KEYDOWN && value.Keysym.Sym == sdl.K_ESCAPE
	default:
		return false
	}
}

func (r *Renderer) drawIfNeeded(
	canvas *sdl.Renderer,
	current vacancy.ViewModel,
	last vacancy.ViewModel,
	seen bool,
) error {
	if seen && current.Equal(last) {
		return nil
	}
	return r.draw(canvas, current)
}

func waitFrame(ctx context.Context, ticker *time.Ticker) bool {
	select {
	case <-ctx.Done():
		return true
	case <-ticker.C:
		return false
	case <-time.After(50 * time.Millisecond):
		return false
	}
}

func (r *Renderer) draw(canvas *sdl.Renderer, vm vacancy.ViewModel) error {
	box, err := r.layoutFor(canvas, vm.HeadlineScale)
	if err != nil {
		return err
	}
	r.fill(canvas, backgroundColor(vm))
	if err := r.drawTexts(canvas, vm, box); err != nil {
		return err
	}
	canvas.Present()
	return nil
}

func (r *Renderer) layoutFor(canvas *sdl.Renderer, scale int) (layout, error) {
	width, height, err := canvas.GetOutputSize()
	if err != nil {
		return layout{}, err
	}
	return newLayout(width, height, scale), nil
}

func newLayout(width int32, height int32, scale int) layout {
	return layout{
		width:        width,
		height:       height,
		hotelSize:    calcHotelSize(width, height, scale),
		headlineSize: calcHeadlineSize(width, height, scale),
		sublineSize:  calcSublineSize(width, height, scale),
		hintY:        height - 80,
	}
}

func calcHotelSize(width int32, height int32, scale int) int {
	return minInt(int(float64(width)*0.05), int(float64(height)*0.10)) * scale / 100
}

func calcHeadlineSize(width int32, height int32, scale int) int {
	return minInt(int(float64(width)*0.14), int(float64(height)*0.28)) * scale / 100
}

func calcSublineSize(width int32, height int32, scale int) int {
	return minInt(int(float64(width)*0.06), int(float64(height)*0.12)) * scale / 100
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func (r *Renderer) drawTexts(
	canvas *sdl.Renderer,
	vm vacancy.ViewModel,
	box layout,
) error {
	headlineCands := vm.HeadlineCandidates
	if len(headlineCands) == 0 {
		headlineCands = []string{vm.Headline}
	}
	headlineMin := maxInt(box.headlineSize*3/5, box.sublineSize)
	headlinePixelSize, err := r.minUnifyFontSize(
		headlineCands, box.headlineSize, headlineMin, box.width*9/10, true)
	if err != nil {
		return err
	}
	hotelPixelSize, err := r.minUnifyFontSize(
		[]string{vm.HotelName}, box.hotelSize, box.hotelSize*3/4, box.width*3/5, false)
	if err != nil {
		return err
	}
	hotel, err := r.hotelSurfaceAtFixed(vm.HotelName, box, hotelColor(vm), hotelPixelSize)
	if err != nil {
		return err
	}
	defer hotel.Free()
	headline, err := r.headlineSurfaceAtFixed(vm.Headline, box, headlineColor(vm), headlinePixelSize)
	if err != nil {
		return err
	}
	defer headline.Free()
	var subline *sdl.Surface
	if vm.ShowSubline {
		subCands := vm.SublineCandidates
		if len(subCands) == 0 {
			subCands = []string{vm.Subline}
		}
		subPixelSize, err := r.minUnifyFontSize(
			subCands, box.sublineSize, box.sublineSize*4/5, box.width*4/5, false)
		if err != nil {
			return err
		}
		subline, err = r.sublineSurfaceAtFixed(vm.Subline, box, headlineColor(vm), subPixelSize)
		if err != nil {
			return err
		}
	}
	defer freeSurface(subline)
	return r.drawTextGroup(canvas, vm, box, hotel, headline, subline)
}

func (r *Renderer) drawTextGroup(
	canvas *sdl.Renderer,
	vm vacancy.ViewModel,
	box layout,
	hotel *sdl.Surface,
	headline *sdl.Surface,
	subline *sdl.Surface,
) error {
	hotelY, headlineY, sublineY := textGroupPositions(box, hotel, headline, subline)
	if err := r.copyText(canvas, hotel, hotelY); err != nil {
		return err
	}
	if err := r.copyText(canvas, headline, headlineY); err != nil {
		return err
	}
	if vm.ShowSubline {
		if err := r.copyText(canvas, subline, sublineY); err != nil {
			return err
		}
	}
	return r.drawHint(canvas, vm, box)
}

func (r *Renderer) drawHint(
	canvas *sdl.Renderer,
	vm vacancy.ViewModel,
	box layout,
) error {
	if !vm.ShowErrorHint {
		return nil
	}
	return r.drawCentered(canvas, "!", 32, box.hintY, hintColor(vm))
}

func (r *Renderer) drawCentered(
	canvas *sdl.Renderer,
	text string,
	size int,
	y int32,
	color sdl.Color,
) error {
	surface, err := r.singleLineSurface(text, size, color)
	if err != nil {
		return err
	}
	defer surface.Free()
	return r.copyText(canvas, surface, y)
}

func (r *Renderer) font(size int) (*ttf.Font, error) {
	if font, ok := r.fonts[size]; ok {
		return font, nil
	}
	font, err := ttf.OpenFont(r.config.FontPath, size)
	if err != nil {
		return nil, err
	}
	r.fonts[size] = font
	return font, nil
}

func (r *Renderer) singleLineSurface(
	text string,
	size int,
	color sdl.Color,
) (*sdl.Surface, error) {
	font, err := r.font(size)
	if err != nil {
		return nil, err
	}
	return font.RenderUTF8Blended(text, color)
}

func (r *Renderer) centeredWrappedSurface(
	text string,
	size int,
	maxWidth int32,
	color sdl.Color,
) (*sdl.Surface, error) {
	font, err := r.font(size)
	if err != nil {
		return nil, err
	}
	lines, err := wrapLines(font, text, int(maxWidth))
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return r.singleLineSurface("", size, color)
	}
	if len(lines) == 1 {
		return font.RenderUTF8Blended(lines[0], color)
	}
	return stackCenteredLines(font, lines, maxWidth, color)
}

func stackCenteredLines(
	font *ttf.Font,
	lines []string,
	maxWidth int32,
	color sdl.Color,
) (*sdl.Surface, error) {
	surfaces := make([]*sdl.Surface, 0, len(lines))
	for _, line := range lines {
		s, err := font.RenderUTF8Blended(line, color)
		if err != nil {
			for _, x := range surfaces {
				x.Free()
			}
			return nil, err
		}
		surfaces = append(surfaces, s)
	}
	defer func() {
		for _, s := range surfaces {
			s.Free()
		}
	}()
	return blitCenteredStack(surfaces, maxWidth)
}

func blitCenteredStack(surfaces []*sdl.Surface, maxWidth int32) (*sdl.Surface, error) {
	h := int32(0)
	gap := int32(2)
	for _, s := range surfaces {
		h += s.H + gap
	}
	h -= gap
	dst, err := sdl.CreateRGBSurfaceWithFormat(0, maxWidth, h, 32, sdl.PIXELFORMAT_RGBA8888)
	if err != nil {
		return nil, err
	}
	if err := dst.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		dst.Free()
		return nil, err
	}
	pix := sdl.MapRGBA(dst.Format, 0, 0, 0, 0)
	dst.FillRect(nil, pix)
	y := int32(0)
	for _, s := range surfaces {
		x := (maxWidth - s.W) / 2
		rect := &sdl.Rect{X: x, Y: y, W: s.W, H: s.H}
		if err := s.Blit(nil, dst, rect); err != nil {
			dst.Free()
			return nil, err
		}
		y += s.H + gap
	}
	return dst, nil
}

func nextSize(current int, minSize int, baseSize int) int {
	next := current - maxInt(2, baseSize/12)
	if next < minSize {
		return minSize
	}
	return next
}

func (r *Renderer) copyText(
	canvas *sdl.Renderer,
	surface *sdl.Surface,
	y int32,
) error {
	texture, err := canvas.CreateTextureFromSurface(surface)
	if err != nil {
		return err
	}
	defer texture.Destroy()
	width, _, err := canvas.GetOutputSize()
	if err != nil {
		return err
	}
	return canvas.Copy(texture, nil, centeredRect(surface, y, width))
}

func centeredRect(surface *sdl.Surface, y int32, width int32) *sdl.Rect {
	return &sdl.Rect{
		X: (width - surface.W) / 2,
		Y: y,
		W: surface.W,
		H: surface.H,
	}
}

func textGroupPositions(
	box layout,
	hotel *sdl.Surface,
	headline *sdl.Surface,
	subline *sdl.Surface,
) (int32, int32, int32) {
	gap := groupGap(box.height)
	total := hotel.H + gap + headline.H
	if subline != nil {
		total += gap + subline.H
	}
	start := max32((box.height-total)/2, 24)
	hotelY := start
	headlineY := hotelY + hotel.H + gap
	sublineY := headlineY + headline.H + gap
	return hotelY, headlineY, sublineY
}

func groupGap(height int32) int32 {
	return min32(max32(height/24, 18), 48)
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func max32(left int32, right int32) int32 {
	if left > right {
		return left
	}
	return right
}

func min32(left int32, right int32) int32 {
	if left < right {
		return left
	}
	return right
}

func freeSurface(surface *sdl.Surface) {
	if surface != nil {
		surface.Free()
	}
}

func (r *Renderer) fill(canvas *sdl.Renderer, color sdl.Color) {
	canvas.SetDrawColor(color.R, color.G, color.B, color.A)
	canvas.Clear()
}

func backgroundColor(vm vacancy.ViewModel) sdl.Color {
	if vm.Dark {
		return sdl.Color{R: 18, G: 18, B: 22, A: 255}
	}
	return sdl.Color{R: 248, G: 248, B: 244, A: 255}
}

func hotelColor(vm vacancy.ViewModel) sdl.Color {
	if vm.Dark {
		return sdl.Color{R: 156, G: 163, B: 175, A: 255}
	}
	return sdl.Color{R: 107, G: 114, B: 128, A: 255}
}

func headlineColor(vm vacancy.ViewModel) sdl.Color {
	if vm.Available {
		return availableColor(vm)
	}
	return fullColor(vm)
}

func availableColor(vm vacancy.ViewModel) sdl.Color {
	if vm.Dark {
		return sdl.Color{R: 52, G: 211, B: 153, A: 255}
	}
	return sdl.Color{R: 6, G: 95, B: 70, A: 255}
}

func fullColor(vm vacancy.ViewModel) sdl.Color {
	if vm.Dark {
		return sdl.Color{R: 120, G: 113, B: 108, A: 255}
	}
	return sdl.Color{R: 87, G: 83, B: 78, A: 255}
}

func hintColor(vm vacancy.ViewModel) sdl.Color {
	if vm.Dark {
		return sdl.Color{R: 217, G: 249, B: 157, A: 255}
	}
	return sdl.Color{R: 120, G: 53, B: 15, A: 255}
}
