package model

import (
	"fmt"
	"time"
)

type Record struct {
	ID          string
	UserID      string
	UserName    string
	SubmittedAt time.Time
	Commenters  []*User
}

func (r *Record) String() string {
	return fmt.Sprintf("%+v", *r)
}

type Week struct {
	Index      int
	Start      time.Time
	End        time.Time
	Buffer     time.Time
	CommentEnd time.Time

	OnTimeRecords []*Record
	LateRecords   []*Record
}

func (w *Week) String() string {
	return fmt.Sprintf("%+v", *w)
}

type Quarter struct {
	Year  int
	Q     int
	Weeks []*Week
}

func (q *Quarter) Head() time.Time {
	return q.Weeks[0].Start
}

func (q *Quarter) Tail() time.Time {
	return q.Weeks[len(q.Weeks)-1].Buffer
}

type User struct {
	ID   string
	Name string
}

func (u *User) String() string {
	return fmt.Sprintf("%+v", *u)
}
