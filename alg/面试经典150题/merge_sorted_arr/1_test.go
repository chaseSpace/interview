package xx

import (
	"reflect"
	"testing"
)

func Test_merge(t *testing.T) {
	type args struct {
		nums1 []int
		m     int
		nums2 []int
		n     int
	}

	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "基本测试 case 1",
			args: args{
				nums1: []int{1, 2, 3, 0, 0, 0},
				m:     3,
				nums2: []int{2, 5, 6},
				n:     3,
			},
			want: []int{1, 2, 2, 3, 5, 6},
		},
		{
			name: "nums1 为空",
			args: args{
				nums1: []int{0},
				m:     0,
				nums2: []int{1},
				n:     1,
			},
			want: []int{1},
		},
		{
			name: "nums2 为空",
			args: args{
				nums1: []int{1},
				m:     1,
				nums2: []int{},
				n:     0,
			},
			want: []int{1},
		},
		{
			name: "nums2 全部小于 nums1",
			args: args{
				nums1: []int{4, 5, 6, 0, 0, 0},
				m:     3,
				nums2: []int{1, 2, 3},
				n:     3,
			},
			want: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name: "nums1 全部小于 nums2",
			args: args{
				nums1: []int{1, 2, 3, 0, 0, 0},
				m:     3,
				nums2: []int{4, 5, 6},
				n:     3,
			},
			want: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name: "混合排列 case",
			args: args{
				nums1: []int{2, 0},
				m:     1,
				nums2: []int{1},
				n:     1,
			},
			want: []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 数组1扩容：+ 数组2长度
			nums1Copy := make([]int, tt.args.m+tt.args.n)
			copy(nums1Copy, tt.args.nums1)

			merge(nums1Copy, tt.args.m, tt.args.nums2, tt.args.n)

			if !reflect.DeepEqual(nums1Copy, tt.want) {
				t.Errorf("merge() = %v, want %v", nums1Copy, tt.want)
			}
		})
	}
}
