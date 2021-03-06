package mock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestObject(t *testing.T) {
	object := Object{
		"key_string":  "mock_key",
		"key_integer": float64(999),
		"key_float":   float64(999.99),
		"key_time":    time.Now().Format(time.RFC3339),
		"key_object": map[string]interface{}{
			"object_key": "object_value",
		},
	}

	driverValue, err := object.Value()
	require.Nil(t, err, "object.value error")
	require.IsType(t, ([]byte)(nil), driverValue, "drivervalue invalid type")

	var jsonObject Object

	err = jsonObject.Scan(new(struct{}))
	require.NotNil(t, err, "object.scan with bad source invalid result")

	err = jsonObject.Scan([]byte(`<xml><value>bad json<value></xml>`))
	require.NotNil(t, err, "object.scan with bad source invalid result")

	err = jsonObject.Scan(driverValue)
	require.Nil(t, err, "object.scan error")

	require.Equal(t, object, jsonObject, "object invalid instance")
}

func TestMockRepository(t *testing.T) {
	var (
		ctx        = context.Background()
		repository = NewMockRepository()
		key        = NewMockEntityKey()
		entity     = NewMockEntity()
		result     = NewMockEntity()
	)
	key.On("EntityName").Return("mockEntityName")
	key.On("Name").Return("mockKeyName")
	key.On("Value").Return("mockKeyValue")
	repository.On("Set", ctx, mock.Anything, mock.Anything).Return(nil)
	repository.On("Get", ctx, mock.Anything, result).Run(
		func(args mock.Arguments) {
			getKey := args.Get(1).(*MockEntityKey)
			require.NotZero(t, getKey.EntityName(), "key.entityname invalid instance")
			require.NotZero(t, getKey.Name(), "key.name invalid instance")
			require.NotZero(t, getKey.Value(), "key.value invalid instance")
			entityResult := args.Get(2).(*MockEntity)
			*entityResult = *entity
		},
	).Return(nil)
	repository.On("Delete", ctx, mock.Anything).Return(nil)
	repository.On("Close", mock.Anything).Return(nil)

	repository.Set(ctx, key, entity)
	repository.Get(ctx, key, result)
	require.Exactly(t, entity, result, "get invalid instance")
	repository.Delete(ctx, key)
	repository.Close(ctx)

	repository.AssertExpectations(t)
}
