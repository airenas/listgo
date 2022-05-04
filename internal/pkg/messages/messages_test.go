package messages

import (
	"reflect"
	"testing"
)

func TestDropTag(t *testing.T) {
	type args struct {
		tags []Tag
		key  string
	}
	tests := []struct {
		name        string
		args        args
		want        []Tag
		wantDropped bool
	}{
		{name: "Drops", args: args{tags: []Tag{NewTag("olia", "v"), NewTag("olia1", "v")}, key: "olia"},
			want: []Tag{NewTag("olia1", "v")}, wantDropped: true},
		{name: "Leaves", args: args{tags: []Tag{NewTag("olia", "v"), NewTag("olia1", "v")}, key: "olia2"},
			want: []Tag{NewTag("olia", "v"), NewTag("olia1", "v")}, wantDropped: false},
		{name: "Empty", args: args{tags: []Tag{}, key: "olia"}, want: nil, wantDropped: false},
		{name: "Nil", args: args{tags: nil, key: "olia"}, want: nil, wantDropped: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := DropTag(tt.args.tags, tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DropTag() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantDropped {
				t.Errorf("DropTag() got1 = %v, want %v", got1, tt.wantDropped)
			}
		})
	}
}
