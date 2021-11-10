package sat

func drand(seed *float64) float64 {
	*seed *= 1389796
	q := int(*seed / 2147483647)
	*seed -= float64(q) * 2147483647
	return *seed / 2147483647
}

func irand(seed *float64, size int) int {
	return int(drand(seed) * float64(size))
}
