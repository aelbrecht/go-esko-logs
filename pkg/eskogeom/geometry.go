package eskogeom

type Rectangle struct {
	Origin        Point
	Width, Height float64
}

func (r Rectangle) Contains(p Point) bool {
	return p.X >= r.Origin.X && p.X <= r.Origin.X+r.Width &&
		p.Y >= r.Origin.Y && p.Y <= r.Origin.Y+r.Height
}

func (r Rectangle) Center() Point {
	return Point{
		X: r.Origin.X + r.Width/2,
		Y: r.Origin.Y + r.Height/2,
	}
}

func (r Rectangle) Apply(o Transformation) Rectangle {
	return Rectangle{
		Origin: r.Origin,
		Width:  r.Width * o.Scale.X,
		Height: r.Height * o.Scale.Y,
	}
}

func Scale(h float64, v float64) Transformation {
	return Transformation{Scale: Vector{h, v}, Move: Origin()}
}

func (r Rectangle) ApplyCenter(t Transformation) Rectangle {
	// Translate center to origin, scale, then translate back
	center := r.Center()
	translatedOrigin := r.Origin.Sub(center).Apply(t).Add(center)
	scaledSize := Point{X: r.Width, Y: r.Height}.Mul(t.Scale)

	return Rectangle{
		Origin: translatedOrigin,
		Width:  scaledSize.X,
		Height: scaledSize.Y,
	}
}

func (r Rectangle) ToTransformation() Transformation {
	return Transformation{
		Scale: Vector{r.Width, r.Height},
		Move:  Point{r.Origin.X, r.Origin.Y},
	}
}

func Unit() Vector {
	return Vector{1, 1}
}

func (v Vector) Mul(scalar float64) Vector {
	return Vector{
		X: scalar * v.X,
		Y: scalar * v.Y,
	}
}

type Vector struct {
	X float64
	Y float64
}

type Transformation struct {
	Scale Vector
	Move  Point
}

func (t Transformation) Compose(o Transformation) Transformation {
	newScale := Vector{
		X: t.Scale.X * o.Scale.X,
		Y: t.Scale.Y * o.Scale.Y,
	}
	newMove := Point{
		X: t.Move.X*o.Scale.X + o.Move.X,
		Y: t.Move.Y*o.Scale.Y + o.Move.Y,
	}
	return Transformation{
		Scale: newScale,
		Move:  newMove,
	}
}

func (t Transformation) Invert() Transformation {
	invScale := Vector{
		X: 1 / t.Scale.X,
		Y: 1 / t.Scale.Y,
	}
	invMove := Point{
		X: -t.Move.X * invScale.X,
		Y: -t.Move.Y * invScale.Y,
	}
	return Transformation{
		Scale: invScale,
		Move:  invMove,
	}
}

func (v Vector) Invert() Vector {
	return Vector{
		X: 1 / v.X,
		Y: 1 / v.Y,
	}
}

type Point struct {
	X, Y float64
}

func Origin() Point {
	return Point{0, 0}
}

func (p Point) Invert() Point {
	return Point{
		X: -p.X,
		Y: -p.Y,
	}
}

func (p Point) Apply(t Transformation) Point {
	return p.Mul(t.Scale).Add(t.Move)
}

func (p Point) Sub(o Point) Point {
	return Point{p.X - o.X, p.Y - o.Y}
}

func (p Point) Add(o Point) Point {
	return Point{p.X + o.X, p.Y + o.Y}
}

func (p Point) Mul(o Vector) Point {
	return Point{p.X * o.X, p.Y * o.Y}
}

func (p Point) Div(o Vector) Point {
	return Point{p.X / o.X, p.Y / o.Y}
}

func (p Point) ToFloat32() (float32, float32) {
	return float32(p.X), float32(p.Y)
}

type Quad struct {
	C, P Point
}

type Cubic struct {
	C1, C2, P Point
}

type Path struct {
	MoveTo Point
	Points []interface{}
	Closed bool
}

type Compound struct {
	Index    int
	SubPaths []Path
	MetaData map[string]string
}
