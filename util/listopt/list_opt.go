package listopt

import "sort"

// SplitList 平均分割一个list到num个list里
func SplitList(arr []string, num int64) [][]string {
	max := int64(len(arr))
	if max < num {
		return nil
	}
	var segmens = make([][]string, 0)
	quantity := max / num
	end := int64(0)
	for i := int64(1); i <= num; i++ {
		qu := i * quantity
		if i != num {
			segmens = append(segmens, arr[i-1+end:qu])
		} else {
			segmens = append(segmens, arr[i-1+end:])
		}
		end = qu - i
	}
	return segmens
}

// RemoveRepeatedElement 移除数组中重复的元素
func RemoveRepeatedElement(slice interface{}) []interface{} {
	// 创建一个map用于记录元素是否出现过
	seen := make(map[interface{}]bool, 256)
	// 创建一个新的切片用于存储去重后的元素
	newSlice := make([]interface{}, 0)
	// 遍历原切片的每个元素，如果该元素没有出现过，则添加到新切片中
	switch slice.(type) {
	case []string:
		for _, v := range slice.([]string) {
			if _, ok := seen[v]; !ok {
				newSlice = append(newSlice, v)
				seen[v] = true
			}
		}
	case []int:
		for _, v := range slice.([]int) {
			if _, ok := seen[v]; !ok {
				newSlice = append(newSlice, v)
				seen[v] = true
			}
		}
	case []float64:
		for _, v := range slice.([]float64) {
			if _, ok := seen[v]; !ok {
				newSlice = append(newSlice, v)
				seen[v] = true
			}
		}
		// 添加其他需要支持的类型
	}
	// 返回去重后的切片
	return newSlice
}

func RemoveDuplicateString(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		// 找不到就加入, 找到(即重复)就不加
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func RemoveDuplicateInt(languages []int) []int {
	result := make([]int, 0, len(languages))
	temp := map[int]struct{}{}
	for _, item := range languages {
		// 找不到就加入, 找到(即重复)就不加
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// DeleteString 删除列表中的字串
func DeleteString(strList []string, delStr string) []string {
	j := 0
	if len(strList) == 0 {
		return strList
	}
	for _, val := range strList {
		if val == delStr {
			break
		}
		j++
	}
	return append(strList[:j], strList[j+1:]...)
}

// IntersectStr 求交集
func IntersectStr(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times, _ := m[v]
		if times == 1 {
			nn = append(nn, v)
		}
	}
	return nn
}

// UnionStr 求并集
func UnionStr(slice1, slice2 []string) []string {
	m := make(map[string]int)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times, _ := m[v]
		if times == 0 {
			slice1 = append(slice1, v)
		}
	}
	return slice1
}

// DifferenceStr1 求差集
func DifferenceStr1(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	inter := IntersectStr(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}

	for _, value := range slice1 {
		times, _ := m[value]
		if times == 0 {
			nn = append(nn, value)
		}
	}
	return nn
}

// DifferenceStr2 求差集
func DifferenceStr2(slice1 []string, slice2 []string) []string {
	set := make(map[string]bool)

	for _, v := range slice1 {
		set[v] = true
	}

	for _, v := range slice2 {
		delete(set, v)
	}

	difference := make([]string, 0, len(set))
	for v := range set {
		difference = append(difference, v)
	}

	return difference
}

// In 判断是否在数组内
func In(strList []string, target string) bool {
	sort.Strings(strList)
	index := sort.SearchStrings(strList, target)
	if index < len(strList) && strList[index] == target {
		return true
	}
	return false
}

// ReverseStr 反序输出
func ReverseStr(arr []string) []string {
	var temp string
	length := len(arr)
	for i := 0; i < length/2; i++ {
		temp = (arr)[i]
		(arr)[i] = (arr)[length-1-i]
		(arr)[length-1-i] = temp
	}
	return arr
}
