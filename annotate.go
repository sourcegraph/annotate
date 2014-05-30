package annotate

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

type Annotation struct {
	Start, End  int
	Left, Right []byte
	WantInner   int
}

type Annotations []*Annotation

func (a Annotations) Len() int { return len(a) }
func (a Annotations) Less(i, j int) bool {
	// Sort by start position, breaking ties by preferring longer
	// matches.
	ai, aj := a[i], a[j]
	if ai.Start == aj.Start {
		if ai.End == aj.End {
			return ai.WantInner < aj.WantInner
		}
		return ai.End > aj.End
	} else {
		return ai.Start < aj.Start
	}
}
func (a Annotations) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Annotates src with annotations in anns.
//
// Annotating an empty byte array always returns an empty byte array.
//
// Assumes anns is sorted (using sort.Sort(anns)).
func Annotate(src []byte, anns Annotations, writeContent func(io.Writer, []byte)) ([]byte, error) {
	var out bytes.Buffer
	var err error

	// Keep a stack of annotations we should close at all future rune offsets.
	closeAnnsAtRune := make(map[int]Annotations, len(src)/10)

	runeCount := utf8.RuneCount(src)
	for r := 0; r < runeCount; r++ {
		// Open annotations that begin here.
		for i, a := range anns {
			if a.Start == r {
				out.Write(a.Left)

				if a.Start == a.End {
					out.Write(a.Right)
				} else {
					// Put this annotation on the stack of annotations that will need
					// to be closed. We remove it from anns at the end of the loop
					// (to avoid modifying anns while we're iterating over it).
					closeAnnsAtRune[a.End] = append(closeAnnsAtRune[a.End], a)
				}
			} else if a.Start > r {
				// Remove all annotations that we opened (we already put them on the
				// stack of annotations that will need to be closed).
				anns = anns[i:]
				break
			} else if a.Start < 0 {
				err = ErrStartOutOfBounds
			}
		}

		_, runeSize := utf8.DecodeRune(src)
		if writeContent == nil {
			out.Write(src[:runeSize])
		} else {
			writeContent(&out, src[:runeSize])
		}
		src = src[runeSize:]

		// Close annotations that after this rune.
		if closeAnns, present := closeAnnsAtRune[r+1]; present {
			for i := len(closeAnns) - 1; i >= 0; i-- {
				out.Write(closeAnns[i].Right)
			}
			delete(closeAnnsAtRune, r+1)
		}
	}

	if len(closeAnnsAtRune) > 0 {
		err = ErrEndOutOfBounds
	}

	return out.Bytes(), err
}

var (
	ErrStartOutOfBounds = errors.New("annotation start out of bounds")
	ErrEndOutOfBounds   = errors.New("annotation end out of bounds")
)
