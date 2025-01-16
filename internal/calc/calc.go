package calc

import (
	"github.com/changhoi/retropp/internal/model"
	"log/slog"
	"slices"
	"time"
)

type (
	userID          = string
	commentWriterID = string
	retroWriterID   = string
)

func CalculationV2(q *model.Quarter, users []*model.User, omitWeeks []int) map[model.User][]*Violation {
	slog.Debug("initialize violation map")
	ret := make(map[model.User][]*Violation, len(users))
	userMap := make(map[userID]*model.User, len(users))
	for _, u := range users {
		ret[*u] = make([]*Violation, 0, len(q.Weeks))
		userMap[u.ID] = u
	}

	for _, w := range q.Weeks {
		if slices.Contains(omitWeeks, w.Index) {
			slog.Debug("omit week", slog.Int("week", w.Index))
			continue
		}

		retroMap := make(map[retroWriterID]struct{}, len(users))
		commentMap := make(map[retroWriterID]map[commentWriterID]bool, len(users))
		weeklyViolationResult := make(map[userID]*Violation, len(users))
		for _, u := range users {
			weeklyViolationResult[u.ID] = &Violation{
				Week: w.Index,
				Term: []time.Time{w.Start, w.End},
			}
		}

		// retro check, find users uncommenting
		for _, r := range w.OnTimeRecords {
			retroMap[r.UserID] = struct{}{}
			commentMap[r.UserID] = make(map[string]bool, len(users))
			for _, u := range users {
				commentMap[r.UserID][u.ID] = false
			}
			commentMap[r.UserID][r.UserID] = true

			for _, c := range r.Commenters {
				commentMap[r.UserID][c.ID] = true
			}
		}

		// late check
		for _, r := range w.LateRecords {
			retroMap[r.UserID] = struct{}{}
			v := weeklyViolationResult[r.UserID]
			v.Reason = append(v.Reason, late)
			v.Fee += late.Fee()
		}

		// missing check
		for _, u := range users {
			if _, ok := retroMap[u.ID]; ok {
				continue
			}

			v := weeklyViolationResult[u.ID]
			v.Reason = append(v.Reason, missing)
			v.Fee += missing.Fee()
		}

		// uncomment check
		uncommentVisited := make(map[commentWriterID]struct{}, len(users))
		for retroWriter, commenterMap := range commentMap {
			for commentWriter, doComment := range commenterMap {
				if doComment {
					continue
				}
				author := userMap[retroWriter]
				v := weeklyViolationResult[commentWriter]

				if _, ok := uncommentVisited[commentWriter]; ok {
					v.Detail.UncommentAuthors = append(v.Detail.UncommentAuthors, author.Name)
					continue
				}

				uncommentVisited[commentWriter] = struct{}{}
				v.Reason = append(v.Reason, uncomment)
				v.Fee += uncomment.Fee()
				v.Detail.UncommentAuthors = append(v.Detail.UncommentAuthors, author.Name)
			}
		}

		for uid, v := range weeklyViolationResult {
			if len(v.Reason) == 0 {
				continue
			}

			userKey := *userMap[uid]
			ret[userKey] = append(ret[userKey], v)
		}
	}

	return ret
}
