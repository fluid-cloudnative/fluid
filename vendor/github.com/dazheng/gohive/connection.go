// Package hivething wraps the hiveserver2 thrift interface in a few
// related interfaces for more convenient use.
package gohive

import (
	"errors"
	"fmt"

	inf "github.com/dazheng/gohive/inf"

	"context"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// Options for opened Hive sessions.
type Options struct {
	PollIntervalSeconds int64
	BatchSize           int64
}

var (
	DefaultOptions = Options{PollIntervalSeconds: 5, BatchSize: 10000}
)

type Connection struct {
	thrift  *inf.TCLIServiceClient
	session *inf.TSessionHandle
	options Options
}

func Connect(host string, options Options) (*Connection, error) {
	transport, err := thrift.NewTSocket(host)
	if err != nil {
		return nil, err
	}

	if err := transport.Open(); err != nil {
		return nil, err
	}

	if transport == nil {
		return nil, errors.New("nil thrift transport")
	}

	/*
		NB: hive 0.13's default is a TSaslProtocol, but
		there isn't a golang implementation in apache thrift as
		of this writing.
	*/
	protocol := thrift.NewTBinaryProtocolFactoryDefault()
	client := inf.NewTCLIServiceClientFactory(transport, protocol)
	s := inf.NewTOpenSessionReq()
	s.ClientProtocol = 6
	session, err := client.OpenSession(context.Background(), s)
	if err != nil {
		return nil, err
	}

	return &Connection{client, session.SessionHandle, options}, nil
}

func ConnectWithUser(host, username, password string, options Options) (*Connection, error) {
	transport, err := thrift.NewTSocket(host)
	if err != nil {
		return nil, err
	}

	if err := transport.Open(); err != nil {
		return nil, err
	}

	if transport == nil {
		return nil, errors.New("nil thrift transport")
	}

	/*
		NB: hive 0.13's default is a TSaslProtocol, but
		there isn't a golang implementation in apache thrift as
		of this writing.
	*/
	protocol := thrift.NewTBinaryProtocolFactoryDefault()
	client := inf.NewTCLIServiceClientFactory(transport, protocol)
	s := inf.NewTOpenSessionReq()
	s.ClientProtocol = 6
	s.Username = &username
	s.Password = &password
	session, err := client.OpenSession(context.Background(), s)
	if err != nil {
		return nil, err
	}

	return &Connection{client, session.SessionHandle, options}, nil
}

func (c *Connection) isOpen() bool {
	return c.session != nil
}

// Closes an open hive session. After using this, the
// connection is invalid for other use.
func (c *Connection) Close() error {
	if c.isOpen() {
		closeReq := inf.NewTCloseSessionReq()
		closeReq.SessionHandle = c.session
		resp, err := c.thrift.CloseSession(context.Background(), closeReq)
		if err != nil {
			return fmt.Errorf("Error closing session: ", resp, err)
		}

		c.session = nil
	}

	return nil
}

// Issue a query on an open connection, returning a RowSet, which
// can be later used to query the operation's status.
func (c *Connection) Query(query string) (RowSet, error) {
	executeReq := inf.NewTExecuteStatementReq()
	executeReq.SessionHandle = c.session
	executeReq.Statement = query

	resp, err := c.thrift.ExecuteStatement(context.Background(), executeReq)
	if err != nil {
		return nil, fmt.Errorf("Error in ExecuteStatement: %+v, %v", resp, err)
	}

	if !isSuccessStatus(resp.Status) {
		return nil, fmt.Errorf("Error from server: %s", resp.Status.String())
	}

	return newRowSet(c.thrift, resp.OperationHandle, c.options), nil
}

func (c *Connection) Exec(query string) (*inf.TExecuteStatementResp, error) {
	executeReq := inf.NewTExecuteStatementReq()
	executeReq.SessionHandle = c.session
	executeReq.Statement = query

	resp, err := c.thrift.ExecuteStatement(context.Background(), executeReq)
	if err != nil {
		return nil, fmt.Errorf("Error in ExecuteStatement: %+v, %v", resp, err)
	}

	if !isSuccessStatus(resp.Status) {
		return nil, fmt.Errorf("Error from server: %s", resp.Status.String())
	}

	return resp, err
}

func isSuccessStatus(p *inf.TStatus) bool {
	status := p.GetStatusCode()
	return status == inf.TStatusCode_SUCCESS_STATUS || status == inf.TStatusCode_SUCCESS_WITH_INFO_STATUS
}
