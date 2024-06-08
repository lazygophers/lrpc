package lrpc_test

import (
	"github.com/lazygophers/lrpc"
	"gotest.tools/v3/assert"
	"testing"
)

func TestSearchTree(t *testing.T) {
	tree := lrpc.NewSearchTree[int]()

	tree.Add("/", 1)
	tree.Add("/user", 2)
	tree.Add("/:user", 3)
	tree.Add("/:user/profile", 4)
	tree.Add("/:user/profile/:name", 5)
	tree.Add("/:user/profile/name", 6)

	type want struct {
		notfound bool
		result   *lrpc.SearchResult[int]
	}

	var (
		tests = []struct {
			name string
			want want
		}{
			{
				name: "/",
				want: want{
					result: &lrpc.SearchResult[int]{
						Item: 1,
					},
				},
			},
			{
				name: "/user",
				want: want{
					result: &lrpc.SearchResult[int]{
						Item: 2,
					},
				},
			},
			{
				name: "/123",
				want: want{
					result: &lrpc.SearchResult[int]{
						Item: 3,
						Params: map[string]string{
							"user": "123",
						},
					},
				},
			},
			{
				name: "/123/profile",
				want: want{
					result: &lrpc.SearchResult[int]{
						Item: 4,
						Params: map[string]string{
							"user": "123",
						},
					},
				},
			},
			{
				name: "/123/profile/456",
				want: want{
					result: &lrpc.SearchResult[int]{
						Item: 5,
						Params: map[string]string{
							"user": "123",
							"name": "456",
						},
					},
				},
			},
			{
				name: "/123/profile/name",
				want: want{
					result: &lrpc.SearchResult[int]{
						Item: 6,
						Params: map[string]string{
							"user": "123",
						},
					},
				},
			},
		}
	)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := tree.Search(tt.name)
			assert.Equal(t, ok, !tt.want.notfound)
			if ok {
				assert.DeepEqual(t, result, tt.want.result)
			}
		})
	}
}
