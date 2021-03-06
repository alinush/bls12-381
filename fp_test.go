package bls12381

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"testing"
)

func TestFpSerialization(t *testing.T) {
	zero := zero()
	t.Run("zero", func(t *testing.T) {
		in := make([]byte, 48)
		fe, err := fromBytes(in)
		if err != nil {
			t.Fatal(err)
		}
		if !equal(fe, zero) {
			t.Fatalf("bad serialization\n")
		}
		if !bytes.Equal(in, toBytes(fe)) {
			t.Fatalf("bad serialization\n")
		}
	})
	t.Run("bytes", func(t *testing.T) {
		for i := 0; i < fuz; i++ {
			a, _ := newRand(rand.Reader)
			b, err := fromBytes(toBytes(a))
			if err != nil {
				t.Fatal(err)
			}
			if !equal(a, b) {
				t.Fatalf("bad serialization\n")
			}
		}
	})
	t.Run("string", func(t *testing.T) {
		for i := 0; i < fuz; i++ {
			a, _ := newRand(rand.Reader)
			b, err := fromString(toString(a))
			if err != nil {
				t.Fatal(err)
			}
			if !equal(a, b) {
				t.Fatalf("bad encoding or decoding\n")
			}
		}
	})
	t.Run("big", func(t *testing.T) {
		for i := 0; i < fuz; i++ {
			a, _ := newRand(rand.Reader)
			b, err := fromBig(toBig(a))
			if err != nil {
				t.Fatal(err)
			}
			if !equal(a, b) {
				t.Fatalf("bad encoding or decoding\n")
			}
		}
	})
}

func TestFpAdditionCrossAgainstBigInt(t *testing.T) {
	for i := 0; i < fuz; i++ {
		a, _ := newRand(rand.Reader)
		b, _ := newRand(rand.Reader)
		c := new(fe)
		big_a := toBig(a)
		big_b := toBig(b)
		big_c := new(big.Int)
		add(c, a, b)
		out_1 := toBytes(c)
		out_2 := padBytes(big_c.Add(big_a, big_b).Mod(big_c, modulus.big()).Bytes(), 48)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied A")
		}
		double(c, a)
		out_1 = toBytes(c)
		out_2 = padBytes(big_c.Add(big_a, big_a).Mod(big_c, modulus.big()).Bytes(), 48)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied B")
		}
		sub(c, a, b)
		out_1 = toBytes(c)
		out_2 = padBytes(big_c.Sub(big_a, big_b).Mod(big_c, modulus.big()).Bytes(), 48)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied C")
		}
		neg(c, a)
		out_1 = toBytes(c)
		out_2 = padBytes(big_c.Neg(big_a).Mod(big_c, modulus.big()).Bytes(), 48)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied D")
		}
	}
}

func TestFpAdditionProperties(t *testing.T) {
	for i := 0; i < fuz; i++ {

		zero := zero()
		a, _ := newRand(rand.Reader)
		b, _ := newRand(rand.Reader)
		c_1, c_2 := new(fe), new(fe)
		add(c_1, a, zero)
		if !equal(c_1, a) {
			t.Fatalf("a + 0 == a")
		}
		sub(c_1, a, zero)
		if !equal(c_1, a) {
			t.Fatalf("a - 0 == a")
		}
		double(c_1, zero)
		if !equal(c_1, zero) {
			t.Fatalf("2 * 0 == 0")
		}
		neg(c_1, zero)
		if !equal(c_1, zero) {
			t.Fatalf("-0 == 0")
		}
		sub(c_1, zero, a)
		neg(c_2, a)
		if !equal(c_1, c_2) {
			t.Fatalf("0-a == -a")
		}
		double(c_1, a)
		add(c_2, a, a)
		if !equal(c_1, c_2) {
			t.Fatalf("2 * a == a + a")
		}
		add(c_1, a, b)
		add(c_2, b, a)
		if !equal(c_1, c_2) {
			t.Fatalf("a + b = b + a")
		}
		sub(c_1, a, b)
		sub(c_2, b, a)
		neg(c_2, c_2)
		if !equal(c_1, c_2) {
			t.Fatalf("a - b = - ( b - a )")
		}
		c_x, _ := newRand(rand.Reader)
		add(c_1, a, b)
		add(c_1, c_1, c_x)
		add(c_2, a, c_x)
		add(c_2, c_2, b)
		if !equal(c_1, c_2) {
			t.Fatalf("(a + b) + c == (a + c ) + b")
		}
		sub(c_1, a, b)
		sub(c_1, c_1, c_x)
		sub(c_2, a, c_x)
		sub(c_2, c_2, b)
		if !equal(c_1, c_2) {
			t.Fatalf("(a - b) - c == (a - c ) -b")
		}
	}
}

