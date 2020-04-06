package bls12381

import (
	"fmt"
	"math"
	"math/big"
)

// PointG2 is type for point in G2.
// PointG2 is both used for Affine and Jacobian point representation.
// If z is equal to one the point is accounted as in affine form.
type PointG2 [3]fe2

// Set copies valeus of one point to another.
func (p *PointG2) Set(p2 *PointG2) *PointG2 {
	p[0][0].Set(&p2[0][0])
	p[1][1].Set(&p2[1][1])
	p[2][0].Set(&p2[2][0])
	p[0][1].Set(&p2[0][1])
	p[1][0].Set(&p2[1][0])
	p[2][1].Set(&p2[2][1])
	return p
}

type tempG2 struct {
	t [9]*fe2
}

// G2 is struct for G2 group.
type G2 struct {
	f *fp2
	tempG2
}

// NewG2 constructs a new G2 instance.
func NewG2(f *fp2) *G2 {
	cfgArch()
	if f == nil {
		f = newFp2()
	}
	t := newTempG2()
	return &G2{f, t}
}

func newTempG2() tempG2 {
	t := [9]*fe2{}
	for i := 0; i < 9; i++ {
		t[i] = &fe2{}
	}
	return tempG2{t}
}

// Q returns group order in big.Int.
func (g *G2) Q() *big.Int {
	return new(big.Int).Set(q)
}

// FromUncompressed expects byte slice larger than 192 bytes and given bytes returns a new point in G2.
// Serialization rules are in line with zcash library. See below for details.
// https://github.com/zcash/librustzcash/blob/master/pairing/src/bls12_381/README.md#serialization
func (g *G2) FromUncompressed(uncompressed []byte) (*PointG2, error) {
	if len(uncompressed) < 192 {
		return nil, fmt.Errorf("input string should be equal or larger than 192")
	}
	var in [192]byte
	copy(in[:], uncompressed[:192])
	if in[0]&(1<<7) != 0 {
		return nil, fmt.Errorf("compression flag should be zero")
	}
	if in[0]&(1<<5) != 0 {
		return nil, fmt.Errorf("sort flag should be zero")
	}
	if in[0]&(1<<6) != 0 {
		for i, v := range in {
			if (i == 0 && v != 0x40) || (i != 0 && v != 0x00) {
				return nil, fmt.Errorf("input string should be zero when infinity flag is set")
			}
		}
		return g.Zero(), nil
	}
	in[0] &= 0x1f
	x, err := g.f.fromBytes(in[:96])
	if err != nil {
		return nil, err
	}
	y, err := g.f.fromBytes(in[96:])
	if err != nil {
		return nil, err
	}
	z := g.f.one()
	p := &PointG2{*x, *y, *z}
	if !g.IsOnCurve(p) {
		return nil, fmt.Errorf("point is not on curve")
	}
	if !g.InCorrectSubgroup(p) {
		return nil, fmt.Errorf("point is not on correct subgroup")
	}
	return p, nil
}

// ToUncompressed given a G2 point returns bytes in uncompressed (x, y) form of the point.
func (g *G2) ToUncompressed(p *PointG2) []byte {
	out := make([]byte, 192)
	g.Affine(p)
	if g.IsZero(p) {
		out[0] |= 1 << 6
	}
	copy(out[:96], g.f.toBytes(&p[0]))
	copy(out[96:], g.f.toBytes(&p[1]))
	return out
}

