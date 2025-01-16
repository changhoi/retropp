package retro

import (
	"github.com/changhoi/retropp/internal/model"
	"reflect"
	"testing"
	"time"
)

func Test_quarterWeeks(t *testing.T) {
	type args struct {
		year int
		q    int
	}
	tests := []struct {
		name string
		args args
		want []*model.Week
	}{
		{
			name: "2024 Q1",
			args: args{year: 2024, q: 1},
			want: []*model.Week{
				{
					Index:      1,
					Start:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
					End:        time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC),
					Buffer:     time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					CommentEnd: time.Date(2024, 1, 9, 12, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quarterWeeks(tt.args.year, tt.args.q); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("quarterWeeks() = %v, want %v", got, tt.want)
			}
		})
	}
}
