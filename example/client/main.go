package main

import (
	"context"
	"encoding/json"
	"flag"
	"time"

	"github.com/kpango/fuid"
	"github.com/kpango/glg"
	"github.com/vdaas/vald-client-go/gateway/vald"
	"github.com/vdaas/vald-client-go/payload"

	"gonum.org/v1/hdf5"
	"google.golang.org/grpc"
)

const (
	insertCount = 400
	testCount   = 20
)

var (
	datasetPath         string
	grpcServerAddr      string
	indexingWaitSeconds uint
)

func init() {
	/**
	Path option specifies hdf file by path. Default value is `fashion-mnist-784-euclidean.hdf5`.
	Addr option specifies grpc server address. Default value is `127.0.0.1:8080`.
	Wait option specifies indexing wait time (in seconds). Default value is  `60`.
	**/
	flag.StringVar(&datasetPath, "path", "fashion-mnist-784-euclidean.hdf5", "dataset path")
	flag.StringVar(&grpcServerAddr, "addr", "127.0.0.1:8081", "gRPC server address")
	flag.UintVar(&indexingWaitSeconds, "wait", 60, "indexing wait seconds")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	// Create a Vald client for connecting to the Vald cluster.
	conn, err := grpc.DialContext(ctx, grpcServerAddr, grpc.WithInsecure())
	if err != nil {
		glg.Fatal(err)
	}
	client := vald.NewValdClient(conn)

	// Gets training data, test data and ids based on the dataset path.
	glg.Infof("Start Inserting %d testing Vector to Vald", insertCount)
	ids, train, test, err := load(datasetPath)
	if err != nil {
		glg.Fatal(err)
	}

	// Insert 400 example vectors into Vald cluster.
	for i := range ids[:insertCount] {
		if i%10 == 0 {
			glg.Infof("Inserted: %d", i)
		}
		_, err := client.Insert(ctx, &payload.Object_Vector{
			Id:     ids[i],
			Vector: train[i],
		})
		if err != nil {
			glg.Fatal(err)
		}
	}
	glg.Info("Finish Inserting dataset. \n\n")

	// Vald starts indexing automatically after insert. It needs to wait until the indexing is completed before a search action is performed.
	glg.Info("Wait for indexing to finish")
	time.Sleep(time.Duration(indexingWaitSeconds) * time.Second)

	// Gets approximate vectors, which is based on the value of `SearchConfig`, from the indexed tree based on the training data.
	// In this example, Vald gets 10 approximate vectors each search vector.
	glg.Infof("Start searching %d times", testCount)
	for i, vec := range test[:testCount] {
		// Send searching vector and configuration object to the Vald server via gRPC.
		res, err := client.Search(ctx, &payload.Search_Request{
			Vector: vec,
			Config: &payload.Search_Config{
				Num:     10,   // the number of search results.
				Radius:  -1,   // Radius is used to determine the space of search candidate radius for neighborhood vectors. -1 means infinite circle.
				Epsilon: 0.01, // Epsilon is used to determines how much to expand from search candidate radius.
			},
		})
		if err != nil {
			glg.Fatal(err)
		}

		b, _ := json.MarshalIndent(res.GetResults(), "", " ")
		glg.Infof("%d - Results : %s\n\n", i+1, string(b))
		time.Sleep(1 * time.Second)
	}
}

// load function loads training and test vector from hdf file. The size of ids is same to the number of training data.
// Each id, which is an element of ids, will be set a random number.
func load(path string) (ids []string, train, test [][]float32, err error) {
	var f *hdf5.File
	f, err = hdf5.OpenFile(path, hdf5.F_ACC_RDONLY)
	if err != nil {
		return nil, nil, nil, err
	}
	defer f.Close()

	// readFn function reads vectors of the hierarchy with the given the name.
	readFn := func(name string) ([][]float32, error) {
		// Opens and returns a named Dataset.
		d, err := f.OpenDataset(name)
		if err != nil {
			return nil, err
		}
		defer d.Close()

		// Space returns an identifier for a copy of the dataspace for a dataset.
		sp := d.Space()
		defer sp.Close()

		// SimpleExtentDims returns dataspace dimension size and maximum size.
		dims, _, _ := sp.SimpleExtentDims()
		row, dim := int(dims[0]), int(dims[1])

		// Gets the stored vector. All are represented as one-dimensional arrays.
		vec := make([]float64, sp.SimpleExtentNPoints())
		if err := d.Read(&vec); err != nil {
			return nil, err
		}

		// Converts a one-dimensional array to a two-dimensional array.
		// Use the `dim` variable as a separator.
		vecs := make([][]float32, row)
		for i := 0; i < row; i++ {
			vecs[i] = make([]float32, dim)
			for j := 0; j < dim; j++ {
				vecs[i][j] = float32(vec[i*dim+j])
			}
		}

		return vecs, nil
	}

	// Gets vector of `train` hierarchy.
	train, err = readFn("train")
	if err != nil {
		return nil, nil, nil, err
	}

	// Gets vector of `test` hierarchy.
	test, err = readFn("test")
	if err != nil {
		return nil, nil, nil, err
	}

	// Generate as many random ids for training vectors.
	ids = make([]string, 0, len(train))
	for i := 0; i < len(train); i++ {
		ids = append(ids, fuid.String())
	}

	return
}
