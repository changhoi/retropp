package calc

import (
	"fmt"
	"time"
)

type Reason int8

type Violation struct {
	Week   int
	Term   []time.Time
	Reason []Reason
	Fee    int
	Detail Detail
}

type Detail struct {
	UncommentAuthors []string
}

func (v *Violation) String() string {
	return fmt.Sprintf("%+v", *v)
}

func (a Reason) String() string {
	switch a {
	case missing:
		return "미작성"
	case late:
		return "지각"
	case uncomment:
		return "댓글 미작성"
	default:
		return "알 수 없음"
	}
}

func (a Reason) Fee() int {
	switch a {
	case missing:
		return 20000
	case late:
		return 10000
	case uncomment:
		return 10000
	}

	return 0
}

const (
	missing Reason = iota + 1
	late
	uncomment
)
