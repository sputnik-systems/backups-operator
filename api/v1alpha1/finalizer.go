package v1alpha1

const (
	FinalizerName = "backups.sputnik.systems/finalizer"
)

// IsContainsFinalizer check if finalizers is set.
func IsContainsFinalizer(src []string, finalizer string) bool {
	for _, f := range src {
		if f == finalizer {
			return true
		}
	}

	return false
}

// RemoveFinalizer - removes given finalizer from finalizers list.
func RemoveFinalizer(src []string, finalizer string) []string {
	dst := make([]string, 0)

	for _, f := range src {
		if f == finalizer {
			continue
		}

		dst = append(dst, f)
	}

	return dst
}
