package LocalDB

import "io"

type ChangePoint struct {
	Oldbias Bias
	Newbias Bias
	Tp      int
	// ChangeKeys   map[string]int
	// ItemPosition int
	Buf  []byte
	Next *ChangePoint
	Pre  *ChangePoint
}

func (ch *ChangePoint) Change(old Bias, buf string) *ChangePoint {
	cho := &ChangePoint{
		Oldbias: old,
		Newbias: Bias{old[0], old[1]},
		Buf:     []byte(buf),
		// ChangeKeys:   changeKeys,
		// ItemPosition: itemPosition,
	}
	if len(buf) != int(old[1]-old[0]) {
		cho.Newbias[1] += int64(len(buf)) - (old[1] - old[0])
	}
	return ch.add(cho)
}

func (ch *ChangePoint) Delete(bias Bias) *ChangePoint {
	cho := &ChangePoint{
		Oldbias: bias,
	}
	return ch.add(cho)
}

func (ch *ChangePoint) add(newch *ChangePoint) *ChangePoint {
	if ch.LessThan(newch) {
		if ch.Next != nil {
			if ch.Next.MoreThan(newch) {
				ch.Next.Pre = newch
				newch.Next = ch.Next
				ch.Next = newch
				newch.Pre = ch
			} else {
				ch.Next.add(newch)
			}

		} else {
			ch.Next = newch
			newch.Pre = ch
		}
		return ch
	} else {
		newch.add(ch)
		return newch
	}
}

func (ch *ChangePoint) LessThan(och *ChangePoint) bool {
	if ch.Oldbias[1] <= och.Oldbias[0] {
		return true
	}
	return false
}

func (ch *ChangePoint) MoreThan(och *ChangePoint) bool {
	if ch.Oldbias[0] >= och.Oldbias[1] {
		return true
	}
	return false
}

func (ch *ChangePoint) ChangedOffset() int64 {
	return ch.Newbias[1] - ch.Oldbias[1]
}

func (ch *ChangePoint) propagate(dst io.Writer, src io.ReadSeeker, start int64) (newstart int64) {
	src.Seek(start, io.SeekStart)
	length := ch.Oldbias[0] - start
	io.CopyN(dst, src, length)

	switch ch.Tp {
	case -1:
	default:
		if _, err := dst.Write(ch.Buf); err != nil {
			panic(err)
		}
	}

	newstart = ch.Oldbias[1]
	// newoffset = offset + ch.ChangedOffset()
	return
}

//  --------------- | x | -----------------------| x2 | -----------------
//  --------------- | ny | ------------------| nx2 | -----------------
func (ch *ChangePoint) First() *ChangePoint {
	if ch.Pre != nil {
		return ch.First()
	} else {
		return ch
	}
}

func (ch *ChangePoint) Last() *ChangePoint {
	if ch.Next != nil {
		return ch.Next.Last()
	} else {
		return ch
	}
}
