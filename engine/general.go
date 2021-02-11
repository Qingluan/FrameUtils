package engine

type ArraysObj struct {
	arrays [][]interface{}
}

func (array ArraysObj) Close()

func FromArrays(doubleArray [][]interface{}) Obj {
	return &BaseObj{
		ArraysObj{
			arrays: doubleArray,
		},
	}
}
