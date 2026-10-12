package native

import (
	"reflect"
	"unsafe"

	"github.com/tetratelabs/wazero/experimental"
)

// cloneSnapshot deep-copies a wazero snapshot so it can be restored more than
// once.
//
// wazero exposes no clone or reusable restore, so this reaches into its
// unexported snapshot struct with reflect and unsafe. The optimizing compiler's
// snapshot holds raw pointers into its own saved machine stack, so after
// copying the stack the clone rebases sp, fp, top and the saved registers onto
// the copy. Two cleaner shapes fail: the wazero interpreter breaks vfork, and a
// fresh snapshot per vfork child cannot help because the reused snapshot is the
// guest's own setjmp. TestBuildrootBusyboxSequentialSynchronousVfork fails
// loudly if a wazero update changes the snapshot layout. Replace this when
// wazero offers a public clone.
func cloneSnapshot(snapshot experimental.Snapshot) experimental.Snapshot {
	// Only a non-nil pointer to a struct can be cloned.
	val := reflect.ValueOf(snapshot)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return snapshot
	}

	// Return the clone, or the original when the clone is not a snapshot.
	cloned := cloneStructPointer(val)
	if !cloned.IsValid() {
		return snapshot
	}
	if snap, ok := cloned.Interface().(experimental.Snapshot); ok {
		return snap
	}
	return snapshot
}

// cloneStructPointer returns a pointer to a copy of the struct val points to.
func cloneStructPointer(val reflect.Value) reflect.Value {
	// Only a non-nil pointer to a struct can be cloned.
	val = accessible(val)
	if val.Kind() != reflect.Pointer || val.IsNil() || val.Elem().Kind() != reflect.Struct {
		return val
	}

	// Copy the struct, rebasing a machine stack's pointers onto the copy.
	cloned := reflect.New(val.Elem().Type())
	if hasMachineStack(val.Elem()) {
		cloneMachineSnapshot(cloned.Elem(), val.Elem())
		return cloned
	}
	cloneStruct(cloned.Elem(), val.Elem())
	return cloned
}

// hasMachineStack reports whether src is an optimizing compiler snapshot: a
// byte slice stack with a uintptr stack pointer into it.
func hasMachineStack(src reflect.Value) bool {
	stack := accessible(src.FieldByName("stack"))
	sp := accessible(src.FieldByName("sp"))
	return stack.IsValid() &&
		stack.Kind() == reflect.Slice &&
		stack.Type().Elem().Kind() == reflect.Uint8 &&
		sp.IsValid() &&
		sp.Kind() == reflect.Uintptr
}

// cloneMachineSnapshot copies an optimizing compiler snapshot from src to dst
// and rebases its stack pointers onto the copied stack.
func cloneMachineSnapshot(dst, src reflect.Value) {
	// Copy the stack and note where the old and new stacks live.
	src = accessible(src)
	stack := accessible(src.FieldByName("stack"))
	clonedStack := cloneSlice(stack)
	var reloc stackRelocation
	if stack.Len() > 0 && clonedStack.Len() > 0 {
		reloc.oldBase = stack.Pointer()
		reloc.oldEnd = reloc.oldBase + uintptr(stack.Len())
		reloc.newBase = clonedStack.Pointer()
	}

	// Copy every field, cloning slices.
	for i := range src.NumField() {
		field := src.Field(i)
		switch {
		case src.Type().Field(i).Name == "stack":
			setValue(dst.Field(i), clonedStack)
		case field.Kind() == reflect.Slice:
			setValue(dst.Field(i), cloneSlice(field))
		default:
			setValue(dst.Field(i), field)
		}
	}

	// Rebase the pointers into the stack.
	if reloc.oldBase == 0 {
		return
	}
	for _, name := range []string{"sp", "fp", "top", "savedRegisters"} {
		reloc.apply(dst.FieldByName(name))
	}
}

// cloneStruct copies the struct src to dst, cloning slices.
func cloneStruct(dst, src reflect.Value) {
	src = accessible(src)
	for i := range src.NumField() {
		field := src.Field(i)
		if field.Kind() == reflect.Slice {
			setValue(dst.Field(i), cloneSlice(field))
			continue
		}
		setValue(dst.Field(i), field)
	}
}

// cloneSlice copies a slice, cloning the structs its pointer elements point
// to.
func cloneSlice(val reflect.Value) reflect.Value {
	// A nil slice stays nil.
	val = accessible(val)
	if val.IsNil() {
		return reflect.Zero(val.Type())
	}

	// Copy each element, cloning a pointed-to struct.
	cloned := reflect.MakeSlice(val.Type(), val.Len(), val.Len())
	for i := range val.Len() {
		elem := accessible(val.Index(i))
		if elem.Kind() == reflect.Pointer && !elem.IsNil() && elem.Elem().Kind() == reflect.Struct {
			elemClone := reflect.New(elem.Elem().Type())
			cloneStruct(elemClone.Elem(), elem.Elem())
			setValue(cloned.Index(i), elemClone)
			continue
		}
		setValue(cloned.Index(i), elem)
	}
	return cloned
}

// setValue assigns src to dst when the types allow it.
func setValue(dst, src reflect.Value) {
	dst = accessible(dst)
	src = accessible(src)
	if dst.IsValid() && src.IsValid() && src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
	}
}

// accessible returns val made readable and writable when it is an unexported
// addressable field.
func accessible(val reflect.Value) reflect.Value {
	if val.CanInterface() || !val.CanAddr() {
		return val
	}
	return reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem()
}

// stackRelocation moves pointers from an old stack to its copy.
type stackRelocation struct {
	// oldBase is the address of the old stack.
	oldBase uintptr
	// oldEnd is the address just past the old stack.
	oldEnd uintptr
	// newBase is the address of the copy.
	newBase uintptr
}

// apply rebases every pointer into the old stack held by val, which is a
// uintptr, a uint64 or an array of them.
func (r stackRelocation) apply(val reflect.Value) {
	val = accessible(val)
	switch val.Kind() {
	case reflect.Array:
		for i := range val.Len() {
			r.apply(val.Index(i))
		}
	case reflect.Uint64, reflect.Uintptr:
		if ptr, ok := r.rebase(uintptr(val.Uint())); ok {
			val.SetUint(uint64(ptr))
		}
	}
}

// rebase returns ptr moved onto the copy, or false when ptr is outside the
// old stack.
func (r stackRelocation) rebase(ptr uintptr) (uintptr, bool) {
	if ptr < r.oldBase || ptr > r.oldEnd {
		return 0, false
	}
	return r.newBase + (ptr - r.oldBase), true
}
