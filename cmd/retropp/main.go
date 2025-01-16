package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/changhoi/retropp/internal/calc"
	"github.com/changhoi/retropp/internal/client"
	"github.com/changhoi/retropp/internal/model"
	"github.com/changhoi/retropp/internal/notice"
	"github.com/changhoi/retropp/internal/retro"
	"github.com/changhoi/retropp/internal/user"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

var (
	cli    *client.Client
	debug  bool
	record bool
)

func init() {
	key, ok := os.LookupEnv("NOTION_API_KEY")
	if !ok {
		panic("missing NOTION_API_KEY")
	}

	cli = client.New(key)

	debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	flag.BoolVar(&record, "notice", false, "send notice")
	flag.Parse()
}

func main() {
	const (
		year    = 2024
		quarter = 4
	)

	var omitWeeks = []int{}

	slog.Info("start retropp violation checker")
	ctx := context.Background()

	slog.Info("get participants...")
	users, err := user.GetParticipants(ctx, cli, year, quarter)
	if err != nil {
		panic(err)
	}
	slog.Info("participants", slog.Int("count", len(users)))

	slog.Info("get quarter...")
	q, err := retro.GetQuarter(ctx, cli, year, quarter)
	if err != nil {
		panic(err)
	}

	setUserOnQuarter(q, users)

	slog.Info("set commenters...")
	if err := user.SetCommenters(ctx, cli, q); err != nil {
		panic(err)
	}

	slog.Info("finish setting commenters")

	res := calc.CalculationV2(q, users, omitWeeks)
	for u, violations := range res {
		// use string builder
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s(ID: %s)\n", u.Name, u.ID))
		for _, v := range violations {
			sb.WriteString(fmt.Sprintf("\t%+v\n", v))
		}
		slog.Info(sb.String())
	}

	if !record {
		return
	}

	if err := notice.Notice(ctx, cli, notice.Param{
		MemberViolations: res,
		Year:             q.Year,
		Quarter:          q.Q,
	}); err != nil {
		panic(err)
	}
}

func setUserOnQuarter(q *model.Quarter, users []*model.User) {
	userSet := make(map[string]*model.User, len(users))
	for _, u := range users {
		userSet[u.ID] = u
	}

	for _, w := range q.Weeks {
		for _, rec := range w.OnTimeRecords {
			rec.UserName = userSet[rec.UserID].Name
		}

		for _, rec := range w.LateRecords {
			rec.UserName = userSet[rec.UserID].Name
		}
	}
}
