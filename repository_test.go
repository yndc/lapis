package lapis_test

import (
	"fmt"
	"testing"
)

func TestNewRepository(t *testing.T) {
	squareMockRepository := newSquareMockRepository(t)
	r, err := squareMockRepository.LoadAll([]int{1, 2, 3, 4, 5})
	fmt.Println(r)
	fmt.Println(err)
}