// FromCompressed expects byte slice larger than 96 bytes and given bytes returns a new point in G2.
// Serialization rules are in line with zcash library. See below for details.
// https://github.com/zcash/librustzcash/blob/master/pairing/src/bls12_381/README.md#serialization
func (g *G2) FromCompressed(compressed []byte) (*PointG2, error) {
	if len(compressed) < 96 {
		return nil, fmt.Errorf("input string should be equal or larger than 96")
	}
	var in [96]byte
	copy(in[:], compressed[:])
	if in[0]&(1<<7) == 0 {
		return nil, fmt.Errorf("bad compression")
	}
	if in[0]&(1<<6) != 0 {
		// in[0] == (1 << 6) + (1 << 7)
		for i, v := range in {
			if (i == 0 && v != 0xc0) || (i != 0 && v != 0x00) {
				return nil, fmt.Errorf("input string should be zero when infinity flag is set")
			}
		}
		return g.Zero(), nil
	}
	a := in[0]&(1<<5) != 0
	in[0] &= 0x1f
	x, err := g.f.fromBytes(in[:])
	if err != nil {
		return nil, err
	}
	// solve curve equation
	y := &fe2{}
	g.f.square(y, x)
	g.f.mul(y, y, x)
	g.f.add(y, y, b2)
	if ok := g.f.sqrt(y, y); !ok {
		return nil, fmt.Errorf("point is not on curve")
	}
	// select lexicographically, should be in normalized form
	negYn, negY, yn := &fe2{}, &fe2{}, &fe2{}
	g.f.fromMont(yn, y)
	g.f.neg(negY, y)
	g.f.neg(negYn, yn)
	if (yn[1].Cmp(&negYn[1]) > 0 != a) || (yn[1].IsZero() && yn[0].Cmp(&negYn[0]) > 0 != a) {
		g.f.copy(y, negY)
	}
	z := g.f.one()
	p := &PointG2{*x, *y, *z}
	if !g.InCorrectSubgroup(p) {
		return nil, fmt.Errorf("point is not on correct subgroup")
	}
	return p, nil
}

// ToCompressed given a G2 point returns bytes in compressed form of the point.
// Serialization rules are in line with zcash library. See below for details.
// https://github.com/zcash/librustzcash/blob/master/pairing/src/bls12_381/README.md#serialization
func (g *G2) ToCompressed(p *PointG2) []byte {
	out := make([]byte, 96)
	g.Affine(p)
	if g.IsZero(p) {
		out[0] |= 1 << 6
	} else {
		copy(out[:], g.f.toBytes(&p[0]))
		y, negY := &fe2{}, &fe2{}
		g.f.fromMont(y, &p[1])
		g.f.neg(negY, y)
		if (y[1].Cmp(&negY[1]) > 0) || (y[1].IsZero() && y[1].Cmp(&negY[1]) > 0) {
			out[0] |= 1 << 5
		}
	}
	out[0] |= 1 << 7
	return out
}

// FromBytes constructs a new point given byte input.
// Byte input expected to be larger than 96 bytes.
// First 96 bytes should be concatenation of x and y values
func (g *G2) FromBytes(in []byte) (*PointG2, error) {
	if len(in) < 192 {
		return nil, fmt.Errorf("input string should be equal or larger than 192")
	}
	p0, err := g.f.fromBytes(in[:96])
	if err != nil {
		return nil, err
	}
	p1, err := g.f.fromBytes(in[96:])
	if err != nil {
		panic(err)
	}
	p2 := g.f.one()
	p := &PointG2{*p0, *p1, *p2}
	if !g.IsOnCurve(p) {
		return nil, fmt.Errorf("point is not on curve")
	}
	return p, nil
}

func (g *G2) fromBytesUnchecked(in []byte) (*PointG2, error) {
	p0, err := g.f.fromBytes(in[:96])
	if err != nil {
		return nil, err
	}
	p1, err := g.f.fromBytes(in[96:])
	if err != nil {
		panic(err)
	}
	p2 := g.f.one()
	return &PointG2{*p0, *p1, *p2}, nil
}

// New creates a new G2 Point which is equal to zero in other words point at infinity.
func (g *G2) New() *PointG2 {
	return g.Zero()
}

// Zero returns a new G2 Point which is equal to point at infinity.
func (g *G2) Zero() *PointG2 {
	return &PointG2{
		*g.f.zero(),
		*g.f.one(),
		*g.f.zero(),
	}
}

// One returns a new G2 Point which is equal to generator point.
func (g *G2) One() *PointG2 {
	return g.Copy(&PointG2{}, &g2One)
}

// Copy copies source point to destination point.
func (g *G2) Copy(dst *PointG2, src *PointG2) *PointG2 {
	return dst.Set(src)
}

// IsZero returns true if given point is equal to zero.
func (g *G2) IsZero(p *PointG2) bool {
	return g.f.isZero(&p[2])
}

