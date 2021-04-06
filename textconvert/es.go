package textconvert

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/Qingluan/FrameUtils/utils"

	"github.com/olivere/elastic"
	// Import the Elasticsearch library packages
	// "github.com/elastic/go-elasticsearch/v8"
	// "github.com/elastic/go-elasticsearch/v8/esapi"
	// "github.com/elastic/go-elasticsearch/v8/esutil"
)

// Declare a struct for Elasticsearch fields
type ElasticFileDocs struct {
	Path    string `json:"path"`
	SomeStr string `json:"body"`
	SomeInt int    `json:"number"`
}

type EsClient struct {
	Context  context.Context
	client   *elastic.Client
	TmpCache []ElasticFileDocs
}

// A function for marshaling structs to JSON string
func (doc ElasticFileDocs) Json() string {

	// Create struct instance of the Elasticsearch fields struct object

	// Marshal the struct to JSON and check for errors
	b, err := json.Marshal(doc)
	if err != nil {
		fmt.Println("json.Marshal ERROR:", err)
		return fmt.Sprintf("{\"Err\":\"%s\"}", err.Error())
	}
	return string(b)
}

func NewEsCli(name, pwd string, address ...string) (es *EsClient, err error) {
	es = new(EsClient)

	// Allow for custom formatting of log output
	log.SetFlags(0)

	// Create a context object for the API calls
	es.Context = context.Background()

	// Create a mapping for the Elasticsearch documents
	var (
		docMap map[string]interface{}
	)
	fmt.Println("docMap:", docMap)
	fmt.Println("docMap TYPE:", reflect.TypeOf(docMap))
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	//
	// Use a third-party package for implementing the backoff function
	//
	// retryBackoff := backoff.NewExponentialBackOff()
	// Declare an Elasticsearch configuration

	// Instantiate a new Elasticsearch client object instance
	es.client, err = elastic.NewClient(
		elastic.SetURL(address...),
		elastic.SetSniff(false),
		// elastic.SetRetrier(elastic.NewRetr()),
		elastic.SetGzip(true),
		elastic.SetHealthcheckInterval(10*time.Second),
	)

	if err != nil {
		fmt.Println("Elasticsearch connection error:", err)
	}

	// Have the client instance return a response
	// if res, err := es.client.Info(); err != nil {
	// log.Fatalf("client.Info() ERROR:%s", err)
	// return nil, err
	// } else {
	// log.Printf("client response:%s", res)
	// }

	return
}

func (es *EsClient) BatchingThenImport(index string, doc ElasticFileDocs, num int) (success, failed uint64) {
	es.TmpCache = append(es.TmpCache, doc)
	if len(es.TmpCache) > num {
		success, failed = es.BatchImport(index, es.TmpCache...)
		es.TmpCache = []ElasticFileDocs{}
		return
	}
	return 0, 0
}

func (es *EsClient) Wait(index string) (success, failed uint64) {
	if len(es.TmpCache) > 0 {
		success, failed = es.BatchImport(index, es.TmpCache...)
		es.TmpCache = []ElasticFileDocs{}

		return

	}
	return 0, 0
}

func (es *EsClient) BatchImport(index string, esobjs ...ElasticFileDocs) (success, failed uint64) {
	start := time.Now()
	defer log.Println(utils.BGreen(fmt.Sprintf("Used %s All: %d Success : %d ", time.Now().Sub(start), len(esobjs), success)) + utils.BRed(fmt.Sprintf("| Failed : %d ", failed)))
	// indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
	// 	Index:         index,            // The default index name
	// 	Client:        es.client,        // The Elasticsearch client
	// 	NumWorkers:    runtime.NumCPU(), // The number of worker goroutines
	// 	FlushBytes:    int(5e+6),        // The flush threshold in bytes
	// 	FlushInterval: 10 * time.Second, // The periodic flush interval
	// })
	// bulkReq := es.client.Bulk()
	bulk := es.client.Bulk()
	// Setup a bulk processor
	// bulk, err := es.client.BulkProcessor().Name("MyBackgroundWorker-1").
	// 	Workers(runtime.NumCPU()).
	// 	BulkActions(2000).               // commit if # requests >= 1000
	// 	BulkSize(2 << 20).               // commit if size of requests >= 2 MB
	// 	FlushInterval(10 * time.Second). // commit every 30s
	// 	Do(context.Background())
	// if err != nil {

	// }
	now, _ := es.client.Count(index).Do(context.Background())

	for i, a := range esobjs {
		req := elastic.NewBulkIndexRequest().Index(index).Id(fmt.Sprintf("%d", now+int64(i))).Type("es-file").Doc(a)
		bulk = bulk.Add(req)
		num := bulk.EstimatedSizeInBytes()

		if float64(num)/float64(1024)/float64(1024) > 10 {
			log.Println("Batch Size: ", float64(num)/float64(1024)/float64(1024), "MB")

			bulk.Do(context.Background())
			bulk = es.client.Bulk()
		}
		// bulk.Add(req)
		// data, err := json.Marshal(a)
		// if err != nil {
		// 	log.Printf("Cannot encode article %s: %s", a.Path, err.Error())
		// }
		// err = indexer.Add(context.Background(), esutil.BulkIndexerItem{
		// 	Action:     "index",
		// 	Index:      index,
		// 	DocumentID: strconv.Itoa(i),
		// 	Body: bytes.NewReader(data),
		// 	OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
		// 		atomic.AddUint64(&success, 1)
		// 	},
		// 	OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
		// 		if err != nil {
		// 			log.Printf("ERROR: %s", err)
		// 		} else {
		// 			log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
		// 		}
		// 		atomic.AddUint64(&failed, 1)
		// 	},
		// },
		// )
		// if err != nil {
		// 	log.Fatalf("Unexpected error: %s", err)
		// }
	}
	// num := bulk.EstimatedSizeInBytes()
	// indexer.Close(context.Background())

	// log.Println(utils.Yellow(bulkReq.NumberOfActions()))
	// time.Sleep(5 * time.Second)

	if res, err := bulk.Do(context.Background()); err != nil {
		failed = uint64(len(esobjs))
		log.Println("err:", err)
	} else {
		success = uint64(len(res.Succeeded()))
		// log.Println("Res:", res.Failed(), res.Created())
	}
	// success = uint64(bulk.Stats().Committed)
	return
}
