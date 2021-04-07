package textconvert

import (
	"strings"

	"github.com/timakin/gonvert"
)

func TOUTF8(raw string) (utf8Str string, err error) {
	converter := gonvert.New(raw, gonvert.UTF8)
	utf8Str, err = converter.Convert()
	// This will print out the utf-8 encoded string: "月日は百代の過客にして、行かふ年も又旅人也。"
	return
}
func (doc ElasticFileDocs) Size() int64 {
	return int64(len(doc.SomeStr))
}

func (doc ElasticFileDocs) MbSize() float64 {
	return float64(len(doc.SomeStr)) / float64(1024) / float64(1024)
}

func (doc ElasticFileDocs) SplitEsDoc(splitMinSizeMB int) (multi []ElasticFileDocs) {
	if doc.MbSize() > float64(splitMinSizeMB) {
		words := strings.Fields(doc.SomeStr)
		batch := []string{}
		batchSize := 0
		for _, word := range words {
			batchSize += len(word)
			batch = append(batch, word)
			if batchSize > splitMinSizeMB*1024*1024 {
				multi = append(multi, ElasticFileDocs{
					Path:    doc.Path,
					SomeStr: strings.Join(batch, " "),
				})
				batch = []string{}
				batchSize = 0
			}
		}
		if len(batch) > 0 {
			multi = append(multi, ElasticFileDocs{
				Path:    doc.Path,
				SomeStr: strings.Join(batch, " "),
			})

		}
	} else {
		multi = append(multi, doc)
	}
	return

}
