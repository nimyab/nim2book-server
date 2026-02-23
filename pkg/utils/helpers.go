package utils

// Val возвращает значение указателя, если оно не nil, иначе возвращает fallback
func Val[T any](ptr *T, fallback T) T {
	if ptr == nil {
		return fallback
	}
	return *ptr
}

// Ptr возвращает указатель на значение
func Ptr[T any](v T) *T { return &v }
