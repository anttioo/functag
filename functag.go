package functag

import "reflect"

var funcs = map[uintptr]reflect.StructTag{}

func Tag(method reflect.Method) reflect.StructTag  {
	return funcs[method.Func.Pointer()]
}

func RegisterFunc(fn interface{}, tag string) reflect.StructTag  {
	v := reflect.ValueOf(fn).Pointer()
	funcs[v] = reflect.StructTag(tag)
	return reflect.StructTag(tag)
}