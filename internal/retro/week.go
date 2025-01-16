package retro

import (
	"github.com/changhoi/retropp/internal/model"
	"time"
)

var loc, _ = time.LoadLocation("Asia/Seoul")

func quarterTerm(year, q int) (start, end time.Time) {
	start = time.Date(year, time.Month((q-1)*3+1), 1, 0, 0, 0, 0, loc)
	end = start.AddDate(0, 3, 0)
	return start, end
}

func quarterWeeks(year, q int) []*model.Week {
	// 주 중 금요일 00시 00분 ~ 화요일 11시 59분 구간을 한 주의 구간으로 만들 것임

	// 쿼터 시작일
	start, end := quarterTerm(year, q)

	// 쿼터 시작일이 목, 금, 토, 일인 경우 다음 주 금요일로 시작일을 조정
	switch start.Weekday() {
	case time.Thursday:
		start = start.AddDate(0, 0, 8)
	case time.Friday:
		start = start.AddDate(0, 0, 7)
	case time.Saturday:
		start = start.AddDate(0, 0, 6)
	case time.Sunday:
		start = start.AddDate(0, 0, 5)
	case time.Monday:
		start = start.AddDate(0, 0, 4)
	case time.Tuesday:
		start = start.AddDate(0, 0, 3)
	case time.Wednesday:
		start = start.AddDate(0, 0, 2)
	}

	var ret []*model.Week
	index := 1
	for start.Before(end) {
		ret = append(ret, &model.Week{
			Index:      index,
			Start:      start,                                      // 금 00시 00분
			End:        start.AddDate(0, 0, 3).Add(time.Hour * 12), // 월 12시 00분
			CommentEnd: start.AddDate(0, 0, 5).Add(time.Hour * 12), // 수 12시 00분
			Buffer:     start.AddDate(0, 0, 6),                     // 목 00시 00분
		})
		start = start.AddDate(0, 0, 7)
		index++
	}

	return ret
}

func NewQuarter(year, q int) *model.Quarter {
	return &model.Quarter{
		Year:  year,
		Q:     q,
		Weeks: quarterWeeks(year, q),
	}
}
