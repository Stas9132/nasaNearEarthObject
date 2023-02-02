package main

import (
	"fmt"
	"sort"
)

func main() {
	planet := make(map[int]string)

	planet[10460] = "Датомир"
	planet[10465] = "Татуин"
	planet[12120] = "Набу"
	planet[12240] = "Корусант"
	planet[12500] = "Альдерван"

	var keys []int
	for k, _ := range planet {
		keys = append(keys, k)
	}

	sortedKeys := sort.IntSlice(keys)

	for _, key := range sortedKeys {
		fmt.Println(planet[key])
	}
}
