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

func TestVM_Int_Add(t *testing.T) {
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

	assert.Equal(t, int(3), vm.stack.data[vm.stack.sp-1])
}

func TestVM_Int_Sub(t *testing.T) {
	// 3 - 2 = 1

	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0c}
	vm := NewVM(data)
	assert.Nil(t, vm.Run())

	assert.Equal(t, int(1), vm.stack.data[vm.stack.sp-1])
}

func TestVM_Byte_Pack(t *testing.T) {
	// stack: [byte, byte, byte, size]

	data := []byte{0x46, 0x0d, 0x4f, 0x0d, 0x4f, 0x0d, 0x03, 0x0a, 0x0e}
	vm := NewVM(data)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop().([]byte)
	assert.Equal(t, "FOO", string(result))
}
