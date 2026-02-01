package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	p := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	num, err := store.Add(p)

	require.NoError(t, err)
	require.NotEmpty(t, num)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	newParcel, err := store.Get(num)
	require.NoError(t, err)

	assert.Equal(t, p.Client, newParcel.Client)
	assert.Equal(t, p.Status, newParcel.Status)
	assert.Equal(t, p.Address, newParcel.Address)
	assert.Equal(t, p.CreatedAt, newParcel.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	assert.NoError(t, store.Delete(num))
	newParcel, err = store.Get(num)
	assert.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	p := getTestParcel()
	num, err := store.Add(p)

	require.NoError(t, err)
	require.NotEmpty(t, num)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	assert.NoError(t, store.SetAddress(num, newAddress))

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	addedParcel, err := store.Get(num)
	assert.Equal(t, newAddress, addedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)

	// add
	p := getTestParcel()
	num, err := store.Add(p)

	require.NoError(t, err)
	require.NotEmpty(t, num)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	assert.NoError(t, store.SetStatus(num, ParcelStatusSent))

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	addedParcel, err := store.Get(num)
	assert.Equal(t, ParcelStatusSent, addedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	// убедитесь в отсутствии ошибки
	require.NoError(t, err)
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	assert.Equal(t, len(parcels), len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		require.NotEmpty(t, parcelMap[parcel.Number])
		assert.Equal(t, parcelMap[parcel.Number].Address, parcel.Address)
		assert.Equal(t, parcelMap[parcel.Number].Client, parcel.Client)
		assert.Equal(t, parcelMap[parcel.Number].CreatedAt, parcel.CreatedAt)
		assert.Equal(t, parcelMap[parcel.Number].Status, parcel.Status)
	}
}
