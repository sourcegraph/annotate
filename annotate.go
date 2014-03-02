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
	runes := []rune(string(src))
	var annotate1 func(left, right int, anns []*Annotation, encode func(io.Writer, []byte), w io.Writer, seen map[*Annotation]struct{}) (bool, error)
	annotate1 = func(left, right int, anns []*Annotation, encode func(io.Writer, []byte), w io.Writer, seen map[*Annotation]struct{}) (bool, error) {
		if encode == nil {
			encode = func(w io.Writer, b []byte) { w.Write(b) }
		}

		rightmost := 0
		for i, ann := range anns {
			if _, exist := seen[ann]; exist {
				continue
			}
			if ann.Start >= right {
				return i != 0, nil
			}
			if ann.End < rightmost {
				continue
			}

			if i == 0 {
				encode(w, []byte(string(runes[left:ann.Start])))
			} else {
				prev := anns[i-1]
				if prev.End >= len(runes) {
					break
				}
				if ann.Start > prev.End {
					encode(w, []byte(string(runes[prev.End:min(ann.Start, len(runes))])))
				}
			}

			w.Write(ann.Left)

			inner, err := annotate1(ann.Start, ann.End, anns[i+1:], encode, w, seen)
			if err != nil {
				return false, err
			}

			if !inner {
				if ann.Start >= len(runes) {
					break
				}
				b := []byte(string(runes[ann.Start:min(ann.End, len(runes))]))
				encode(w, b)
			}

			w.Write(ann.Right)
			seen[ann] = struct{}{}

			if i == len(anns)-1 {
				if ann.End < len(runes) {
					// TODO(sqs): fix this. it chops off a portion of an
					// annotation.
					if right < ann.End {
						right = ann.End
					}
					encode(w, []byte(string(runes[ann.End:min(right, len(runes))])))
				}
			}

			rightmost = ann.End
		}
		return len(anns) > 0, nil
	}

	seen := make(map[*Annotation]struct{})
	return annotate1(left, right, anns, encode, w, seen)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
