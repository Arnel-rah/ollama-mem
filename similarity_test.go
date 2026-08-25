package main

import (
	"math"
	"testing"
)

func cosineSimTest(a, b []float64) float64 {
	var aNormSq float64
	for _, v := range a {
		aNormSq += v * v
	}
	return cosineSim(a, b, math.Sqrt(aNormSq))
}

func TestCosineSimIdentical(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{1, 2, 3}
	got := cosineSimTest(a, b)
	if math.Abs(got-1.0) > 1e-9 {
		t.Errorf("cosineSim(identical vectors) = %v, want 1.0", got)
	}
}

func TestCosineSimOrthogonal(t *testing.T) {
	a := []float64{1, 0}
	b := []float64{0, 1}
	got := cosineSimTest(a, b)
	if math.Abs(got-0.0) > 1e-9 {
		t.Errorf("cosineSim(orthogonal vectors) = %v, want 0.0", got)
	}
}

func TestCosineSimOpposite(t *testing.T) {
	a := []float64{1, 1}
	b := []float64{-1, -1}
	got := cosineSimTest(a, b)
	if math.Abs(got-(-1.0)) > 1e-9 {
		t.Errorf("cosineSim(opposite vectors) = %v, want -1.0", got)
	}
}

func TestCosineSimZeroVector(t *testing.T) {
	a := []float64{0, 0, 0}
	b := []float64{1, 2, 3}
	got := cosineSimTest(a, b)
	if got != 0 {
		t.Errorf("cosineSim(zero vector) = %v, want 0 (no division by zero)", got)
	}
}

func TestCosineSimSimilarButNotIdentical(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{2, 4, 6}
	got := cosineSimTest(a, b)
	if math.Abs(got-1.0) > 1e-9 {
		t.Errorf("cosineSim(parallel vectors) = %v, want 1.0 (direction matters, not magnitude)", got)
	}
}