func TestFpMultiplicationCrossAgainstBigInt(t *testing.T) {
	for i := 0; i < fuz; i++ {
		a, _ := newRand(rand.Reader)
		b, _ := newRand(rand.Reader)
		c := new(fe)
		big_a := toBig(a)
		big_b := toBig(b)
		big_c := new(big.Int)
		mul(c, a, b)
		out_1 := toBytes(c)
		out_2 := padBytes(big_c.Mul(big_a, big_b).Mod(big_c, modulus.big()).Bytes(), 48)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied")
		}
	}
}

func TestFpMultiplicationProperties(t *testing.T) {
	for i := 0; i < fuz; i++ {
		a, _ := newRand(rand.Reader)
		b, _ := newRand(rand.Reader)
		zero := zero()
		one := one()
		c_1, c_2 := new(fe), new(fe)
		mul(c_1, a, zero)
		if !equal(c_1, zero) {
			t.Fatalf("a * 0 == 0")
		}
		mul(c_1, a, one)
		if !equal(c_1, a) {
			t.Fatalf("a * 1 == a")
		}
		mul(c_1, a, b)
		mul(c_2, b, a)
		if !equal(c_1, c_2) {
			t.Fatalf("a * b == b * a")
		}
		c_x, _ := newRand(rand.Reader)
		mul(c_1, a, b)
		mul(c_1, c_1, c_x)
		mul(c_2, c_x, b)
		mul(c_2, c_2, a)
		if !equal(c_1, c_2) {
			t.Fatalf("(a * b) * c == (a * c) * b")
		}
	}
}

func TestFpExponentiation(t *testing.T) {
	for i := 0; i < fuz; i++ {
		a, _ := newRand(rand.Reader)
		u := new(fe)
		exp(u, a, big.NewInt(0))
		if !equal(u, one()) {
			t.Fatalf("a^0 == 1")
		}
		exp(u, a, big.NewInt(1))
		if !equal(u, a) {
			t.Fatalf("a^1 == a")
		}
		v := new(fe)
		mul(u, a, a)
		mul(u, u, u)
		mul(u, u, u)
		exp(v, a, big.NewInt(8))
		if !equal(u, v) {
			t.Fatalf("((a^2)^2)^2 == a^8")
		}
		p := modulus.big()
		exp(u, a, p)
		if !equal(u, a) {
			t.Fatalf("a^p == a")
		}
		exp(u, a, p.Sub(p, big.NewInt(1)))
		if !equal(u, one()) {
			t.Fatalf("a^(p-1) == 1")
		}
	}
}

func TestFpInversion(t *testing.T) {
	for i := 0; i < fuz; i++ {
		u := new(fe)
		zero := zero()
		one := one()
		inverse(u, zero)
		if !equal(u, zero) {
			t.Fatalf("(0^-1) == 0)")
		}
		inverse(u, one)
		if !equal(u, one) {
			t.Fatalf("(1^-1) == 1)")
		}
		a, _ := newRand(rand.Reader)
		inverse(u, a)
		mul(u, u, a)
		if !equal(u, one) {
			t.Fatalf("(r*a) * r*(a^-1) == r)")
		}
		v := new(fe)
		p := modulus.big()
		exp(u, a, p.Sub(p, big.NewInt(2)))
		inverse(v, a)
		if !equal(v, u) {
			t.Fatalf("a^(p-2) == a^-1")
		}
	}
}

