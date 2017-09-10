package util

type CompareFunc func(e1, e2 interface{}) bool

type List []interface{}

func NewList(capacity int) List {
	return make([]interface{}, 0, capacity)
}

func (list *List) Append(el interface{}) {
	*list = append(*list, el)
}

func (list *List) Clone() (List, int) {
	tmpList := NewList(list.Len())
	n := copy(tmpList[0:list.Len()], *list)
	tmpList = tmpList[0:list.Len()]
	return tmpList, n
}

func (list *List) Remove(el interface{}, fun CompareFunc) (index int){
	index = -1
	for i, element := range *list {
		if fun(element, el) {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	if index == len(*list) - 1 {
		*list = (*list)[0:index]
		return
	}
	backList := (*list)[index + 1 : len(*list)]
	*list = (*list)[0:index]
	*list = append(*list, backList...)
	return
}

func (list *List) Len() int {
	return len(*list)
}
