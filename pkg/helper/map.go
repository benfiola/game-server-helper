package helper

// Map is a typed map that has some helper methods attached to it
type Map[K comparable, V any] map[K]V

// Returns a list of keys in the [Map]
func (m *Map[K, V]) Keys() []K {
	list := []K{}
	for key := range *m {
		list = append(list, key)
	}
	return list
}

// Returns a list of values in the [Map]
func (m *Map[K, V]) Values() []V {
	list := []V{}
	for _, value := range *m {
		list = append(list, value)
	}
	return list
}