func TestFpSquareRoot(t *testing.T) {
	r := new(fe)
	if sqrt(r, nonResidue1) {
		t.Fatalf("non residue cannot have a sqrt")
	}
	for i := 0; i < fuz; i++ {
		a, _ := newRand(rand.Reader)
		aa, rr, r := &fe{}, &fe{}, &fe{}
		square(aa, a)
		if !sqrt(r, aa) {
			t.Fatalf("bad sqrt 1")
		}
		square(rr, r)
		if !equal(rr, aa) {
			t.Fatalf("bad sqrt 2")
		}
	}
}
func TestFp2Serialization(t *testing.T) {
	field := newFp2()
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		b, err := field.fromBytes(field.toBytes(a))
		if err != nil {
			t.Fatal(err)
		}
		if !field.equal(a, b) {
			t.Fatalf("bad serialization\n")
		}
	}
}

func TestFp2AdditionProperties(t *testing.T) {
	field := newFp2()
	for i := 0; i < fuz; i++ {
		zero := field.zero()
		a, _ := field.rand(rand.Reader)
		b, _ := field.rand(rand.Reader)
		c_1 := field.new()
		c_2 := field.new()
		field.add(c_1, a, zero)
		if !field.equal(c_1, a) {
			t.Fatalf("a + 0 == a")
		}
		field.sub(c_1, a, zero)
		if !field.equal(c_1, a) {
			t.Fatalf("a - 0 == a")
		}
		field.double(c_1, zero)
		if !field.equal(c_1, zero) {
			t.Fatalf("2 * 0 == 0")
		}
		field.neg(c_1, zero)
		if !field.equal(c_1, zero) {
			t.Fatalf("-0 == 0")
		}
		field.sub(c_1, zero, a)
		field.neg(c_2, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("0-a == -a")
		}
		field.double(c_1, a)
		field.add(c_2, a, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("2 * a == a + a")
		}
		field.add(c_1, a, b)
		field.add(c_2, b, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("a + b = b + a")
		}
		field.sub(c_1, a, b)
		field.sub(c_2, b, a)
		field.neg(c_2, c_2)
		if !field.equal(c_1, c_2) {
			t.Fatalf("a - b = - ( b - a )")
		}
		c_x, _ := field.rand(rand.Reader)
		field.add(c_1, a, b)
		field.add(c_1, c_1, c_x)
		field.add(c_2, a, c_x)
		field.add(c_2, c_2, b)
		if !field.equal(c_1, c_2) {
			t.Fatalf("(a + b) + c == (a + c ) + b")
		}
		field.sub(c_1, a, b)
		field.sub(c_1, c_1, c_x)
		field.sub(c_2, a, c_x)
		field.sub(c_2, c_2, b)
		if !field.equal(c_1, c_2) {
			t.Fatalf("(a - b) - c == (a - c ) -b")
		}
	}
}

func TestFp2MultiplicationProperties(t *testing.T) {
	field := newFp2()
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		b, _ := field.rand(rand.Reader)
		zero := field.zero()
		one := field.one()
		c_1, c_2 := field.new(), field.new()
		field.mul(c_1, a, zero)
		if !field.equal(c_1, zero) {
			t.Fatalf("a * 0 == 0")
		}
		field.mul(c_1, a, one)
		if !field.equal(c_1, a) {
			t.Fatalf("a * 1 == a")
		}
		field.mul(c_1, a, b)
		field.mul(c_2, b, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("a * b == b * a")
		}
		c_x, _ := field.rand(rand.Reader)
		field.mul(c_1, a, b)
		field.mul(c_1, c_1, c_x)
		field.mul(c_2, c_x, b)
		field.mul(c_2, c_2, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("(a * b) * c == (a * c) * b")
		}
	}
}

