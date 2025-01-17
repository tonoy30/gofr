package shop

import (
	"os"
	"strconv"
	"testing"

	"gofr.dev/examples/using-ycql/models"
	"gofr.dev/pkg/datastore"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/config"
	"gofr.dev/pkg/log"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger := log.NewLogger()
	c := config.NewGoDotEnvProvider(logger, "../../configs")
	cassandraPort, _ := strconv.Atoi(c.Get("CASS_DB_PORT"))
	ycqlCfg := datastore.CassandraCfg{
		Hosts:    c.Get("CASS_DB_HOST"),
		Port:     cassandraPort,
		Username: c.Get("CASS_DB_USER"),
		Password: c.Get("CASS_DB_PASS"),
		Keyspace: "system",
	}

	ycqlDB, err := datastore.GetNewYCQL(logger, &ycqlCfg)
	if err != nil {
		logger.Errorf("got error while connecting to YCQL")
	}

	err = ycqlDB.Session.Query(
		"CREATE KEYSPACE IF NOT EXISTS test WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': '1'} " +
			"AND DURABLE_WRITES = true;").Exec()
	if err != nil {
		logger.Errorf("got error while connecting to YCQL")
	}

	ycqlCfg.Keyspace = "test"

	ycqlDB, err = datastore.GetNewYCQL(logger, &ycqlCfg)
	if err != nil {
		logger.Errorf("got error while connecting to YCQL")
	}

	os.Exit(m.Run())
}

func initializeTest(t *testing.T) *gofr.Context {
	app := gofr.New()
	// initializing the seeder
	sd := datastore.NewSeeder(&app.DataStore, "../../db")
	q := "CREATE TABLE IF NOT EXISTS shop (id int PRIMARY KEY, name varchar, location varchar , state varchar ) " +
		"WITH transactions = { 'enabled' : true };"

	err := app.YCQL.Session.Query(q).Exec()
	if err != nil {
		t.Errorf("[Test_Init]\tYCQL failed during table creation\n")
	}

	ctx := gofr.NewContext(nil, nil, app)

	sd.RefreshYCQL(t, "shop")

	return ctx
}

func TestGet(t *testing.T) {
	tests := []struct {
		desc  string
		input models.Shop
		resp  []models.Shop
		err   error
	}{
		{"get by id-SUCCESS", models.Shop{ID: 1}, []models.Shop{{ID: 1, Name: "Pramod", Location: "Gaya", State: "Bihar"}}, nil},
		{"get by name-SUCCESS", models.Shop{Name: "Pramod"}, []models.Shop{{ID: 1, Name: "Pramod", Location: "Gaya", State: "Bihar"}}, nil},
		{"get by all fields-SUCCESS", models.Shop{Name: "Pramod", ID: 1, State: "Bihar", Location: "Gaya"},
			[]models.Shop{{ID: 1, Name: "Pramod", Location: "Gaya", State: "Bihar"}}, nil},
		{"get by empty fields-SUCCESS", models.Shop{}, []models.Shop{
			{ID: 1, Name: "Pramod", Location: "Gaya", State: "Bihar"}, {ID: 2, Name: "Shubh", Location: "HSR", State: "Karnataka"}}, nil},
		{"get unknown shop item-SUCCESS", models.Shop{ID: 9, State: "Bihar"}, nil, nil},
	}

	ctx := initializeTest(t)

	store := New()

	for i, tc := range tests {
		resp := store.Get(ctx, tc.input)

		assert.Equal(t, tc.resp, resp, "TEST[%d], failed.\n%s", i, tc.desc)
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		desc  string
		input models.Shop
		resp  []models.Shop
		err   error
	}{
		{"create with all fields-SUCCESS", models.Shop{ID: 1, Name: "himalaya", Location: "Gaya", State: "bihar"},
			[]models.Shop{{ID: 1, Name: "himalaya", Location: "Gaya", State: "bihar"}}, nil},
	}

	ctx := initializeTest(t)

	store := New()

	for i, tc := range tests {
		resp, err := store.Create(ctx, tc.input)

		assert.Equal(t, tc.resp, resp, "TEST[%d], failed.\n%s", i, tc.desc)

		assert.Equal(t, tc.err, err, "TEST[%d], failed.\n%s", i, tc.desc)
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		desc  string
		input models.Shop
		resp  []models.Shop
		err   error
	}{
		{"update by id", models.Shop{ID: 2}, []models.Shop{{ID: 2, Name: "Shubh", Location: "HSR", State: "Karnataka"}}, nil},
		{"udpate all fields", models.Shop{ID: 2, Name: "Mahi", Location: "Dhanbad", State: "Jharkhand"},
			[]models.Shop{{ID: 2, Name: "Mahi", Location: "Dhanbad", State: "Jharkhand"}}, nil},
		{"udpate few fields", models.Shop{ID: 2, Location: "Gaya", State: "Bihar"},
			[]models.Shop{{ID: 2, Name: "Mahi", Location: "Gaya", State: "Bihar"}}, nil},
	}

	ctx := initializeTest(t)

	store := New()

	for i, tc := range tests {
		resp, err := store.Update(ctx, tc.input)

		assert.Equal(t, tc.resp, resp, "TEST[%d], failed.\n%s", i, tc.desc)

		assert.Equal(t, tc.err, err, "TEST[%d], failed.\n%s", i, tc.desc)
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		err   error
	}{
		{"delete by id-SUCCESS", "3", nil},
	}

	ctx := initializeTest(t)

	store := New()

	for i, tc := range tests {
		err := store.Delete(ctx, tc.input)

		assert.Equal(t, tc.err, err, "TEST[%d], failed.\n%s", i, tc.desc)
	}
}

func Test_Errors(t *testing.T) {
	ctx := initializeTest(t)
	ctx.YCQL.Session.Close() // to generate errors

	store := New()

	// error test for create
	_, err := store.Create(ctx, models.Shop{})
	assert.NotNil(t, err, "TEST:Create, failed.\n")

	// error test for Update
	_, err = store.Update(ctx, models.Shop{ID: 1, Name: "Name_Update"})
	assert.NotNil(t, err, "TEST:Update, failed.\n")

	// error test for Delete
	err = store.Delete(ctx, "1")
	assert.NotNil(t, err, "TEST:Delete, failed.\n")
}
