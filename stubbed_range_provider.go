package effe

type stubbedRange struct {
	clow, chi int
	rlow, rhi int
}

func (sr stubbedRange) IsSingle() bool {
	return clow == chi && rlow == rhi
}

type stubbedRangeProvider struct{}

func (srp stubbedRangeProvider) ParseRange(text string) Range {
	s := strings.Split(text, ":")
	
	first := s[0]
}

func (srp stubbedRangeProvider) Intersect(a Range, b Range) Range {

} 
ImplicitIntersect(a Range, b Range) Range
Single(a Range) Value
Values(a Range) <-chan Value