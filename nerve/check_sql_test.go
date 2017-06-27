package nerve

//_ "github.com/mattn/go-sqlite3"
//"github.com/n0rad/go-erlog/errs"
//"github.com/stretchr/testify/require"
//"testing"

//func TestSqlCheckFailForWrongUrl(t *testing.T) {
//	check := NewCheckSql()
//	check.Init(&Service{})
//
//	require.Error(t, check.Check())
//}
//
//func TestSqlCheckForGoodUrl(t *testing.T) {
//	check := NewCheckSql()
//	check.Driver = "sqlite3"
//	check.datasource = "/tmp/foo.db"
//	check.Init(&Service{})
//
//	require.NoError(t, check.Check())
//}

//
//func TestMysql(t *testing.T) {
//	check := NewCheckSql()
//	check.Driver = "mysql"
//	check.datasource = "yopla"
//	check.Init(&Service{})
//
//	require.Equal(t, check.Check().(*errs.EntryError).Message, "Check query failed")
//}
