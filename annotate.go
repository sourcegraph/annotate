package annotate

import (
	"io"
	"log"
	"sort"
)

func init() { log.SetFlags(0) }

type Annotation struct {
	Start, End  int
	Left, Right []byte
	WantInner   int
}

type annotations []*Annotation

func (a annotations) Len() int { return len(a) }
func (a annotations) Less(i, j int) bool {
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
func (a annotations) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func WithHTML(src []byte, anns []*Annotation, encode func(io.Writer, []byte), w io.Writer) error {
	sort.Sort(annotations(anns))
	_, err := annotate(src, 0, len(src), anns, encode, w)
	return err
}

func annotate(src []byte, left, right int, anns []*Annotation, encode func(io.Writer, []byte), w io.Writer) (bool, error) {
	if encode == nil {
		encode = func(w io.Writer, b []byte) { w.Write(b) }
	}

	rightmost := 0
	for i, ann := range anns {
		if ann.Start >= right {
			return i != 0, nil
		}
		if ann.End < rightmost {
			continue
		}

		if i == 0 {
			encode(w, src[left:ann.Start])
		} else {
			prev := anns[i-1]
			if prev.End >= len(src) {
				break
			}
			if ann.Start > prev.End {
				encode(w, src[prev.End:min(ann.Start, len(src))])
			}
			if ann.Start < prev.End {
				// already recursed to this one
				continue
			}
		}

		w.Write(ann.Left)

		inner, err := annotate(src, ann.Start, ann.End, anns[i+1:], encode, w)
		if err != nil {
			return false, err
		}

		if !inner {
			if ann.Start >= len(src) {
				break
			}
			b := src[ann.Start:min(ann.End, len(src))]
			encode(w, b)
		}

		w.Write(ann.Right)

		if i == len(anns)-1 {
			if ann.End < len(src) {
				encode(w, src[ann.End:min(right, len(src))])
			}
		}

		rightmost = ann.End
	}
	return len(anns) > 0, nil
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
