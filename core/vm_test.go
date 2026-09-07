package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack_Push_Pop(t *testing.T) {
	stack := NewStack(1024)

	stack.Push(1)
	stack.Push(2)

	fmt.Println(stack)

	value := stack.Pop()
	assert.Equal(t, value.(int), 2)
	fmt.Println(stack)
}

func TestVM(t *testing.T) {
	// 1 + 2 = 3
	// 0x01
	// push stack 0x0a
	// 0x02
	// push stack 0x0a
	// add 0x0b
	// 3
	// push stack
	
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	vm := NewVM(data)
	assert.Nil(t, vm.Run())

	assert.Equal(t, int(3), vm.stack.data[vm.stack.sp - 1])
}