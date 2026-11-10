package gui

import (
	"math"
	"time"

	"cairn/internal/config"
)

// Pose is how far one thing of the board has come at an instant of its intro: how much of it is seen, how far it is moved, grown, stretched and turned about its pivot — a share of its box — what light or shadow lies over it, how far it is lifted and blurred, whether it flies free of its column, and the draft it is still drawn as.
type Pose struct {
	Seen, DX, DY, Grow, Wide, High, Turn float64
	PivotX, PivotY                       float64
	Glare, Shade, Lift, Soft             float64
	Free                                 bool
	Draft                                *Draft
}

// Draft is a thing being drawn as a plan before it is inked: how much of its fill is in, how far its stripe is drawn down, how many letters of its title are typed, and how much of its facts is seen.
type Draft struct {
	Fill, Stripe, Facts float64
	Typed               int
}

// AtRest is a thing of the board at rest, as it is drawn once the intro is over.
func AtRest() Pose { return Pose{Seen: 1, Grow: 1, Wide: 1, High: 1, PivotX: 0.5, PivotY: 0.5} }

// Resting says whether a pose draws its thing exactly as the still board does.
func (p Pose) Resting() bool {
	return p.Seen == 1 && p.DX == 0 && p.DY == 0 && p.Grow == 1 && p.Wide == 1 && p.High == 1 && p.Turn == 0 &&
		p.Glare == 0 && p.Shade == 0 && p.Lift == 0 && p.Soft == 0 && !p.Free && p.Draft == nil
}

// Stage is what an intro is played on: the window's size, the grounds the cards stand on, the cards seen, in the order the board draws them, the card the intro starts from, and where its focus mark is.
type Stage struct {
	Wide, High int
	Grounds    []Box
	Cards      []Staged
	Lead       int
	Mark       Box
}

// Staged is a card on the stage: its slot, the ground it stands on and its row there, its priority, whether it is in focus, and how many letters its title has.
type Staged struct {
	Slot        Box
	Ground, Row int
	Priority    string
	Focus       bool
	Letters     int
}

// Scene is the board at an instant of its intro: the camera over it all, the top and the bar, each ground and each card, the order the cards flying free are drawn in, and what is laid over the board.
type Scene struct {
	Camera, Top, Bar Pose
	Grounds, Cards   []Pose
	Order            []int
	Wave             *Wave
	Sweep            *Sweep
	Grains           Grains
	Beat             *Pulse
	Stones           bool
	Progress         float64
	ends             float64
}

// Wave is the ring a ripple sends out: its centre, its radius and how much of it is seen.
type Wave struct{ X, Y, Radius, Seen float64 }

// Sweep is the band of light crossing the board: where its middle is, how wide it is, how far it leans, in degrees, and how much of it is seen.
type Sweep struct{ X, Wide, Tilt, Seen float64 }

// Grain is one speck of the dust a card condenses out of: where it is, how big, how much of it is seen, and what it is painted as — a card's stripe in its priority, its title in ink, or its edge.
type Grain struct {
	X, Y, Side, Seen float64
	Part, Priority   string
	Focus            bool
}

// The parts of a card a grain is painted as.
const (
	GrainStripe = "stripe"
	GrainTitle  = "title"
	GrainEdge   = "edge"
)

// Pulse is the focus mark speaking: the mark, how big its dot is and how much of it is seen, and the ring it sends out.
type Pulse struct {
	Mark                         Box
	Dot, DotSeen, Ring, RingSeen float64
}

type introOf struct {
	says   string
	limit  float64
	stages func(*Scene, Stage, float64)
}

var intros = map[string]introOf{
	config.IntroSpring:    {"Spring cascade", 1100, springing},
	config.IntroDeck:      {"Dealt from the deck", 1400, dealing},
	config.IntroAccordion: {"Accordion columns", 950, unfolding},
	config.IntroSweep:     {"Light sweep", 1250, sweeping},
	config.IntroFlap:      {"Split-flap board", 1200, flapping},
	config.IntroPull:      {"Focus pull", 1000, pulling},
	config.IntroRipple:    {"Ripple from the focus", 1100, rippling},
	config.IntroParticles: {"Particles assemble", 1400, assembling},
	config.IntroDolly:     {"Camera dolly", 1400, craning},
	config.IntroBlueprint: {"Blueprint, then ink", 1450, drafting},
	config.IntroStones:    {"Falling stones", 1800, stoning},
}

