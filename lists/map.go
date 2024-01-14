package lists

func Map[T, U any](slice []T, f func(T) U) []U {
	result := make([]U, len(slice))
	for i, x := range slice {
		result[i] = f(x)
	}
	return result
}

func Fold[T, R any](slice []T, initial R, operation func(R, T) R) R {
	r := initial
	for _, t := range slice {
		r = operation(r, t)
	}
	return r
}
