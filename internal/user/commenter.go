package user

import (
	"context"
	"encoding/gob"
	"errors"
	"github.com/changhoi/retropp/internal/client"
	"github.com/changhoi/retropp/internal/model"
	"github.com/jomei/notionapi"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

func SetCommenters(ctx context.Context, cli *client.Client, q *model.Quarter) error {
	for _, w := range q.Weeks {
		slog.Info("get commenters", slog.Int("week", w.Index))
		for _, r := range w.OnTimeRecords {
			commenters, err := GetCommenters(ctx, cli, r.ID)
			if err != nil {
				return err
			}
			r.Commenters = commenters
		}
	}

	return nil
}

func GetCommenters(ctx context.Context, cli *client.Client, pageID string) ([]*model.User, error) {
	var cached []*model.User
	if err := useCache(pageID, &cached); err == nil {
		slog.Info("use cache", slog.String("pageID", pageID))
		return cached, nil
	}

	c, err := cli.C(ctx)
	if err != nil {
		return nil, err
	}
	slog.Debug("get comments", slog.String("pageID", pageID))

	res, err := c.Comment.Get(ctx, notionapi.BlockID(pageID), &notionapi.Pagination{
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}

	var (
		users []*model.User
		set   = make(map[string]struct{})
	)
	for _, comment := range res.Results {
		if _, ok := set[comment.CreatedBy.ID.String()]; ok {
			continue
		}
		set[comment.CreatedBy.ID.String()] = struct{}{}
		users = append(users, &model.User{
			ID:   comment.CreatedBy.ID.String(),
			Name: comment.CreatedBy.Name,
		})
	}

	slog.Info("save cache", slog.String("pageID", pageID))
	saveCache(pageID, users)
	return users, nil
}

func useCache(pageID string, v any) error {
	f, err := os.Open(filepath.Join(".cache", "comments", pageID))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Warn("open cache", slog.Any("err", err))
		}

		return err
	}

	if err := gob.NewDecoder(f).Decode(v); err != nil {
		return err
	}

	return nil
}

func saveCache(pageID string, v any) {
	commentsPath := filepath.Join(".cache", "comments", pageID)

	if err := os.MkdirAll(filepath.Dir(commentsPath), os.ModePerm); err != nil {
		slog.Warn("make cache dir", slog.Any("err", err))
		return
	}

	f, err := os.Create(commentsPath)
	if err != nil {
		slog.Warn("create cache", slog.Any("err", err))
		return
	}

	defer func() {
		if err := f.Close(); err != nil {
			slog.Warn("close cache", slog.Any("err", err))
		}
	}()

	if err := gob.NewEncoder(f).Encode(v); err != nil {
		slog.Warn("encode cache", slog.Any("err", err))
		return
	}
}