func introNamed(name string) introOf {
	if one, ok := intros[name]; ok {
		return one
	}
	return intros[config.DefaultIntro]
}

// IntroSays is what an intro is called on screen.
func IntroSays(name string) string { return introNamed(name).says }

// IntroLasts is how long the intro takes on that stage: as long as its last thing takes to land, and never longer than a fifth over what it takes on a board of a few cards, a busier board playing faster.
func IntroLasts(name string, s Stage) time.Duration {
	natural, limit := naturalEnd(name, s), introNamed(name).limit*1.2
	return time.Duration(min(natural, limit) * float64(time.Millisecond))
}

func naturalEnd(name string, s Stage) float64 {
	sc := blank(s)
	introNamed(name).stages(&sc, s, 0)
	return max(1, sc.ends)
}

func blank(s Stage) Scene {
	sc := Scene{Camera: AtRest(), Top: AtRest(), Bar: AtRest(), Grounds: make([]Pose, len(s.Grounds)), Cards: make([]Pose, len(s.Cards))}
	for i := range sc.Grounds {
		sc.Grounds[i] = AtRest()
	}
	for i := range sc.Cards {
		sc.Cards[i] = AtRest()
	}
	return sc
}

// Staging is the board at an instant of an intro, a pure function of the time since the window first drew: at or past the intro's length, the still board.
func Staging(name string, at time.Duration, s Stage) Scene {
	natural, lasts := naturalEnd(name, s), float64(IntroLasts(name, s))/float64(time.Millisecond)
	ms := float64(at) / float64(time.Millisecond)
	sc := blank(s)
	if ms >= lasts {
		return sc
	}
	introNamed(name).stages(&sc, s, ms*natural/lasts)
	sc.Progress = ms / lasts
	return sc
}

func (sc *Scene) reach(at float64) { sc.ends = max(sc.ends, at) }

func phase(sc *Scene, at, delay, lasts float64) float64 {
	sc.reach(delay + lasts)
	return clamp01((at - delay) / lasts)
}

func clamp01(x float64) float64 { return min(1, max(0, x)) }

func mix(a, b, t float64) float64 { return a + (b-a)*t }

func keyed(t float64, at, values []float64) float64 {
	if t <= at[0] {
		return values[0]
	}
	for i := 1; i < len(at); i++ {
		if t <= at[i] {
			return mix(values[i-1], values[i], (t-at[i-1])/(at[i]-at[i-1]))
		}
	}
	return values[len(values)-1]
}

func bezier(x1, y1, x2, y2 float64) func(float64) float64 {
	curve := func(a, b, t float64) float64 { return 3*a*t*(1-t)*(1-t) + 3*b*t*t*(1-t) + t*t*t }
	return func(x float64) float64 {
		if x <= 0 || x >= 1 {
			return clamp01(x)
		}
		lo, hi := 0.0, 1.0
		for range 40 {
			mid := (lo + hi) / 2
			if curve(x1, x2, mid) < x {
				lo = mid
			} else {
				hi = mid
			}
		}
		return curve(y1, y2, (lo+hi)/2)
	}
}

func spring(zeta float64) func(float64) float64 {
	w := 4.6 / zeta
	wd := w * math.Sqrt(1-zeta*zeta)
	return func(t float64) float64 {
		if t >= 1 {
			return 1
		}
		if t <= 0 {
			return 0
		}
		return 1 - math.Exp(-zeta*w*t)*(math.Cos(wd*t)+zeta*w/wd*math.Sin(wd*t))
	}
}

var (
	easeOut    = bezier(0.22, 0.8, 0.3, 1)
	expoOut    = bezier(0.16, 1, 0.3, 1)
	cssEaseOut = bezier(0, 0, 0.58, 1)
	springMid  = spring(0.55)
	springSnap = spring(0.42)
)

func (sc *Scene) chrome(at, delay, lasts float64) {
	e := easeOut(phase(sc, at, delay, lasts))
	sc.Top.Seen, sc.Top.DY = e, -8*(1-e)
	sc.Bar.Seen, sc.Bar.DY = e, 8*(1-e)
}

func (sc *Scene) chromeFades(at, delay, lasts float64) {
	e := phase(sc, at, delay, lasts)
	sc.Top.Seen, sc.Bar.Seen = e, e
}

