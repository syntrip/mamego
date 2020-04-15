// Copyright 2015 Zack Guo <gizak@icloud.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package termui

import "strings"

// PointedList displays []string as its items, and resalt a pointed one
/*
  strs := []string{
		"[0] github.com/gizak/termui",
		"[1] editbox.go",
		"[2] iterrupt.go",
		"[3] keyboard.go",
		"[4] output.go",
		"[5] random_out.go",
		"[6] dashboard.go",
		"[7] nsf/termbox-go"}

  ls := termui.NewPointedList()
  ls.Items = strs
  ls.ItemFgColor = termui.ColorYellow
  ls.Border.Label = "PointedList"
  ls.Height = 7
  ls.Width = 25
  ls.Y = 0
	ls.ItemPointed 3
*/
type PointedList struct {
	Block
	Items               []string
	ItemPointed         int
	ItemFgColor         Attribute
	ItemBgColor         Attribute
	ItemSelectedFgColor Attribute
	ItemSelectedBgColor Attribute
}

// NewPointedList returns a new *PointedList with current theme.
func NewPointedList() *PointedList {
	l := &PointedList{Block: *NewBlock()}
	l.ItemFgColor = theme.ListItemFg
	l.ItemBgColor = theme.ListItemBg
	return l
}

// Buffer implements Bufferer interface.
func (l *PointedList) Buffer() []Point {
	ps := l.Block.Buffer()

	trimItems := l.Items
	if len(trimItems) > l.innerHeight {
		trimItems = trimItems[:l.innerHeight]
	}
	ifgc := l.ItemFgColor
	ibgc := l.ItemBgColor
	for i, v := range trimItems {
		if len(v) < l.innerWidth {
			v += strings.Repeat(" ", l.innerWidth-len(v))
		}
		rs := trimStr2Runes(v, l.innerWidth)

		j := 0
		if i == l.ItemPointed {
			ifgc = l.ItemSelectedFgColor
			ibgc = l.ItemSelectedBgColor
		} else {
			ifgc = l.ItemFgColor
			ibgc = l.ItemBgColor
		}
		for _, vv := range rs {
			w := charWidth(vv)
			p := Point{}
			p.X = l.innerX + j
			p.Y = l.innerY + i

			p.Ch = vv
			p.Bg = ibgc
			p.Fg = ifgc

			ps = append(ps, p)
			j += w
		}
	}

	return l.Block.chopOverflow(ps)
}
