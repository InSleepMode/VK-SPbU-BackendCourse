package main

import "fmt"

func twoSum(nums []int, target int) []int {
	hashTable := map[int]int{}
	for i := 0; i < len(nums); i++ {
		difference := target - nums[i]
		if j, isFound := hashTable[difference]; isFound {
			return []int{j, i}
		}
		hashTable[nums[i]] = i
	}
	return nil
}

func main() {
	nums := []int{2, 7, 11, 15}
	target := 9
	result := twoSum(nums, target)
	fmt.Printf("[%d,%d]", result[0], result[1])
}