// Equal checks if given two G2 point is equal in their affine form.
func (g *G2) Equal(p1, p2 *PointG2) bool {
	if g.IsZero(p1) {
		return g.IsZero(p2)
	}
	if g.IsZero(p2) {
		return g.IsZero(p1)
	}
	t := g.t
	g.f.square(t[0], &p1[2])
	g.f.square(t[1], &p2[2])
	g.f.mul(t[2], t[0], &p2[0])
	g.f.mul(t[3], t[1], &p1[0])
	g.f.mul(t[0], t[0], &p1[2])
	g.f.mul(t[1], t[1], &p2[2])
	g.f.mul(t[1], t[1], &p1[1])
	g.f.mul(t[0], t[0], &p2[1])
	return g.f.equal(t[0], t[1]) && g.f.equal(t[2], t[3])
}

// InCorrectSubgroup checks whether given point is in correct subgroup.
func (g *G2) InCorrectSubgroup(p *PointG2) bool {
	tmp := &PointG2{}
	g.MulScalar(tmp, p, q)
	return g.IsZero(tmp)
}

// IsOnCurve checks a G2 point is on curve.
func (g *G2) IsOnCurve(p *PointG2) bool {
	if g.IsZero(p) {
		return true
	}
	t := g.t
	g.f.square(t[0], &p[1])
	g.f.square(t[1], &p[0])
	g.f.mul(t[1], t[1], &p[0])
	g.f.square(t[2], &p[2])
	g.f.square(t[3], t[2])
	g.f.mul(t[2], t[2], t[3])
	g.f.mul(t[2], b2, t[2])
	g.f.add(t[1], t[1], t[2])
	return g.f.equal(t[0], t[1])
}

// IsAffine checks a G2 point whether it is in affine form.
func (g *G2) IsAffine(p *PointG2) bool {
	return g.f.equal(&p[2], g.f.one())
}

// Affine calculates affine form of given G2 point.
func (g *G2) Affine(p *PointG2) *PointG2 {
	if g.IsZero(p) {
		return p
	}
	if !g.IsAffine(p) {
		t := g.t
		g.f.inverse(t[0], &p[2])
		g.f.square(t[1], t[0])
		g.f.mul(&p[0], &p[0], t[1])
		g.f.mul(t[0], t[0], t[1])
		g.f.mul(&p[1], &p[1], t[0])
		g.f.copy(&p[2], g.f.one())
	}
	return p
}

// Add adds two G2 points p1, p2 and assigns the result to point at first argument.
func (g *G2) Add(r, p1, p2 *PointG2) *PointG2 {
	// http://www.hyperelliptic.org/EFD/gp/auto-shortw-jacobian-0.html#addition-add-2007-bl
	if g.IsZero(p1) {
		g.Copy(r, p2)
		return r
	}
	if g.IsZero(p2) {
		g.Copy(r, p1)
		return r
	}
	t := g.t
	g.f.square(t[7], &p1[2])
	g.f.mul(t[1], &p2[0], t[7])
	g.f.mul(t[2], &p1[2], t[7])
	g.f.mul(t[0], &p2[1], t[2])
	g.f.square(t[8], &p2[2])
	g.f.mul(t[3], &p1[0], t[8])
	g.f.mul(t[4], &p2[2], t[8])
	g.f.mul(t[2], &p1[1], t[4])
	if g.f.equal(t[1], t[3]) {
		if g.f.equal(t[0], t[2]) {
			return g.Double(r, p1)
		} else {
			return g.Copy(r, infinity2)
		}
	}
	g.f.sub(t[1], t[1], t[3])
	g.f.double(t[4], t[1])
	g.f.square(t[4], t[4])
	g.f.mul(t[5], t[1], t[4])
	g.f.sub(t[0], t[0], t[2])
	g.f.double(t[0], t[0])
	g.f.square(t[6], t[0])
	g.f.sub(t[6], t[6], t[5])
	g.f.mul(t[3], t[3], t[4])
	g.f.double(t[4], t[3])
	g.f.sub(&r[0], t[6], t[4])
	g.f.sub(t[4], t[3], &r[0])
	g.f.mul(t[6], t[2], t[5])
	g.f.double(t[6], t[6])
	g.f.mul(t[0], t[0], t[4])
	g.f.sub(&r[1], t[0], t[6])
	g.f.add(t[0], &p1[2], &p2[2])
	g.f.square(t[0], t[0])
	g.f.sub(t[0], t[0], t[7])
	g.f.sub(t[0], t[0], t[8])
	g.f.mul(&r[2], t[0], t[1])
	return r
}

