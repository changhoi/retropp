package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/changhoi/retropp/internal/client"
	"github.com/changhoi/retropp/internal/model"
	"github.com/jomei/notionapi"
)

const (
	participantDatabaseID = "a9aa08c2425c4cc7812294d2dc8602ce"
)

const (
	participantPropertyKey = "참가자"
)

func GetParticipants(ctx context.Context, cli *client.Client, year, q int) ([]*model.User, error) {
	c, err := cli.C(ctx)
	if err != nil {
		return nil, err
	}

	res, err := c.Database.Query(ctx, participantDatabaseID, &notionapi.DatabaseQueryRequest{
		Filter: &notionapi.PropertyFilter{
			Property: "이름",
			RichText: &notionapi.TextFilterCondition{
				Equals: pageName(year, q),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	if len(res.Results) != 1 {
		return nil, errors.New("participants query result is not 1")
	}
	page := res.Results[0]
	people, ok := page.Properties[participantPropertyKey].(*notionapi.PeopleProperty)
	if !ok {
		return nil, errors.New("participants property is not people")
	}

	ret := make([]*model.User, 0, len(people.People))
	for _, person := range people.People {
		ret = append(ret, &model.User{
			ID:   person.ID.String(),
			Name: person.Name,
		})
	}

	return ret, nil
}

func pageName(year, q int) string {
	return fmt.Sprintf("%d-Q%d", year, q)
}
