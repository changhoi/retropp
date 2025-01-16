package notice

import (
	"context"
	"fmt"
	"github.com/changhoi/retropp/internal/calc"
	"github.com/changhoi/retropp/internal/client"
	"github.com/changhoi/retropp/internal/model"
	"github.com/jomei/notionapi"
	"strings"
)

const feeBoardID = "1e91545e18774ce8afa6da67f2c9dc78"

type Param struct {
	MemberViolations map[model.User][]*calc.Violation
	Year             int
	Quarter          int
}

func Notice(ctx context.Context, cli *client.Client, param Param) error {
	c, err := cli.C(ctx)
	if err != nil {
		return err
	}

	header := notionapi.TableRowBlock{
		BasicBlock: notionapi.BasicBlock{
			Object: notionapi.ObjectTypeBlock,
			Type:   notionapi.BlockTypeTableRowBlock,
		},
		TableRow: notionapi.TableRow{
			Cells: [][]notionapi.RichText{
				// HEADER
				{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: "날짜",
						},
					},
				},
				{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: "이유",
						},
					},
				},
				{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: "벌금",
						},
					},
				},
			},
		},
	}

	for u, v := range param.MemberViolations {
		var (
			fee  int
			rows []notionapi.Block
		)

		for _, violation := range v {
			fee += violation.Fee
			rows = append(rows, notionapi.TableRowBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeTableRowBlock,
				},
				TableRow: notionapi.TableRow{
					Cells: [][]notionapi.RichText{
						{
							{
								Type: "mention",
								Mention: &notionapi.Mention{
									Type: notionapi.MentionTypeDate,
									Date: &notionapi.DateObject{
										Start: (*notionapi.Date)(ptr(violation.Term[0])),
									},
								},
							},
							{
								Type: "mention",
								Mention: &notionapi.Mention{
									Type: notionapi.MentionTypeDate,
									Date: &notionapi.DateObject{
										Start: (*notionapi.Date)(ptr(violation.Term[1])),
									},
								},
							},
						},
						{
							{
								Type: notionapi.ObjectTypeText,
								Text: &notionapi.Text{
									Content: strings.Join(slicesMap(violation.Reason, func(t calc.Reason) string {
										return t.String()
									}), ", "),
								},
							},
						},
						{
							{
								Type: notionapi.ObjectTypeText,
								Text: &notionapi.Text{
									Content: fmt.Sprintf("%d", violation.Fee),
								},
							},
						},
					},
				},
			})
		}

		tableBlock := &notionapi.TableBlock{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeTableBlock,
			},
			Table: notionapi.Table{
				TableWidth:      3, // 날짜, 이유, 벌금
				HasColumnHeader: true,
				Children:        append([]notionapi.Block{header}, rows...),
			},
		}

		createReq := &notionapi.PageCreateRequest{
			Parent: notionapi.Parent{
				DatabaseID: feeBoardID,
			},
			Properties: notionapi.Properties{
				"제목": notionapi.TitleProperty{
					Title: []notionapi.RichText{
						{
							Type: notionapi.ObjectTypeText,
							Text: &notionapi.Text{
								Content: pageName(u.Name, param.Year, param.Quarter),
							},
						},
					},
				},
				"참가자": notionapi.PeopleProperty{
					People: []notionapi.User{
						{
							ID: notionapi.UserID(u.ID),
						},
					},
				},
				"연도": notionapi.NumberProperty{
					Type:   notionapi.PropertyTypeNumber,
					Number: float64(param.Year),
				},
				"분기": notionapi.NumberProperty{
					Type:   notionapi.PropertyTypeNumber,
					Number: float64(param.Quarter),
				},
				"벌금": notionapi.NumberProperty{
					Type:   notionapi.PropertyTypeNumber,
					Number: float64(fee),
				},
			},
			Children: []notionapi.Block{tableBlock},
		}

		if _, err := c.Page.Create(ctx, createReq); err != nil {
			return err
		}
	}

	return nil
}

func pageName(username string, year, q int) string {
	return fmt.Sprintf("%s: %d-Q%d", username, year, q)
}

func ptr[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

func slicesMap[T any, U any](slices []T, f func(T) U) []U {
	ret := make([]U, 0, len(slices))
	for _, v := range slices {
		ret = append(ret, f(v))
	}
	return ret
}