// Double doubles a G2 point p and assigns the result to the point at first argument.
func (g *G2) Double(r, p *PointG2) *PointG2 {
	// http://www.hyperelliptic.org/EFD/gp/auto-shortw-jacobian-0.html#doubling-dbl-2009-l
	if g.IsZero(p) {
		g.Copy(r, p)
		return r
	}
	t := g.t
	g.f.square(t[0], &p[0])
	g.f.square(t[1], &p[1])
	g.f.square(t[2], t[1])
	g.f.add(t[1], &p[0], t[1])
	g.f.square(t[1], t[1])
	g.f.sub(t[1], t[1], t[0])
	g.f.sub(t[1], t[1], t[2])
	g.f.double(t[1], t[1])
	g.f.double(t[3], t[0])
	g.f.add(t[0], t[3], t[0])
	g.f.square(t[4], t[0])
	g.f.double(t[3], t[1])
	g.f.sub(&r[0], t[4], t[3])
	g.f.sub(t[1], t[1], &r[0])
	g.f.double(t[2], t[2])
	g.f.double(t[2], t[2])
	g.f.double(t[2], t[2])
	g.f.mul(t[0], t[0], t[1])
	g.f.sub(t[1], t[0], t[2])
	g.f.mul(t[0], &p[1], &p[2])
	g.f.copy(&r[1], t[1])
	g.f.double(&r[2], t[0])
	return r
}

// Neg negates a G2 point p and assigns the result to the point at first argument.
func (g *G2) Neg(r, p *PointG2) *PointG2 {
	g.f.copy(&r[0], &p[0])
	g.f.neg(&r[1], &p[1])
	g.f.copy(&r[2], &p[2])
	return r
}

// Sub subtracts two G2 points p1, p2 and assigns the result to point at first argument.
func (g *G2) Sub(c, a, b *PointG2) *PointG2 {
	d := &PointG2{}
	g.Neg(d, b)
	g.Add(c, a, d)
	return c
}

// MulScalar multiplies a point by given scalar value in big.Int and assigns the result to point at first argument.
func (g *G2) MulScalar(c, p *PointG2, e *big.Int) *PointG2 {
	q, n := &PointG2{}, &PointG2{}
	g.Copy(n, p)
	l := e.BitLen()
	for i := 0; i < l; i++ {
		if e.Bit(i) == 1 {
			g.Add(q, q, n)
		}
		g.Double(n, n)
	}
	return g.Copy(c, q)
}

// ClearCofactor maps given a G2 point to correct subgroup
func (g *G2) ClearCofactor(p *PointG2) {
	g.MulScalar(p, p, cofactorEFFG2)
}

// MultiExp calculates multi exponentiation. Given pairs of G2 point and scalar values
// (P_0, e_0), (P_1, e_1), ... (P_n, e_n) calculates r = e_0 * P_0 + e_1 * P_1 + ... + e_n * P_n
// Length of points and scalars are expected to be equal, otherwise an error is returned.
// Result is assigned to point at first argument.
func (g *G2) MultiExp(r *PointG2, points []*PointG2, powers []*big.Int) (*PointG2, error) {
	if len(points) != len(powers) {
		return nil, fmt.Errorf("point and scalar vectors should be in same length")
	}
	var c uint32 = 3
	if len(powers) >= 32 {
		c = uint32(math.Ceil(math.Log10(float64(len(powers)))))
	}
	bucketSize, numBits := (1<<c)-1, uint32(g.Q().BitLen())
	windows := make([]*PointG2, numBits/c+1)
	bucket := make([]*PointG2, bucketSize)
	acc, sum := g.New(), g.New()
	for i := 0; i < bucketSize; i++ {
		bucket[i] = g.New()
	}
	mask := (uint64(1) << c) - 1
	j := 0
	var cur uint32
	for cur <= numBits {
		g.Copy(acc, g.Zero())
		bucket = make([]*PointG2, (1<<c)-1)
		for i := 0; i < len(bucket); i++ {
			bucket[i] = g.New()
		}
		for i := 0; i < len(powers); i++ {
			s0 := powers[i].Uint64()
			index := uint(s0 & mask)
			if index != 0 {
				g.Add(bucket[index-1], bucket[index-1], points[i])
			}
			powers[i] = new(big.Int).Rsh(powers[i], uint(c))
		}

		g.Copy(sum, g.Zero())
		for i := len(bucket) - 1; i >= 0; i-- {
			g.Add(sum, sum, bucket[i])
			g.Add(acc, acc, sum)
		}
		windows[j] = g.New()
		g.Copy(windows[j], acc)
		j++
		cur += c
	}
	g.Copy(acc, g.Zero())
	for i := len(windows) - 1; i >= 0; i-- {
		for j := uint32(0); j < c; j++ {
			g.Double(acc, acc)
		}
		g.Add(acc, acc, windows[i])
	}
	g.Copy(r, acc)
	return r, nil
}

