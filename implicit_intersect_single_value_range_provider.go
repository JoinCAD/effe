package effe

type implicitIntersector struct {
	rp     RangeProvider
	target Range
}

func (i implicitIntersector) ParseRange(text string) Range {
	return i.rp.ParseRange(text)
}

func (i implicitIntersector) Intersect(a Range, b Range) Range {
	return i.rp.Intersect(a, b)
}

func (i implicitIntersector) ImplicitIntersect(a Range, b Range) Range {
	return i.rp.ImplicitIntersect(a, b)
}

func (i implicitIntersector) Single(a Range) Value {
	if a.IsSingleValue() {
		return i.rp.Single(a)
	}
	return i.rp.Single(i.rp.ImplicitIntersect(target, a))
}

func (i implicitIntersector) Values(a Range) <-chan Value {
	return i.rp.Values(a)
}