func (sc *Scene) groundsFade(at float64, from, step, lasts float64) {
	for i := range sc.Grounds {
		sc.Grounds[i].Seen = phase(sc, at, from+float64(i)*step, lasts)
	}
}

func (sc *Scene) beat(s Stage, at, from float64) {
	if s.Mark.W == 0 {
		return
	}
	p := phase(sc, at, from, 520)
	q := phase(sc, at, from+160, 760)
	if p >= 1 && q >= 1 {
		return
	}
	e := easeOut(p)
	b := &Pulse{Mark: s.Mark, Dot: keyed(e, []float64{0, 0.45, 1}, []float64{0, 1.9, 1}), DotSeen: keyed(e, []float64{0, 0.45, 1}, []float64{0, 1, 1})}
	if q > 0 && q < 1 {
		r := bezier(0.2, 0.7, 0.3, 1)(q)
		b.Ring, b.RingSeen = 1+4*r, 0.9*(1-r)
	}
	sc.Beat = b
}

func centre(b Box) (float64, float64) {
	return float64(b.X) + float64(b.W)/2, float64(b.Y) + float64(b.H)/2
}

func (s Stage) lead() (float64, float64) {
	if s.Lead < 0 || s.Lead >= len(s.Cards) {
		return float64(s.Wide) / 2, float64(s.High) / 2
	}
	return centre(s.Cards[s.Lead].Slot)
}

func (s Stage) distances() []float64 {
	cx, cy := s.lead()
	far := 1.0
	out := make([]float64, len(s.Cards))
	for i, c := range s.Cards {
		x, y := centre(c.Slot)
		out[i] = math.Hypot(x-cx, y-cy)
		far = max(far, out[i])
	}
	for i := range out {
		out[i] /= far
	}
	return out
}

func springing(sc *Scene, s Stage, at float64) {
	sc.chrome(at, 0, 380)
	sc.groundsFade(at, 0, 50, 300)
	for i, c := range s.Cards {
		p := phase(sc, at, 90+float64(c.Ground)*70+float64(c.Row)*48, 680)
		if p >= 1 {
			continue
		}
		e := springMid(p)
		sc.Cards[i].Seen = clamp01(p / 0.3)
		sc.Cards[i].DY = 26 * (1 - e)
		sc.Cards[i].Grow = 0.965 + 0.035*e
	}
	sc.beat(s, at, 560)
}

func (s Stage) dealt() []int {
	rows := 0
	for _, c := range s.Cards {
		rows = max(rows, c.Row+1)
	}
	var order []int
	for r := range rows {
		for g := range s.Grounds {
			for i, c := range s.Cards {
				if c.Ground == g && c.Row == r {
					order = append(order, i)
				}
			}
		}
	}
	return order
}

func dealing(sc *Scene, s Stage, at float64) {
	sc.chrome(at, 0, 300)
	order := s.dealt()
	gap := 50.0
	if len(order) > 1 {
		gap = min(gap, (1400-120-430)/float64(len(order)-1))
	}
	deckX, deckY := float64(s.Wide)/2, float64(s.High)*0.8625
	lead := 0.0
	at3 := []float64{0, 0.45, 1}
	for n, i := range order {
		cx, cy := centre(s.Cards[i].Slot)
		dx, dy := deckX-cx, deckY-cy
		turn := float64((n*37)%13 - 6)
		p := phase(sc, at, 120+float64(n)*gap, 430)
		if i == s.Lead {
			lead = float64(n)
		}
		if p >= 1 {
			continue
		}
		e := bezier(0.3, 0.7, 0.25, 1)(p)
		pose := &sc.Cards[i]
		pose.DX = keyed(e, at3, []float64{dx, dx * 0.45, 0})
		pose.DY = keyed(e, at3, []float64{dy - float64(n)*0.8, dy*0.5 - 40, 0})
		pose.Turn = keyed(e, at3, []float64{turn * 0.3, turn, 0})
		pose.Grow = keyed(e, at3, []float64{0.92, 1.04, 1})
		pose.Lift = keyed(e, at3, []float64{0.1, 1, 0})
		pose.Free = true
	}
	for n := len(order) - 1; n >= 0; n-- {
		if sc.Cards[order[n]].Free {
			sc.Order = append(sc.Order, order[n])
		}
	}
	sc.groundsFade(at, 0, 40, 260)
	sc.beat(s, at, 120+lead*gap+380)
}

