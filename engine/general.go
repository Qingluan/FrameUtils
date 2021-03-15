package engine

import "github.com/Qingluan/FrameUtils/utils"

type ArraysObj struct {
	EmptyBaseClass
	// Arrays []Line
}

func FromArrays(doubleArray [][]string, keys ...utils.Line) Obj {

	arrays := []utils.Line{}
	for _, v := range doubleArray {
		arrays = append(arrays, utils.Line(v))
	}
	k := utils.Line{}
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

func FromLines(arrays []utils.Line, keys ...utils.Line) Obj {

	k := utils.Line{}
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
