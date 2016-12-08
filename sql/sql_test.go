package cassandra

import (
	"database/sql"
	"errors"
	"farm.e-pedion.com/repo/logger"
	"farm.e-pedion.com/repo/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func init() {
	if err := logger.Setup(&logger.Configuration{}); err != nil {
		panic(err)
	}
}

func TestUnitDelegateDB(t *testing.T) {
	assert.NotPanics(t,
		func() {
			db := NewDelegateDB(new(sql.DB))
			assert.NotNil(t, db)
		},
	)
}

func TestUnitDelegateRow(t *testing.T) {
	assert.NotPanics(t,
		func() {
			row := NewDelegateRow(new(sql.Row))
			assert.NotNil(t, row)

			var err error
			assert.Panics(t,
				func() {
					var out1 string
					var out2 int
					err = row.Scan(&out1, &out2)
				},
			)
			assert.Nil(t, err)
			assert.Panics(t,
				func() {
					row.Scan(nil)
				},
			)
		},
	)
}

func TestUnitClient(t *testing.T) {
	assert.NotPanics(t,
		func() {
			client := NewClient(NewDBMock())
			assert.Nil(t, client.Close())
		},
	)
}

func TestUnitClientPool(t *testing.T) {
	dbMock := NewDBMock()
	dbMock.On("Ping").Return(nil)
	dbMock.On("Close").Return(nil)
	setupPool := &Pool{
		datasource: new(Datasource),
		db:         dbMock,
	}
	Config = &Configuration{}
	assert.Nil(t, persistence.Setup(setupPool))
	pool, err := persistence.GetPool()
	assert.Nil(t, err)
	assert.NotNil(t, pool)
	client, err := pool.Get()
	assert.Nil(t, err)
	assert.NotNil(t, client)
	err = client.Close()
	assert.Nil(t, err)
	err = pool.Close()
	assert.Nil(t, err)
}

func TestUnitQueryExec(t *testing.T) {
	rowMock := NewRowMock()
	rowMock.On("Scan", mock.Anything).Return(nil)
	dbMock := NewDBMock()
	dbMock.On("Close").Return(nil)
	dbMock.On("QueryRow", mock.Anything, mock.Anything).Return(rowMock)
	persistenceClient := NewClient(dbMock)
	assert.NotNil(t, persistenceClient)
	err := persistenceClient.QueryOne("select id, name from sql.mock m where m.mockField = ?",
		func(f persistence.Fetchable) error {
			assert.NotNil(t, f)
			var id int
			var name string
			return f.Scan(&id, &name)
		}, "mockValue")
	assert.Nil(t, err)
}

func TestUnitQueryExecErr(t *testing.T) {
	mockErr := errors.New("FetchMockErr")
	rowMock := NewRowMock()
	rowMock.On("Scan", mock.Anything).Return(mockErr)
	dbMock := NewDBMock()
	dbMock.On("Close").Return(nil)
	dbMock.On("QueryRow", mock.Anything, mock.Anything).Return(rowMock)
	persistenceClient := NewClient(dbMock)
	assert.NotNil(t, persistenceClient)
	err := persistenceClient.QueryOne("select * from xql.mock m where m.mockField = ?",
		func(f persistence.Fetchable) error {
			assert.NotNil(t, f)
			scanErr := f.Scan(nil)
			assert.Equal(t, mockErr, scanErr)
			return scanErr
		}, "mockValue")
	assert.NotNil(t, err)
}

func TestUnitExec(t *testing.T) {
	mockResult := NewResultMock()
	dbMock := NewDBMock()
	dbMock.On("Close").Return(nil)
	dbMock.On("Exec", mock.Anything, mock.Anything).Return(mockResult, nil)
	persistenceClient := NewClient(dbMock)
	assert.NotNil(t, persistenceClient)
	err := persistenceClient.Exec("insert into sql.mock values (?)", "mockValue1", "mockValue2")
	assert.Nil(t, err)
}

func TestUnitExecErr(t *testing.T) {
	dbMock := NewDBMock()
	dbMock.On("Close").Return(nil)
	dbMock.On("Exec", mock.Anything, mock.Anything).Return(nil, errors.New("ExecMockErr"))
	persistenceClient := NewClient(dbMock)
	assert.NotNil(t, persistenceClient)
	err := persistenceClient.Exec("insert into sql.mock values (?)", "mockValue1", "mockValue2")
	assert.NotNil(t, err)
}