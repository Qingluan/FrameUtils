package LocalDB

// func (dict Dict) String() string {
// 	buf, err := json.Marshal(dict)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return string(buf)
// }

func (index *Index) Add(bias Bias) {
	index.Include = append(index.Include, bias)
}
