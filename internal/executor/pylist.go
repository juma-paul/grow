package executor

import (
	"github.com/go-python/gpython/py"
	"github.com/juma-paul/grow/internal/events"
	"github.com/juma-paul/grow/internal/simulator"
)

// PyVisualListType is the gpython type for VisualList.
var PyVisualListType = py.NewType("VisualList", "A list that emits resize/copy events for visualization.")

func init() {
	PyVisualListType.Dict["append"] = py.MustNewMethod("append", pyListAppend, 0, "Append an item.")
	PyVisualListType.Dict["pop"] = py.MustNewMethod("pop", pyListPop, 0, "Remove and return the last item.")
	PyVisualListType.Dict["insert"] = py.MustNewMethod("insert", pyListInsert, 0, "Insert an item at index.")
	PyVisualListType.Dict["extend"] = py.MustNewMethod("extend", pyListExtend, 0, "Extend with items from an iterable.")
}

// PyVisualList wraps simulator.VisualList for gpython.
type PyVisualList struct {
	inner  *simulator.VisualList
	items  []py.Object
	Events []events.Event
}

func (p *PyVisualList) Type() *py.Type {
	return PyVisualListType
}

// NewPyVisualList creates a VisualList that collects events.
func NewPyVisualList() *PyVisualList {
	pvl := &PyVisualList{}
	pvl.inner = simulator.NewVisualList(simulator.CPythonGrowth{}, func(e events.Event) {
		pvl.Events = append(pvl.Events, e)
	})
	return pvl
}

// NewVisualListFactory returns a callable that constructs PyVisualList
// instances. All instances share the same event emitter, so events from
// multiple lists land in one stream.
func NewVisualListFactory(emit func(events.Event)) *py.Method {
	return py.MustNewMethod("VisualList", func(self py.Object, args py.Tuple) (py.Object, error) {
		pvl := &PyVisualList{}
		pvl.inner = simulator.NewVisualListWithLimits(simulator.CPythonGrowth{}, emit, simulator.Limits{
			MaxOps:   100_000,
			MaxAlloc: 10_000_000,
		})
		if len(args) == 0 {
			return pvl, nil
		}
		iter, err := py.Iter(args[0])
		if err != nil {
			return nil, err
		}
		for {
			item, err := py.Next(iter)
			if err != nil {
				if py.IsException(py.StopIteration, err) {
					break
				}
				return nil, err
			}
			pvl.inner.Append(item)
			pvl.items = append(pvl.items, item)
		}
		return pvl, nil
	}, 0, "VisualList([iterable]) -- create a visualized list")
}

// M__len__ implements len(lst).
func (p *PyVisualList) M__len__() (py.Object, error) {
	return py.Int(p.inner.Len()), nil
}

// M__getitem__ implements lst[i].
func (p *PyVisualList) M__getitem__(key py.Object) (py.Object, error) {
	idx, err := py.IndexIntCheck(key, p.inner.Len())
	if err != nil {
		return nil, err
	}
	return p.items[idx], nil
}

// M__setitem__ implements lst[i] = val.
func (p *PyVisualList) M__setitem__(key, value py.Object) (py.Object, error) {
	idx, err := py.IndexIntCheck(key, p.inner.Len())
	if err != nil {
		return nil, err
	}
	p.items[idx] = value
	return py.None, nil
}

// M__iter__ implements for x in lst.
func (p *PyVisualList) M__iter__() (py.Object, error) {
	items := make(py.Tuple, p.inner.Len())
	copy(items, p.items[:p.inner.Len()])
	return py.NewIterator(items), nil
}

func pyListAppend(self py.Object, args py.Tuple) (py.Object, error) {
	p := self.(*PyVisualList)
	if len(args) != 1 {
		return nil, py.ExceptionNewf(py.TypeError, "append() takes exactly one argument (%d given)", len(args))
	}
	p.inner.Append(args[0])
	p.items = append(p.items, args[0])
	return py.None, nil
}

func pyListPop(self py.Object, args py.Tuple) (py.Object, error) {
	p := self.(*PyVisualList)
	if p.inner.Len() == 0 {
		return nil, py.ExceptionNewf(py.IndexError, "pop from empty list")
	}
	p.inner.Pop()
	last := p.items[len(p.items)-1]
	p.items = p.items[:len(p.items)-1]
	return last, nil
}

func pyListInsert(self py.Object, args py.Tuple) (py.Object, error) {
	p := self.(*PyVisualList)
	if len(args) != 2 {
		return nil, py.ExceptionNewf(py.TypeError, "insert() takes exactly 2 arguments (%d given)", len(args))
	}
	idx, err := py.GetInt(args[0])
	if err != nil {
		return nil, err
	}
	p.inner.Insert(int(idx), args[1])
	// Insert into our items slice at the same position
	p.items = append(p.items, nil)
	copy(p.items[idx+1:], p.items[idx:])
	p.items[idx] = args[1]
	return py.None, nil
}

func pyListExtend(self py.Object, args py.Tuple) (py.Object, error) {
	p := self.(*PyVisualList)
	if len(args) != 1 {
		return nil, py.ExceptionNewf(py.TypeError, "extend() takes exactly one argument (%d given)", len(args))
	}
	iter, err := py.Iter(args[0])
	if err != nil {
		return nil, err
	}
	var goItems []any
	var pyItems []py.Object
	for {
		item, err := py.Next(iter)
		if err != nil {
			if py.IsException(py.StopIteration, err) {
				break
			}
			return nil, err
		}
		goItems = append(goItems, item)
		pyItems = append(pyItems, item)
	}
	p.inner.Extend(goItems)
	p.items = append(p.items, pyItems...)
	return py.None, nil
}
