package core

import (
	"testing"

	"github.com/ch1hiro4002/Block-Chain/util"
	"github.com/stretchr/testify/assert"
)

func TestStack_Push_Pop(t *testing.T) {
	stack := NewStack(1024)

	stack.Push(1)
	stack.Push(2)

	// fmt.Println(stack)

	value := stack.Pop()
	assert.Equal(t, value.(int), 2)
	// fmt.Println(stack)
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
	contractState := NewState()
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	assert.Equal(t, int(3), vm.stack.data[vm.stack.sp-1])
}

func TestVM_Int_Sub(t *testing.T) {
	// 3 - 2 = 1

	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0c}
	contractState := NewState()
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	assert.Equal(t, int(1), vm.stack.data[vm.stack.sp-1])
}

func TestVM_Byte_Pack(t *testing.T) {
	// stack: [byte, byte, byte, size]
	// "FOO" ==> []byte{0x46, 0x4f, 0x4f}

	data := []byte{0x46, 0x0d, 0x4f, 0x0d, 0x4f, 0x0d, 0x03, 0x0a, 0x0e, 0x03, 0x0a, 0x02, 0x0a, 0x0c}
	contractState := NewState()
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	result1 := vm.stack.Pop().(int)
	assert.Equal(t, 1, result1)
	result2 := vm.stack.Pop().([]byte)
	assert.Equal(t, "FOO", string(result2))
}

func TestVM_State(t *testing.T) {
	// State: map[string][]byte
	// "FOO" => []byte{3 - 2}

	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0c, 0x46, 0x0d, 0x4f, 0x0d, 0x4f, 0x0d, 0x03, 0x0a, 0x0e, 0x0f}
	contractState := NewState()
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	value, err := vm.contractState.Get([]byte("FOO"))
	assert.Nil(t, err)
	assert.Equal(t, value, util.SerializeInt64(int64(1)))
}
