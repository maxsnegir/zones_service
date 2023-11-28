package psql

//
//var storage *Storage
//
//func TestMain(m *testing.M) {
//	ctx := context.Background()
//	t, err := NewTestStorage(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	storage = t.S
//	code := m.Run()
//	t.Resource.Close()
//
//	os.Exit(code)
//}
//
//func TestDbWork(t *testing.T) {
//	t.Run("test ping", func(t *testing.T) {
//		err := storage.db.Ping(context.Background())
//		if err != nil {
//			t.Fatal(err)
//		}
//	})
//
//	t.Run("simple select", func(t *testing.T) {
//		const expected = 1
//		var actual int
//		err := storage.db.QueryRow(context.Background(), "select 1").Scan(&actual)
//		if err != nil {
//			t.Fatal(err)
//		}
//		if actual != expected {
//			t.Errorf("expected %d, got %d", expected, actual)
//		}
//	})
//
//	t.Run("check postgis installed", func(t *testing.T) {
//		var version string
//		err := storage.db.QueryRow(context.Background(), "select postgis_version()").Scan(&version)
//		if err != nil {
//			t.Fatal(err)
//		}
//		if version == "" {
//			t.Error("empty version")
//		}
//	})
//
//	t.Run("check zone created", func(t *testing.T) {
//		const expected = 0
//		var actual int
//		err := storage.db.QueryRow(context.Background(), "select count(*) from zone").Scan(&actual)
//		if err != nil {
//			t.Fatal(err)
//		}
//		if actual != expected {
//			t.Errorf("expected %d, got %d", expected, actual)
//		}
//	})
//}
