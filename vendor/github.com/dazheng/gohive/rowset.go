package gohive

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	inf "github.com/dazheng/gohive/inf"

	"git.apache.org/thrift.git/lib/go/thrift"
)

type rowSet struct {
	thrift    *inf.TCLIServiceClient
	operation *inf.TOperationHandle
	options   Options

	columns    []*inf.TColumnDesc
	columnStrs []string

	offset    int
	rowSet    *inf.TRowSet
	hasMore   bool
	ready     bool
	resultSet [][]interface{}
	nextRow   []interface{}
}

// A RowSet represents an asyncronous hive operation. You can
// Reattach to a previously submitted hive operation if you
// have a valid thrift client, and the serialized Handle()
// from the prior operation.
type RowSet interface {
	Handle(ctx context.Context) ([]byte, error)
	Columns() []string
	Next() bool
	Scan(dest ...interface{}) error
	Poll() (*Status, error)
	Wait() (*Status, error)
}

// Represents job status, including success state and time the
// status was updated.
type Status struct {
	state *inf.TOperationState
	Error error
	At    time.Time
}

func newRowSet(thrift *inf.TCLIServiceClient, operation *inf.TOperationHandle, options Options) RowSet {
	return &rowSet{thrift, operation, options, nil, nil, 0, nil, true, false, nil, nil}
}

// Issue a thrift call to check for the job's current status.
func (r *rowSet) Poll() (*Status, error) {
	req := inf.NewTGetOperationStatusReq()
	req.OperationHandle = r.operation

	resp, err := r.thrift.GetOperationStatus(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("Error getting status: %+v, %v", resp, err)
	}

	if !isSuccessStatus(resp.Status) {
		return nil, fmt.Errorf("GetStatus call failed: %s", resp.Status.String())
	}

	if resp.OperationState == nil {
		return nil, errors.New("No error from GetStatus, but nil status!")
	}

	return &Status{resp.OperationState, nil, time.Now()}, nil
}

// Wait until the job is complete, one way or another, returning Status and error.
func (r *rowSet) Wait() (*Status, error) {
	for {
		status, err := r.Poll()

		if err != nil {
			return nil, err
		}

		if status.IsComplete() {
			if status.IsSuccess() {
				// Fetch operation metadata.
				metadataReq := inf.NewTGetResultSetMetadataReq()
				metadataReq.OperationHandle = r.operation

				metadataResp, err := r.thrift.GetResultSetMetadata(context.Background(), metadataReq)
				if err != nil {
					return nil, err
				}

				if !isSuccessStatus(metadataResp.Status) {
					return nil, fmt.Errorf("GetResultSetMetadata failed: %s", metadataResp.Status.String())
				}

				r.columns = metadataResp.Schema.Columns
				r.ready = true

				return status, nil
			}
			return nil, fmt.Errorf("Query failed execution: %s", status.state.String())
		}

		time.Sleep(time.Duration(r.options.PollIntervalSeconds) * time.Second)
	}
}

func (r *rowSet) waitForSuccess() error {
	if !r.ready {
		status, err := r.Wait()
		if err != nil {
			return err
		}
		if !status.IsSuccess() || !r.ready {
			return fmt.Errorf("Unsuccessful query execution: %+v", status)
		}
	}

	return nil
}

func (r *rowSet) fetchAll() bool {
	if !r.hasMore {
		return false
	}

	fetchReq := inf.NewTFetchResultsReq()
	fetchReq.OperationHandle = r.operation
	fetchReq.Orientation = inf.TFetchOrientation_FETCH_NEXT
	fetchReq.MaxRows = r.options.BatchSize

	resp, err := r.thrift.FetchResults(context.Background(), fetchReq)
	if err != nil {
		log.Printf("FetchResults failed: %v\n", err)
		return false
	}

	if !isSuccessStatus(resp.Status) {
		log.Printf("FetchResults failed: %s\n", resp.Status.String())
		return false
	}

	r.offset = 0
	r.rowSet = resp.GetResults()
	r.hasMore = *resp.HasMoreRows

	rs := r.rowSet.Columns
	colLen := len(rs)
	r.resultSet = make([][]interface{}, colLen)

	// 先列后行
	for i := 0; i < colLen; i++ {
		v, length := convertColumn(rs[i])
		c := make([]interface{}, length)
		for j := 0; j < length; j++ {
			c[j] = reflect.ValueOf(v).Index(j).Interface()
		}
		r.resultSet[i] = c
	}

	return true

}

