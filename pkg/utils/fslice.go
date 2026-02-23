package utils

import (
	"slices"
)

// FSlice functional slice - обертка над слайсом, которая позволяет применять функции к слайсу
type FSlice[T any] []T

// New возвращает новый FSlice
func New[T any](slice []T) FSlice[T] {
	return FSlice[T](slice)
}

// Len возвращает длину слайса
func (s FSlice[T]) Len() int { return len(s) }

// IsEmpty возвращает true, если слайс пустой
func (s FSlice[T]) IsEmpty() bool { return s.Len() == 0 }

// Map применяет функцию к каждому элементу слайса с индексом и возвращает новый слайс
func (s FSlice[T]) Map(f func(int, T) T) FSlice[T] {
	if s.IsEmpty() {
		return s
	}
	res := make(FSlice[T], 0, s.Len())
	for i, v := range s {
		res = append(res, f(i, v))
	}
	return res
}

// Filter применяет функцию к каждому элементу слайса и возвращает новый слайс,
// содержащий только те элементы, для которых функция вернула true
func (s FSlice[T]) Filter(f func(T) bool) FSlice[T] {
	if s.IsEmpty() {
		return s
	}
	res := make(FSlice[T], 0, s.Len())
	for _, v := range s {
		if f(v) {
			res = append(res, v)
		}
	}
	return res
}

// Reduce применяет функцию к каждому элементу слайса и возвращает единственное значение
func (s FSlice[T]) Reduce(f func(current T, element T) T, initial T) T {
	if s.IsEmpty() {
		return initial
	}
	res := initial
	for _, v := range s {
		res = f(res, v)
	}
	return res
}

// Reverse меняет порядок элементов в слайсе на обратный
func (s FSlice[T]) Reverse() FSlice[T] {
	res := s.Clone()
	if res.IsEmpty() {
		return res
	}
	slices.Reverse(res)
	return res
}

// FindAll применяет функцию к каждому элементу слайса и возвращает новый слайс,
// содержащий только те элементы, для которых функция вернула true
func (s FSlice[T]) FindAll(f func(T) bool) FSlice[T] {
	if s.IsEmpty() {
		return s
	}
	res := make(FSlice[T], 0, s.Len())
	for _, v := range s {
		if f(v) {
			res = append(res, v)
		}
	}
	return res
}

// FindAllIndexes применяет функцию к каждому элементу слайса и возвращает новый слайс,
// содержащий индексы тех элементов, для которых функция вернула true
func (s FSlice[T]) FindAllIndexes(f func(T) bool) FSlice[int] {
	if s.IsEmpty() {
		return make(FSlice[int], 0)
	}
	res := make(FSlice[int], 0, s.Len())
	for i, v := range s {
		if f(v) {
			res = append(res, i)
		}
	}
	return res
}

// FindFirst применяет функцию к каждому элементу слайса и возвращает первый элемент,
// для которого функция вернула true, и true, если такой элемент был найден,
// иначе возвращает нулевое значение типа T и false
func (s FSlice[T]) FindFirst(f func(T) bool) (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	for _, v := range s {
		if f(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindFirstIndex применяет функцию к каждому элементу слайса и возвращает индекс первого элемента,
// для которого функция вернула true, и true, если такой элемент был найден,
// иначе возвращает -1 и false
func (s FSlice[T]) FindFirstIndex(f func(T) bool) (int, bool) {
	if s.IsEmpty() {
		return -1, false
	}
	for i, v := range s {
		if f(v) {
			return i, true
		}
	}
	return -1, false
}

// FindLast применяет функцию к каждому элементу слайса и возвращает последний элемент,
// для которого функция вернула true, и true, если такой элемент был найден,
// иначе возвращает нулевое значение типа T и false
func (s FSlice[T]) FindLast(f func(T) bool) (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	for i := s.Len() - 1; i >= 0; i-- {
		if f(s[i]) {
			return s[i], true
		}
	}
	var zero T
	return zero, false
}

// FindLastIndex применяет функцию к каждому элементу слайса и возвращает индекс последнего элемента,
// для которого функция вернула true, и true, если такой элемент был найден,
// иначе возвращает -1 и false
func (s FSlice[T]) FindLastIndex(f func(T) bool) (int, bool) {
	if s.IsEmpty() {
		return -1, false
	}
	for i := s.Len() - 1; i >= 0; i-- {
		if f(s[i]) {
			return i, true
		}
	}
	return -1, false
}

// Sort сортирует слайс по заданному критерию и возвращает новый отсортированный слайс
func (s FSlice[T]) Sort(f func(a, b T) int) FSlice[T] {
	res := s.Clone()
	if res.IsEmpty() {
		return res
	}
	slices.SortFunc(res, f)
	return res
}

// Contains применяет функцию к каждому элементу слайса и возвращает true,
// если хотя бы один элемент удовлетворяет условию, иначе возвращает false
func (s FSlice[T]) Contains(f func(T) bool) bool {
	return slices.ContainsFunc(s, f)
}

// Push добавляет элемент в конец слайса
func (s FSlice[T]) Push(v T) {
	s = append(s, v)
}

// Pop удаляет последний элемент слайса и возвращает его
func (s FSlice[T]) Pop() T {
	if s.IsEmpty() {
		var zero T
		return zero
	}
	res := s[s.Len()-1]
	s = s[:s.Len()-1]
	return res
}

// Concat объединяет слайсы и возвращает новый слайс
func (s FSlice[T]) Concat(sl FSlice[T]) FSlice[T] {
	return slices.Concat(s, sl)
}

func (s FSlice[T]) Clone() FSlice[T] {
	return slices.Clone(s)
}
