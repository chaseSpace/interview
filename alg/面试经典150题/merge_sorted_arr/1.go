package xx

//Title：合并2个非递减有序数组 - 简单

// 最优解题思路：既然 A 后面有空座位，我们“从后往前”找最大的元素，直接放到 A 的最后一个座位，谁大谁坐，坐完就往前挪。这样不会踩到还没处理的元素。

func merge(nums1 []int, m int, nums2 []int, n int) {
	i, j := m-1, n-1
	k := m + n - 1

	for i >= 0 && j >= 0 {
		if nums1[i] > nums2[j] {
			nums1[k] = nums1[i]
			i--
		} else {
			nums1[k] = nums2[j]
			j--
		}
		k--
	}
	// 将nums2的剩余元素复制到nums1（若nums1还有剩余，则不用管，已经在前面了）
	for j >= 0 {
		nums1[k] = nums2[j]
		j--
		k--
	}
}
