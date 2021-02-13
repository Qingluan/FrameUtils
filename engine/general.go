package engine

type ArraysObj struct {
	EmptyBaseClass
	// Arrays []Line
}

func FromArrays(doubleArray [][]string, keys ...Line) Obj {

	arrays := []Line{}
	for _, v := range doubleArray {
		arrays = append(arrays, Line(v))
	}
	k := Line{}
	if keys != nil {
		k = keys[0]
	}
	return &BaseObj{
		EmptyBaseClass{
			Arrays: arrays,
			Key:    k,
		},
	}
}

func FromLines(arrays []Line, keys ...Line) Obj {

	k := Line{}
	if keys != nil {
		k = keys[0]
	}
	return &BaseObj{
		EmptyBaseClass{
			Arrays: arrays,
			Key:    k,
		},
	}

}
