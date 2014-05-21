package main

// uniq returns a slice with all the duplicate elements in input removed;
// input must be sorted such that same elements are next to each other.
//
// TODO: This is a shoddy algorithm.
func uniq(input []string) []string {
	n := len(input)
	if n == 0 {
		return []string{}
	}

	output := make([]string, 0, n)
	var head string
	for _, v := range input {
		if head != v {
			output = append(output, head)
			head = v
		}
	}
	return output
}