func TestFp2Exponentiation(t *testing.T) {
	field := newFp2()
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		u := field.new()
		field.exp(u, a, big.NewInt(0))
		if !field.equal(u, field.one()) {
			t.Fatalf("a^0 == 1")
		}
		field.exp(u, a, big.NewInt(1))
		if !field.equal(u, a) {
			t.Fatalf("a^1 == a")
		}
		v := field.new()
		field.mul(u, a, a)
		field.mul(u, u, u)
		field.mul(u, u, u)
		field.exp(v, a, big.NewInt(8))
		if !field.equal(u, v) {
			t.Fatalf("((a^2)^2)^2 == a^8")
		}
		// p := modulus.big()
		// field.exp(u, a, p)
		// if !field.equal(u, a) {
		// 	t.Fatalf("a^p == a")
		// }
		// field.exp(u, a, p.Sub(p, big.NewInt(1)))
		// if !field.equal(u, field.one()) {
		// 	t.Fatalf("a^(p-1) == 1")
		// }
	}
}

func TestFp2Inversion(t *testing.T) {
	field := newFp2()
	u := field.new()
	zero := field.zero()
	one := field.one()
	field.inverse(u, zero)
	if !field.equal(u, zero) {
		t.Fatalf("(0 ^ -1) == 0)")
	}
	field.inverse(u, one)
	if !field.equal(u, one) {
		t.Fatalf("(1 ^ -1) == 1)")
	}
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		field.inverse(u, a)
		field.mul(u, u, a)
		if !field.equal(u, one) {
			t.Fatalf("(r * a) * r * (a ^ -1) == r)")
		}
		// v := field.new()
		// p := modulus.big()
		// field.exp(u, a, p.Sub(p, big.NewInt(2)))
		// field.inverse(v, a)
		// if !field.equal(v, u) {
		// 	t.Fatalf("a^(p-2) == a^-1")
		// }
	}
}

func TestFp2SquareRoot(t *testing.T) {
	field := newFp2()
	r := field.new()
	if field.sqrt(r, nonResidue2) {
		t.Fatalf("non residue cannot have a sqrt")
	}
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		aa, rr, r := field.new(), field.new(), field.new()
		field.square(aa, a)
		if !field.sqrt(r, aa) {
			t.Fatalf("bad sqrt 1")
		}
		field.square(rr, r)
		if !field.equal(rr, aa) {
			t.Fatalf("bad sqrt 2")
		}
	}
}

func TestFp6Serialization(t *testing.T) {
	field := newFp6(nil)
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		b, err := field.fromBytes(field.toBytes(a))
		if err != nil {
			t.Fatal(err)
		}
		if !field.equal(a, b) {
			t.Fatalf("bad serialization\n")
		}
	}
}