func unfolding(sc *Scene, s Stage, at float64) {
	sc.chrome(at, 0, 360)
	for i := range sc.Grounds {
		p := phase(sc, at, 60+float64(i)*105, 560)
		if p >= 1 {
			continue
		}
		e := expoOut(p)
		g := &sc.Grounds[i]
		g.PivotX, g.Wide = 0, 0.04+0.96*e
		g.Seen = clamp01(e / 0.2)
		g.Shade = 0.55 * (1 - e)
	}
	for i, c := range s.Cards {
		e := easeOut(phase(sc, at, 60+float64(c.Ground)*105+200+float64(c.Row)*30, 340))
		if e < 1 {
			sc.Cards[i].Seen, sc.Cards[i].DX = e, -18*(1-e)
		}
	}
	sc.beat(s, at, 570)
}

func (s Stage) inFrame(x, y float64) (float64, float64) {
	return x * 1200 / max(1, float64(s.Wide)), y * 800 / max(1, float64(s.High))
}

func sweeping(sc *Scene, s Stage, at float64) {
	const crossing = 1000.0
	tan := math.Tan(12 * math.Pi / 180)
	when := func(x, y float64) float64 {
		fx, fy := s.inFrame(x, y)
		return max(0, (fx-(fy-400)*tan+250)/1920*crossing*0.95)
	}
	scale := float64(s.Wide) / 1200
	if p := phase(sc, at, 0, crossing); p < 1 {
		e := bezier(0.45, 0.05, 0.4, 1)(p)
		sc.Sweep = &Sweep{X: (mix(-420, 1500, e) + 170) * scale, Wide: 340 * scale, Tilt: 12, Seen: 1}
	}
	sc.chromeFades(at, 60, 300)
	for i, g := range s.Grounds {
		sc.Grounds[i].Seen = phase(sc, at, when(float64(g.X)+30, float64(s.High)/2), 300)
	}
	lead := 0.0
	at3 := []float64{0, 0.25, 1}
	for i, c := range s.Cards {
		cx, cy := centre(c.Slot)
		if i == s.Lead {
			lead = when(cx, cy)
		}
		e := easeOut(phase(sc, at, when(float64(c.Slot.X)+20, cy), 420))
		if e >= 1 {
			continue
		}
		sc.Cards[i].Seen = keyed(e, at3, []float64{0, 1, 1})
		sc.Cards[i].Glare = (keyed(e, at3, []float64{2.2, 1.7, 1}) - 1) / 1.2 * 0.3
		sc.Cards[i].DX = -6 * (1 - e)
	}
	sc.beat(s, at, lead+220)
}

func flapping(sc *Scene, s Stage, at float64) {
	sc.chrome(at, 0, 300)
	sc.groundsFade(at, 0, 30, 260)
	fall, settle := bezier(0.55, 0, 0.9, 0.55), bezier(0.3, 0, 0.5, 1)
	for i, c := range s.Cards {
		p := phase(sc, at, 80+float64(c.Row)*105+float64(c.Ground)*45, 540)
		if p >= 1 {
			continue
		}
		var angle, light float64
		switch {
		case p <= 0.55:
			e := fall(p / 0.55)
			angle, light = mix(-92, 14, e), mix(0.35, 1.18, e)
		case p <= 0.8:
			e := settle((p - 0.55) / 0.25)
			angle, light = mix(14, -4, e), mix(1.18, 0.96, e)
		default:
			e := (p - 0.8) / 0.2
			angle, light = mix(-4, 0, e), mix(0.96, 1, e)
		}
		pose := &sc.Cards[i]
		pose.Seen = clamp01(p / 0.08)
		pose.PivotY, pose.High = 0, math.Abs(math.Cos(angle*math.Pi/180))
		pose.Shade, pose.Glare = max(0, 1-light), max(0, light-1)*0.35
	}
	sc.beat(s, at, 590)
}

