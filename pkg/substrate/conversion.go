// Copyright 2021 PolyCrypt GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package substrate

import (
	"fmt"
	"math/big"
)

// Dot wraps a *big.Int and provides conversion and formatting for Dot values.
type Dot struct {
	plank *big.Int
}

const (
	// PlankPerDot number of planks per Dot.
	PlankPerDot = 1e12
	// PrintPrecision is the precision with which floats are printed
	// as defined by big.Float.Text.
	PrintPrecision = 3
)

// NewDotFromPlank creates a new Dot from the given amount of Planks.
func NewDotFromPlank(plank *big.Int) *Dot {
	return &Dot{plank}
}

// NewDotsFromPlanks creates new Dots from the given amounts of Planks.
func NewDotsFromPlanks(plank ...*big.Int) []*Dot {
	ret := make([]*Dot, len(plank))
	for i, p := range plank {
		ret[i] = NewDotFromPlank(p)
	}
	return ret
}

// String formats a Dot with the correct unit.
// Works for positive and negative values.
func (d *Dot) String() string {
	prefices := []struct {
		thresh *big.Float
		prefix string
	}{
		{big.NewFloat(PlankPerDot * 1e6), "MDot"},
		{big.NewFloat(PlankPerDot * 1e3), "KDot"},
		{big.NewFloat(PlankPerDot), "Dot"},
		{big.NewFloat(PlankPerDot / 1e3), "mDot"},
		{big.NewFloat(PlankPerDot / 1e6), "uDot"},
		{big.NewFloat(1), "Plank"},
	}

	plank := new(big.Float).SetInt(d.plank)
	plankAbs := new(big.Float).Abs(plank)
	for _, prefix := range prefices {
		if plankAbs.Cmp(prefix.thresh) >= 0 {
			value := new(big.Float).Quo(plank, prefix.thresh)
			return fmt.Sprintf("%s %s", formatFloat(value), prefix.prefix)
		}
	}

	return "0 Plank"
}

// Plank converts a Dot to Planks.
func (d *Dot) Plank() *big.Int {
	return new(big.Int).Set(d.plank)
}

// Abs returns the absolute value.
func (d *Dot) Abs() *Dot {
	return NewDotFromPlank(new(big.Int).Abs(d.plank))
}

func formatFloat(f *big.Float) string {
	return f.Text('f', PrintPrecision)
}
