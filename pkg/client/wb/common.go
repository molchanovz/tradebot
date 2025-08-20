package wb

import "sort"

func sortOrdersByArticle(orders []OrderWB) {
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].Article < orders[j].Article
	})
}