func TestFp6AdditionProperties(t *testing.T) {
	field := newFp6(nil)
	for i := 0; i < fuz; i++ {
		zero := field.zero()
		a, _ := field.rand(rand.Reader)
		b, _ := field.rand(rand.Reader)
		c_1 := field.new()
		c_2 := field.new()
		field.add(c_1, a, zero)
		if !field.equal(c_1, a) {
			t.Fatalf("a + 0 == a")
		}
		field.sub(c_1, a, zero)
		if !field.equal(c_1, a) {
			t.Fatalf("a - 0 == a")
		}
		field.double(c_1, zero)
		if !field.equal(c_1, zero) {
			t.Fatalf("2 * 0 == 0")
		}
		field.neg(c_1, zero)
		if !field.equal(c_1, zero) {
			t.Fatalf("-0 == 0")
		}
		field.sub(c_1, zero, a)
		field.neg(c_2, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("0-a == -a")
		}
		field.double(c_1, a)
		field.add(c_2, a, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("2 * a == a + a")
		}
		field.add(c_1, a, b)
		field.add(c_2, b, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("a + b = b + a")
		}
		field.sub(c_1, a, b)
		field.sub(c_2, b, a)
		field.neg(c_2, c_2)
		if !field.equal(c_1, c_2) {
			t.Fatalf("a - b = - ( b - a )")
		}
		c_x, _ := field.rand(rand.Reader)
		field.add(c_1, a, b)
		field.add(c_1, c_1, c_x)
		field.add(c_2, a, c_x)
		field.add(c_2, c_2, b)
		if !field.equal(c_1, c_2) {
			t.Fatalf("(a + b) + c == (a + c ) + b")
		}
		field.sub(c_1, a, b)
		field.sub(c_1, c_1, c_x)
		field.sub(c_2, a, c_x)
		field.sub(c_2, c_2, b)
		if !field.equal(c_1, c_2) {
			t.Fatalf("(a - b) - c == (a - c ) -b")
		}
	}
}

func TestFp6SparseMultiplication(t *testing.T) {
	fp6 := newFp6(nil)
	fq2 := fp6.fp2
	var a, b, u *fe6
	for j := 0; j < fuz; j++ {
		a, _ = fp6.rand(rand.Reader)
		b, _ = fp6.rand(rand.Reader)
		u, _ = fp6.rand(rand.Reader)
		fq2.copy(&b[2], fq2.zero())
		fp6.mul(u, a, b)
		fp6.mulBy01(a, a, &b[0], &b[1])
		if !fp6.equal(a, u) {
			t.Fatal("bad mul by 01")
		}
	}
	for j := 0; j < fuz; j++ {
		a, _ = fp6.rand(rand.Reader)
		b, _ = fp6.rand(rand.Reader)
		u, _ = fp6.rand(rand.Reader)
		fq2.copy(&b[2], fq2.zero())
		fq2.copy(&b[0], fq2.zero())
		fp6.mul(u, a, b)
		fp6.mulBy1(a, a, &b[1])
		if !fp6.equal(a, u) {
			t.Fatal("bad mul by 1")
		}
	}
}

func TestFp6MultiplicationProperties(t *testing.T) {
	field := newFp6(nil)
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		b, _ := field.rand(rand.Reader)
		zero := field.zero()
		one := field.one()
		c_1, c_2 := field.new(), field.new()
		field.mul(c_1, a, zero)
		if !field.equal(c_1, zero) {
			t.Fatalf("a * 0 == 0")
		}
		field.mul(c_1, a, one)
		if !field.equal(c_1, a) {
			t.Fatalf("a * 1 == a")
		}
		field.mul(c_1, a, b)
		field.mul(c_2, b, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("a * b == b * a")
		}
		c_x, _ := field.rand(rand.Reader)
		field.mul(c_1, a, b)
		field.mul(c_1, c_1, c_x)
		field.mul(c_2, c_x, b)
		field.mul(c_2, c_2, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("(a * b) * c == (a * c) * b")
		}
	}
}

func TestFp6Exponentiation(t *testing.T) {
	field := newFp6(nil)
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		u := field.new()
		field.exp(u, a, big.NewInt(0))
		if !field.equal(u, field.one()) {
			t.Fatalf("a^0 == 1")
		}
		field.exp(u, a, big.NewInt(1))
		if !field.equal(u, a) {
			t.Fatalf("a^1 == a")
		}
		v := field.new()
		field.mul(u, a, a)
		field.mul(u, u, u)
		field.mul(u, u, u)
		field.exp(v, a, big.NewInt(8))
		if !field.equal(u, v) {
			t.Fatalf("((a^2)^2)^2 == a^8")
		}
		// p := modulus.big()
		// field.exp(u, a, p)
		// if !field.equal(u, a) {
		// 	t.Fatalf("a^p == a")
		// }
		// field.exp(u, a, p.Sub(p, big.NewInt(1)))
		// if !field.equal(u, field.one()) {
		// 	t.Fatalf("a^(p-1) == 1")
		// }
	}
}

