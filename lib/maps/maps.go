package maps

// Invert inverts a map, returning a new mapping of value -> to list of keys that mapped to those values.
func Invert[K comparable, V comparable](m map[K]V) map[V][]K {
	inverted := make(map[V][]K, len(m))
	for k, v := range m {
		inverted[v] = append(inverted[v], k)
	}
	return inverted
}
