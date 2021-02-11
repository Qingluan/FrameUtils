package engine

type ArraysObj struct {
	EmptyBaseClass
	// Arrays []Line
}

func FromArrays(doubleArray [][]string) Obj {

	arrays := []Line{}
	for _, v := range doubleArray {
		arrays = append(arrays, Line(v))
	}
	return &BaseObj{
		EmptyBaseClass{
			Arrays: arrays,
		},
	}

}