func TestFp6Inversion(t *testing.T) {
	field := newFp6(nil)
	for i := 0; i < fuz; i++ {
		u := field.new()
		zero := field.zero()
		one := field.one()
		field.inverse(u, zero)
		if !field.equal(u, zero) {
			t.Fatalf("(0^-1) == 0)")
		}
		field.inverse(u, one)
		if !field.equal(u, one) {
			t.Fatalf("(1^-1) == 1)")
		}
		a, _ := field.rand(rand.Reader)
		field.inverse(u, a)
		field.mul(u, u, a)
		if !field.equal(u, one) {
			t.Fatalf("(r*a) * r*(a^-1) == r)")
		}
		// v := field.new()
		// p := modulus.big()
		// field.exp(u, a, p.Sub(p, big.NewInt(2)))
		// field.inverse(v, a)
		// if !field.equal(v, u) {
		// 	t.Fatalf("a^(p-2) == a^-1")
		// }
	}
}

func TestFp12Serialization(t *testing.T) {
	field := newFp12(nil)
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		b, err := field.fromBytes(field.toBytes(a))
		if err != nil {
			t.Fatal(err)
		}
		if !field.equal(a, b) {
			t.Fatalf("bad serialization\n")
		}
	}
}

func TestFp12AdditionProperties(t *testing.T) {
	field := newFp12(nil)
	for i := 0; i < fuz; i++ {
		zero := field.zero()
		a, _ := field.rand(rand.Reader)
		b, _ := field.rand(rand.Reader)
		c_1 := field.new()
		c_2 := field.new()
		field.add(c_1, a, zero)
		if !field.equal(c_1, a) {
			t.Fatalf("a + 0 == a")
		}
		field.sub(c_1, a, zero)
		if !field.equal(c_1, a) {
			t.Fatalf("a - 0 == a")
		}
		field.double(c_1, zero)
		if !field.equal(c_1, zero) {
			t.Fatalf("2 * 0 == 0")
		}
		field.neg(c_1, zero)
		if !field.equal(c_1, zero) {
			t.Fatalf("-0 == 0")
		}
		field.sub(c_1, zero, a)
		field.neg(c_2, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("0-a == -a")
		}
		field.double(c_1, a)
		field.add(c_2, a, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("2 * a == a + a")
		}
		field.add(c_1, a, b)
		field.add(c_2, b, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("a + b = b + a")
		}
		field.sub(c_1, a, b)
		field.sub(c_2, b, a)
		field.neg(c_2, c_2)
		if !field.equal(c_1, c_2) {
			t.Fatalf("a - b = - ( b - a )")
		}
		c_x, _ := field.rand(rand.Reader)
		field.add(c_1, a, b)
		field.add(c_1, c_1, c_x)
		field.add(c_2, a, c_x)
		field.add(c_2, c_2, b)
		if !field.equal(c_1, c_2) {
			t.Fatalf("(a + b) + c == (a + c ) + b")
		}
		field.sub(c_1, a, b)
		field.sub(c_1, c_1, c_x)
		field.sub(c_2, a, c_x)
		field.sub(c_2, c_2, b)
		if !field.equal(c_1, c_2) {
			t.Fatalf("(a - b) - c == (a - c ) -b")
		}
	}
}