func pulling(sc *Scene, s Stage, at float64) {
	if e := easeOut(phase(sc, at, 150, 600)); e < 1 {
		sc.Top.Seen, sc.Top.Soft = e, 6*(1-e)
		sc.Bar.Seen, sc.Bar.Soft = e, 6*(1-e)
	}
	cx, cy := s.lead()
	if z := expoOut(phase(sc, at, 0, 1000)); z < 1 {
		sc.Camera.Grow = mix(1.035, 1, z)
		sc.Camera.PivotX, sc.Camera.PivotY = cx/max(1, float64(s.Wide)), cy/max(1, float64(s.High))
	}
	if g := easeOut(phase(sc, at, 0, 650)); g < 1 {
		for i := range sc.Grounds {
			sc.Grounds[i].Seen, sc.Grounds[i].Soft = mix(0.2, 1, g), 8*(1-g)
		}
	}
	far := s.distances()
	for i := range s.Cards {
		delay, lasts := 120+far[i]*280, 600.0
		if i == s.Lead {
			delay, lasts = 0, 420
		}
		if g := easeOut(phase(sc, at, delay, lasts)); g < 1 {
			sc.Cards[i].Seen, sc.Cards[i].Soft = mix(0.25, 1, g), 12*(1-g)
		}
	}
	sc.beat(s, at, 260)
}

func rippling(sc *Scene, s Stage, at float64) {
	cx, cy := s.lead()
	far := s.distances()
	for i := range s.Cards {
		pose := &sc.Cards[i]
		if i == s.Lead {
			if p := phase(sc, at, 0, 520); p < 1 {
				e := springSnap(p)
				pose.Seen, pose.Grow = clamp01(e), 0.9+0.1*e
			}
			continue
		}
		p := phase(sc, at, 170+far[i]*380, 560)
		if p >= 1 {
			continue
		}
		x, y := centre(s.Cards[i].Slot)
		angle := math.Atan2(y-cy, x-cx)
		e := springMid(p)
		pose.Seen = clamp01(e / 0.35)
		pose.DX, pose.DY = -math.Cos(angle)*14*(1-e), -math.Sin(angle)*14*(1-e)
		pose.Grow = 0.88 + 0.12*e
	}
	if s.Mark.W > 0 {
		cx, cy = centre(s.Mark)
	}
	if p := phase(sc, at, 140, 950); p > 0 && p < 1 {
		e := bezier(0.25, 0.6, 0.35, 1)(p)
		sc.Wave = &Wave{X: cx, Y: cy, Radius: 1300 * float64(s.Wide) / 1200 * mix(0.004, 1, e), Seen: keyed(e, []float64{0, 0.4, 1}, []float64{0.9, 0.55, 0})}
	}
	for i, g := range s.Grounds {
		gx, _ := centre(g)
		d := min(1, math.Abs(gx-cx)/900)
		sc.Grounds[i].Seen = phase(sc, at, 120+d*260, 380)
	}
	sc.chrome(at, 380, 380)
	sc.beat(s, at, 80)
}

type grainRand uint64

func (r *grainRand) next() float64 {
	*r = (*r * 16807) % 2147483647
	return float64(*r) / 2147483647
}

func assembling(sc *Scene, s Stage, at float64) {
	n := len(s.Cards)
	step, per := 28.0, 70
	if n > 1 {
		step = min(step, (1400-40-940)/float64(n-1))
	}
	if n > 0 {
		per = max(8, min(70, 1500/n))
	}
	r := grainRand(7)
	ox, oy, scale := float64(s.Wide)/2, float64(s.High)/2, float64(s.Wide)/1200
	lead := 0.0
	for i, c := range s.Cards {
		start := 40 + float64(i)*step
		if i == s.Lead {
			lead = start
		}
		x, y, w, h := float64(c.Slot.X), float64(c.Slot.Y), float64(c.Slot.W), float64(c.Slot.H-gap)
		for j := range per {
			part := GrainEdge
			switch {
			case j < per/7:
				part = GrainStripe
			case j < per*4/7:
				part = GrainTitle
			}
			tx, ty := x+16+r.next()*(w-28), y+12+r.next()*(h-20)
			switch part {
			case GrainStripe:
				tx, ty = x+2, y+8+r.next()*(h-16)
			case GrainTitle:
				ty = y + 10 + r.next()*14
			}
			angle, reach := r.next()*math.Pi*2, (500+r.next()*500)*scale
			sx, sy := ox+math.Cos(angle)*reach, oy+math.Sin(angle)*reach*0.7
			delay, side, swirl := start+r.next()*120, 1.2+r.next()*1.6, (r.next()-0.5)*160
			sc.reach(delay + 860)
			u := clamp01((at - delay) / 680)
			if u <= 0 {
				continue
			}
			e := 1 - (1-u)*(1-u)*(1-u)
			fade := clamp01((at - delay - 600) / 260)
			seen := min(1, u*3) * (1 - fade) * 0.9
			if seen <= 0 {
				continue
			}
			sc.Grains = append(sc.Grains, Grain{
				X: mix(sx, tx, e) + math.Sin(e*math.Pi)*swirl*0.3, Y: mix(sy, ty, e) + math.Sin(e*math.Pi)*swirl,
				Side: side, Seen: seen, Part: part, Priority: c.Priority, Focus: c.Focus,
			})
		}
		if e := easeOut(phase(sc, at, start+640, 300)); e < 1 {
			sc.Cards[i].Seen, sc.Cards[i].Glare = e, 0.33*(1-e)
		}
	}
	sc.groundsFade(at, 60, 40, 500)
	sc.chrome(at, 80, 400)
	sc.beat(s, at, lead+760)
}

