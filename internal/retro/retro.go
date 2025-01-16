package retro

import (
	"context"
	"errors"
	"github.com/changhoi/retropp/internal/client"
	"github.com/changhoi/retropp/internal/model"
	"github.com/jomei/notionapi"
	"log/slog"
	"time"
)

const (
	retroDatabaseID = "91384b34a33a4aabbdceea9a8b9a139e"
)

const (
	createdByPropertyName   = "생성자"
	createdAtPropertyName   = "생성 일시"
	submittedAtPropertyName = "작성 완료 시각"
)

func GetQuarter(ctx context.Context, cli *client.Client, year, q int) (*model.Quarter, error) {
	c, err := cli.C(ctx)
	if err != nil {
		return nil, err
	}

	quarter := NewQuarter(year, q)
	var (
		head = notionapi.Date(quarter.Head())
		tail = notionapi.Date(quarter.Tail())
	)
	slog.Debug("quarter term", slog.Any("head", head), slog.Any("tail", tail))

	res, err := c.Database.Query(ctx, retroDatabaseID, &notionapi.DatabaseQueryRequest{
		Filter: notionapi.AndCompoundFilter{
			&notionapi.PropertyFilter{
				Property: createdAtPropertyName,
				Date: &notionapi.DateFilterCondition{
					OnOrAfter: &head,
				},
			},
			&notionapi.PropertyFilter{
				Property: createdAtPropertyName,
				Date: &notionapi.DateFilterCondition{
					Before: &tail,
				},
			},
		},
		Sorts: []notionapi.SortObject{
			{
				Property:  createdAtPropertyName,
				Direction: notionapi.SortOrderASC,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(res.Results) == 0 {
		return nil, errors.New("record not found")
	}

	slog.Debug("records", slog.Int("count", len(res.Results)))

	pages := res.Results
	for _, w := range quarter.Weeks {
		slog.Debug("pages", slog.Int("count", len(pages)), slog.Any("weekStart", w.Start), slog.Any("weekEnd", w.End))
		if len(pages) == 0 {
			break
		}

		pages = makeRecordsForWeek(w, pages)
	}

	return quarter, nil
}

func makeRecordsForWeek(week *model.Week, pages []notionapi.Page) []notionapi.Page {
	for i, p := range pages {
		// TODO: use only submittedAt. if not exist, skip it.
		var (
			submittedAt time.Time
			createdAt   time.Time
		)
		submitTime := p.Properties[submittedAtPropertyName].(*notionapi.DateProperty).Date
		createdTime := p.Properties[createdAtPropertyName].(*notionapi.CreatedTimeProperty)

		if submitTime == nil {
			slog.Warn("submittedAt is nil", slog.Any("pageID", p.ID.String()))
			continue
		}

		if createdTime == nil {
			slog.Warn("createdAt is nil", slog.Any("pageID", p.ID.String()))
			continue
		}

		submittedAt = time.Time(*submitTime.Start)
		createdAt = createdTime.CreatedTime

		// submittedAt이 week.Start 이전이면 다음 record로 넘어감
		if submittedAt.Before(week.Start) {
			slog.Debug("before. skip record", slog.Any("submittedAt", submittedAt), slog.Any("weekStart", week.Start))
			continue
		}

		if createdAt.After(week.Buffer) {
			// 버퍼 이후 생성이면 더 이상의 record는 없음
			slog.Debug("after buffer. no more records", slog.Any("createdAt", createdAt), slog.Any("weekBuffer", week.Buffer))
			return pages[i:]
		}

		if submittedAt.Before(week.End) {
			week.OnTimeRecords = append(week.OnTimeRecords, &model.Record{
				ID:          p.ID.String(),
				UserID:      p.Properties[createdByPropertyName].(*notionapi.CreatedByProperty).CreatedBy.ID.String(),
				SubmittedAt: submittedAt,
			})
			continue
		}

		week.LateRecords = append(week.LateRecords, &model.Record{
			ID:          p.ID.String(),
			UserID:      p.Properties[createdByPropertyName].(*notionapi.CreatedByProperty).CreatedBy.ID.String(),
			SubmittedAt: submittedAt,
		})
	}

	return nil
}
