package effe

// Ranges are 1-indexed and inclusive.
type rangeSpec struct {
	cLo int
	cHi int
	rLo int
	rHi int
}

func makeColumnsRangeSpec(cLo int, cHi int) rangeSpec {
	return rangeSpec{
		cLo: cLo,
		cHi: cHi,
	}
}

func makeRowsRangeSpec(rLo int, rHi int) rangeSpec {
	return rangeSpec{
		rLo: rLo,
		rHi: rHi,
	}
}

func makeRangeSpec(cLo int, cHi int, rLo int, rHi int) rangeSpec {
	return rangeSpec{
		cLo: cLo,
		cHi: cHi,
		rLo: rLo,
		rHi: rHi,
	}
}

func makeCellRangeSpec(c int, r int) rangeSpec {
	return rangeSpec{
		cLo: c,
		cHi: c,
		rLo: r,
		rHi: r,
	}
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func intersectRange(aLo int, aHi int, bLo int, bHi int) (int, int) {
	if aLo == 0 {
		return bLo, bHi
	} else if bLo == 0 {
		return aLo, aHi
	} else {
		return max(aLo, bLo), min(aHi, bHi)
	}
}

func (r rangeSpec) intersect(other rangeSpec) rangeSpec {
	cLo, cHi := intersectRange(r.cLo, r.cHi, other.cLo, other.cHi)
	rLo, rHi := intersectRange(r.rLo, r.rHi, other.rLo, other.rHi)
	return makeRangeSpec(cLo, cHi, rLo, rHi)
}
