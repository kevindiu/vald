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
	insertCont = 400
	testCount  = 20
)

var (
	datasetPath         string
	grpcServerAddr      string
	indexingWaitSeconds uint
)

func init() {
	/**
	Registers path, addr and wait option.
	Path option specifies hdf file by path. By default, `fashion-mnist-784-euclidean.hdf5` is registered.
	Addr option specifies grpc server address. By default, `127.0.0.1:8080` is registered.
	Wait option specifies indexing wait time. Vald starts indexing automatically after insert. Therefore, it needs to wait until indexing is completed before searching.By default `60` seconds is registered.
	**/
	flag.StringVar(&datasetPath, "path", "fashion-mnist-784-euclidean.hdf5", "set dataset path")
	flag.StringVar(&grpcServerAddr, "addr", "127.0.0.1:8081", "set gRPC server address")
	flag.UintVar(&indexingWaitSeconds, "wait", 60, "set indexing wait seconds")
	flag.Parse()
}

func main() {
	/**
	Gets training data, test data and ids based on the dataset path.
	the number of ids is equal to that of training dataset.
	**/
	ids, train, test, err := load(datasetPath)
	if err != nil {
		glg.Fatal(err)
	}

	ctx := context.Background()

	/**
	Creates a client connection to the given the target.
	Then, creates a Vald client based on this connection.
	**/
	conn, err := grpc.DialContext(ctx, grpcServerAddr, grpc.WithInsecure())
	if err != nil {
		glg.Fatal(err)
	}
	// Creates Vald client for gRPC.
	client := vald.NewValdClient(conn)

	glg.Infof("Start Inserting %d Vector", insertCont)

	/**
	Starts inserting vectors specified by insertCount(400).
	**/
	for i := range ids[:insertCont] {
		if i%10 == 0 {
			glg.Infof("Inserted: %d", i)
		}
		// Calls `Insert` function of Vald client.
		// Sends set of vector and id to server via gRPC.
		_, err := client.Insert(ctx, &payload.Object_Vector{
			Id:     ids[i],
			Vector: train[i],
		})
		if err != nil {
			glg.Fatal(err)
		}
	}

	glg.Info("Finish Inserting. \n\n")
	glg.Info("Wait for indexing to finish")
	time.Sleep(time.Duration(indexingWaitSeconds) * time.Second)

	glg.Infof("Start search %d times", testCount)

	/**
	Gets approximate vectors, which is based on the value of `SearchConfig`, from the indexed tree based on the training data.
	In this example, Vald gets 10 approximate vectors each search vector.
	**/
	for i, vec := range test[:testCount] {
		// Calls `Search` function of Vald client.
		// Sends vector and configuration object to server via gRPC.
		res, err := client.Search(ctx, &payload.Search_Request{
			Vector: vec,
			// Conditions for hitting the search.
			Config: &payload.Search_Config{
				// the number of search results.
				Num: 10,
				// Radius is used to determine the space of search candidate radius for neighborhood vectors.
				// Defaults to -1. That is, infinite circle.
				Radius: -1,
				// Epsilon is the parameter that determines how much to expand from search candidate radius.
				Epsilon: 0.01,
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
		// The returned dataset must be closed by the user when it is no longer needed.
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