func craning(sc *Scene, s Stage, at float64) {
	cx, cy := s.lead()
	const close = 1.85
	tx, ty := float64(s.Wide)/2-close*cx, float64(s.High)*0.45-close*cy
	if p := phase(sc, at, 0, 1400); p < 1 {
		e := bezier(0.6, 0, 0.2, 1)(p)
		at3 := []float64{0, 0.7, 1}
		sc.Camera.PivotX, sc.Camera.PivotY = 0, 0
		sc.Camera.DX = keyed(e, at3, []float64{tx, tx * 0.12, 0})
		sc.Camera.DY = keyed(e, at3, []float64{ty, ty * 0.12, 0})
		sc.Camera.Grow = keyed(e, at3, []float64{close, 1.06, 1})
		sc.Camera.Turn = keyed(e, at3, []float64{-1.2, -0.15, 0})
	}
	far := s.distances()
	for i := range s.Cards {
		if i == s.Lead {
			continue
		}
		if e := easeOut(phase(sc, at, 250+far[i]*600, 420)); e < 1 {
			sc.Cards[i].Seen = e
		}
	}
	if g := phase(sc, at, 200, 700); g < 1 {
		for i := range sc.Grounds {
			sc.Grounds[i].Seen = mix(0.3, 1, g)
		}
	}
	sc.chromeFades(at, 800, 400)
	sc.beat(s, at, 60)
}

func drafting(sc *Scene, s Stage, at float64) {
	sc.chrome(at, 0, 300)
	for i := range sc.Grounds {
		p := phase(sc, at, float64(i)*40, 600)
		if p >= 1 {
			continue
		}
		sc.Grounds[i].Seen = clamp01(p / 0.2)
		sc.Grounds[i].Draft = &Draft{Fill: keyed(p, []float64{0, 0.55, 1}, []float64{0, 0, 1}), Stripe: 1, Facts: 1, Typed: -1}
	}
	lead := 0.0
	for i, c := range s.Cards {
		base := 80 + float64(i)*26
		if i == s.Lead {
			lead = base
		}
		p := phase(sc, at, base, 560)
		per := min(16, 520/float64(max(1, c.Letters)))
		typing := base + 340
		typed := 0
		if at >= typing {
			typed = min(c.Letters, int(math.Floor((at-typing)/per))+1)
		}
		facts := easeOut(phase(sc, at, typing+float64(c.Letters)*per, 260))
		stripe := easeOut(phase(sc, at, base+300, 280))
		if p >= 1 && typed >= c.Letters && facts >= 1 && stripe >= 1 {
			continue
		}
		sc.Cards[i].Seen = keyed(p, []float64{0, 0.15}, []float64{0, 1})
		sc.Cards[i].Draft = &Draft{Fill: keyed(cssEaseOut(p), []float64{0, 0.45, 1}, []float64{0, 0, 1}), Stripe: stripe, Facts: facts, Typed: typed}
	}
	sc.beat(s, at, lead+340+33*16)
}

func stoning(sc *Scene, s Stage, at float64) {
	sc.reach(1800)
	progress := clamp01(at / 1800)
	if progress >= 1 {
		return
	}
	for i := range s.Cards {
		_, fade := Arriving(progress, i, len(s.Cards))
		sc.Cards[i].Seen = fade
	}
	sc.Stones = true
}