func TestFp12MultiplicationProperties(t *testing.T) {
	field := newFp12(nil)
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		b, _ := field.rand(rand.Reader)
		zero := field.zero()
		one := field.one()
		c_1, c_2 := field.new(), field.new()
		field.mul(c_1, a, zero)
		if !field.equal(c_1, zero) {
			t.Fatalf("a * 0 == 0")
		}
		field.mul(c_1, a, one)
		if !field.equal(c_1, a) {
			t.Fatalf("a * 1 == a")
		}
		field.mul(c_1, a, b)
		field.mul(c_2, b, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("a * b == b * a")
		}
		c_x, _ := field.rand(rand.Reader)
		field.mul(c_1, a, b)
		field.mul(c_1, c_1, c_x)
		field.mul(c_2, c_x, b)
		field.mul(c_2, c_2, a)
		if !field.equal(c_1, c_2) {
			t.Fatalf("(a * b) * c == (a * c) * b")
		}
	}
}

func TestFp12SparseMultiplication(t *testing.T) {
	fp12 := newFp12(nil)
	fp2 := fp12.fp2()
	var a, b, u *fe12
	for j := 0; j < fuz; j++ {
		a, _ = fp12.rand(rand.Reader)
		b, _ = fp12.rand(rand.Reader)
		u, _ = fp12.rand(rand.Reader)
		fp2.copy(&b[0][2], fp2.zero())
		fp2.copy(&b[1][0], fp2.zero())
		fp2.copy(&b[1][2], fp2.zero())
		fp12.mul(u, a, b)
		fp12.mulBy014Assign(a, &b[0][0], &b[0][1], &b[1][1])
		if !fp12.equal(a, u) {
			t.Fatal("bad mul by 01")
		}
	}
}

func TestFp12Exponentiation(t *testing.T) {
	field := newFp12(nil)
	for i := 0; i < fuz; i++ {
		a, _ := field.rand(rand.Reader)
		u := field.new()
		field.exp(u, a, big.NewInt(0))
		if !field.equal(u, field.one()) {
			t.Fatalf("a^0 == 1")
		}
		field.exp(u, a, big.NewInt(1))
		if !field.equal(u, a) {
			t.Fatalf("a^1 == a")
		}
		v := field.new()
		field.mul(u, a, a)
		field.mul(u, u, u)
		field.mul(u, u, u)
		field.exp(v, a, big.NewInt(8))
		if !field.equal(u, v) {
			t.Fatalf("((a^2)^2)^2 == a^8")
		}
		// p := modulus.big()
		// field.exp(u, a, p)
		// if !field.equal(u, a) {
		// 	t.Fatalf("a^p == a")
		// }
		// field.exp(u, a, p.Sub(p, big.NewInt(1)))
		// if !field.equal(u, field.one()) {
		// 	t.Fatalf("a^(p-1) == 1")
		// }
	}
}

func TestFp12Inversion(t *testing.T) {
	field := newFp12(nil)
	for i := 0; i < fuz; i++ {
		u := field.new()
		zero := field.zero()
		one := field.one()
		field.inverse(u, zero)
		if !field.equal(u, zero) {
			t.Fatalf("(0^-1) == 0)")
		}
		field.inverse(u, one)
		if !field.equal(u, one) {
			t.Fatalf("(1^-1) == 1)")
		}
		a, _ := field.rand(rand.Reader)
		field.inverse(u, a)
		field.mul(u, u, a)
		if !field.equal(u, one) {
			t.Fatalf("(r*a) * r*(a^-1) == r)")
		}
		// v := field.new()
		// p := modulus.big()
		// field.exp(u, a, p.Sub(p, big.NewInt(2)))
		// field.inverse(v, a)
		// if !field.equal(v, u) {
		// 	t.Fatalf("a^(p-2) == a^-1")
		// }
	}
}

func BenchmarkMultiplication(t *testing.B) {
	a, _ := newRand(rand.Reader)
	b, _ := newRand(rand.Reader)
	c, _ := newRand(rand.Reader)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		mul(c, a, b)
	}
}

func padBytes(in []byte, size int) []byte {
	out := make([]byte, size)
	if len(in) > size {
		panic("bad input for padding")
	}
	copy(out[size-len(in):], in)
	return out
}
