package cache

import "github.com/shipa988/hw_otus_architect/internal/domain/entity"

type newsStack struct {
	arr []entity.News
	size int
}

func newNewsStack(size int) *newsStack {
	return &newsStack{
		arr:  make([]entity.News,0,size),
		size: size,
	}
}
// IsEmpty: check if stack is empty
func (s *newsStack) IsEmpty() bool {
	return len(s.arr) == 0
}

// Push a new value onto the stack
func (s *newsStack) Push(n entity.News) {
	s.arr = append(s.arr, n) // Simply append the new value to the end of the stack
	if len(s.arr)>s.size{
		s.Pop()
	}
}

// Remove and return top element of stack. Return false if stack is empty.
func (s *newsStack) Pop() (entity.News, bool) {
	if s.IsEmpty() {
		return entity.News{}, false
	} else {
		index := len(s.arr) - 1 // Get the index of the top most element.
		element := (s.arr)[index] // Index into the slice and obtain the element.
		s.arr = (s.arr)[:index] // Remove it from the stack by slicing it off.
		return element, true
	}
}