func (g *G2) wnafMul(c, p *PointG2, e []uint64) *PointG2 {
	windowSize := uint(6)
	precompTable := make([]*PointG2, (1 << (windowSize - 1)))
	for i := 0; i < len(precompTable); i++ {
		precompTable[i] = g.New()
	}
	var indexForPositive uint64
	indexForPositive = (1 << (windowSize - 2))
	g.Copy(precompTable[indexForPositive], p)
	g.Neg(precompTable[indexForPositive-1], p)
	doubled, precomp := g.New(), g.New()
	g.Double(doubled, p)
	g.Copy(precomp, p)
	for i := uint64(1); i < indexForPositive; i++ {
		g.Add(precomp, precomp, doubled)
		g.Copy(precompTable[indexForPositive+i], precomp)
		g.Neg(precompTable[indexForPositive-1-i], precomp)
	}

	wnaf := wnaf(e, windowSize)
	q := g.Zero()
	l := len(wnaf)
	found := false
	var idx uint64
	for i := l - 1; i >= 0; i-- {
		if found {
			g.Double(q, q)
		}
		if wnaf[i] != 0 {
			found = true
			if wnaf[i] > 0 {
				idx = uint64(wnaf[i] >> 1)
				g.Add(q, q, precompTable[indexForPositive+idx])
			} else {
				idx = uint64(((0 - wnaf[i]) >> 1))
				g.Add(q, q, precompTable[indexForPositive-1-idx])
			}
		}
	}
	g.Copy(c, q)
	return c
}

// MapToPointTI given a byte slice returns a valid G2 point.
// This mapping function implements the 'try and increment' method.
func (g *G2) MapToPointTI(in []byte) (*PointG2, error) {
	fp2 := g.f
	x, err := fp2.fromBytes(in)
	if err != nil {
		return nil, err
	}
	y := &fe2{}
	one := fp2.one()
	for {
		fp2.square(y, x)
		fp2.mul(y, y, x)
		fp2.add(y, y, b2)
		if ok := fp2.sqrt(y, y); ok {
			// favour negative y
			negYn, negY, yn := &fe2{}, &fe2{}, &fe2{}
			fp2.fromMont(yn, y)
			fp2.neg(negY, y)
			fp2.neg(negYn, yn)
			if yn[1].Cmp(&negYn[1]) > 0 || (yn[1].IsZero() && yn[0].Cmp(&negYn[0]) > 0) {
				// fp2.copy(y, y)
			} else {
				fp2.copy(y, negY)
			}
			p := &PointG2{*x, *y, *one}
			g.ClearCofactor(p)
			return p, nil
		}
		fp2.add(x, x, one)
	}
}

// MapToPointSWU given a byte slice returns a valid G2 point.
// This mapping function implements the Simplified Shallue-van de Woestijne-Ulas method.
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-05#section-6.6.2
// Input byte slice should be a valid field element, otherwise an error is returned.
// Clearing cofactor h is done with wnaf multiplication with windows size 6.
func (g *G2) MapToPointSWU(in []byte) (*PointG2, error) {
	fp2 := g.f
	u, err := fp2.fromBytes(in)
	if err != nil {
		return nil, err
	}
	x, y := fp2.swuMap(u)
	fp2.isogenyMap(x, y)
	one := fp2.one()
	q := &PointG2{*x, *y, *one}
	if !g.IsOnCurve(q) {
		return nil, fmt.Errorf("Found point is not on curve")
	}
	cofactor := []uint64{
		0xe8020005aaa95551,
		0x59894c0adebbf6b4,
		0xe954cbc06689f6a3,
		0x2ec0ec69d7477c1a,
		0x6d82bf015d1212b0,
		0x329c2f178731db95,
		0x9986ff031508ffe1,
		0x88e2a8e9145ad768,
		0x584c6a0ea91b3528,
		0x0bc69f08f2ee75b3,
	}
	g.wnafMul(q, q, cofactor)
	return g.Affine(q), nil
}
