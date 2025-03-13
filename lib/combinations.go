package lib

import "math/bits"

// Obtains combinations of K-length from the N-length array.
func Combinations[T any](set []T, k int) (subsets [][]T) {
	length := uint(len(set))

	if k > len(set) {
		k = len(set)
	}

	// Go through all possible combinations of objects
	// from 1 (only first object in subset) to 2^length (all objects in subset)
	for subsetBits := 1; subsetBits < (1 << length); subsetBits++ {
		if k > 0 && bits.OnesCount(uint(subsetBits)) != k {
			continue
		}

		var subset []T

		for object := uint(0); object < length; object++ {
			// checks if object is contained in subset
			// by checking if bit 'object' is set in subsetBits
			if (subsetBits>>object)&1 == 1 {
				// add object to subset
				subset = append(subset, set[object])
			}
		}
		// add subset to subsets
		subsets = append(subsets, subset)
	}
	return subsets
}