// Prepares a row for scanning into memory, by reading data from hive if
// the operation is successful, blocking until the operation is
// complete, if necessary.
// Returns true is a row is available to Scan(), and false if the
// results are empty or any other error occurs.
func (r *rowSet) Next() bool {
	if err := r.waitForSuccess(); err != nil {
		return false
	}

	if r.resultSet == nil {
		r.fetchAll()
		r.offset = 0
		//		fmt.Println(r.resultSet)
	}

	if len(r.resultSet) <= 0 {
		return false
	}
	if r.offset >= len(r.resultSet[0]) {
		return false
	}
	r.nextRow = make([]interface{}, 0)
	for _, v := range r.resultSet {
		r.nextRow = append(r.nextRow, v[r.offset])
	}
	//	fmt.Println(r.nextRow)
	r.offset++
	return true
}

// Scan the last row prepared via Next() into the destination(s) provided,
// which must be pointers to value types, as in database.sql. Further,
// only pointers of the following types are supported:
// 	- int, int16, int32, int64
// 	- string, []byte
// 	- float64
//	 - bool
func (r *rowSet) Scan(dest ...interface{}) error {
	// TODO: Add type checking and conversion between compatible
	// types where possible, as well as some common error checking,
	// like passing nil. database/sql's method is very convenient,
	// for example: http://golang.org/src/pkg/database/sql/convert.go, like 85
	if r.nextRow == nil {
		return errors.New("No row to scan! Did you call Next() first?")
	}

	if len(dest) != len(r.nextRow) {
		return fmt.Errorf("Can't scan into %d arguments with input of length %d", len(dest), len(r.nextRow))
	}

	for i, val := range r.nextRow {
		d := dest[i]
		switch dt := d.(type) {
		case *string:
			switch st := val.(type) {
			case string:
				*dt = st
			default:
				*dt = fmt.Sprintf("%v", val)
			}
		case *[]byte:
			*dt = []byte(val.(string))
		case *int:
			*dt = int(val.(int32))
		case *int64:
			*dt = val.(int64)
		case *int32:
			*dt = val.(int32)
		case *int16:
			*dt = val.(int16)
		case *float64:
			*dt = val.(float64)
		case *bool:
			*dt = val.(bool)
		default:
			return fmt.Errorf("Can't scan value of type %T with value %v", dt, val)
		}
	}

	return nil
}

// Returns the names of the columns for the given operation,
// blocking if necessary until the information is available.
func (r *rowSet) Columns() []string {
	if r.columnStrs == nil {
		if err := r.waitForSuccess(); err != nil {
			return nil
		}

		ret := make([]string, len(r.columns))
		for i, col := range r.columns {
			ret[i] = col.ColumnName
		}

		r.columnStrs = ret
	}

	return r.columnStrs
}

// Return a serialized representation of an identifier that can later
// be used to reattach to a running operation. This identifier and
// serialized representation should be considered opaque by users.
func (r *rowSet) Handle(ctx context.Context) ([]byte, error) {
	return serializeOp(ctx, r.operation)
}

func convertColumn(col *inf.TColumn) (colValues interface{}, length int) {
	switch {
	case col.IsSetStringVal():
		return col.GetStringVal().GetValues(), len(col.GetStringVal().GetValues())
	case col.IsSetBoolVal():
		return col.GetBoolVal().GetValues(), len(col.GetBoolVal().GetValues())
	case col.IsSetByteVal():
		return col.GetByteVal().GetValues(), len(col.GetByteVal().GetValues())
	case col.IsSetI16Val():
		return col.GetI16Val().GetValues(), len(col.GetI16Val().GetValues())
	case col.IsSetI32Val():
		return col.GetI32Val().GetValues(), len(col.GetI32Val().GetValues())
	case col.IsSetI64Val():
		return col.GetI64Val().GetValues(), len(col.GetI64Val().GetValues())
	case col.IsSetDoubleVal():
		return col.GetDoubleVal().GetValues(), len(col.GetDoubleVal().GetValues())
	default:
		return nil, 0
	}
}

// Returns a string representation of operation status.
func (s Status) String() string {
	if s.state == nil {
		return "unknown"
	}
	return s.state.String()
}

// Returns true if the job has completed or failed.
func (s Status) IsComplete() bool {
	if s.state == nil {
		return false
	}

	switch *s.state {
	case inf.TOperationState_FINISHED_STATE,
		inf.TOperationState_CANCELED_STATE,
		inf.TOperationState_CLOSED_STATE,
		inf.TOperationState_ERROR_STATE:
		return true
	}

	return false
}

// Returns true if the job compelted successfully.
func (s Status) IsSuccess() bool {
	if s.state == nil {
		return false
	}

	return *s.state == inf.TOperationState_FINISHED_STATE
}

func deserializeOp(handle []byte) (*inf.TOperationHandle, error) {
	ser := thrift.NewTDeserializer()
	var val inf.TOperationHandle

	if err := ser.Read(&val, handle); err != nil {
		return nil, err
	}

	return &val, nil
}

func serializeOp(ctx context.Context, operation *inf.TOperationHandle) ([]byte, error) {
	ser := thrift.NewTSerializer()
	return ser.Write(ctx, operation)
}
