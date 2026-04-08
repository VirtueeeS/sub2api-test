package repository

import (
	"context"
	"regexp"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestListTodayUsageTotalsByAccountIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newAccountRepositoryWithSQL(nil, db, nil)
	rows := sqlmock.NewRows([]string{"account_id", "usage_total"}).
		AddRow(int64(101), 12.5).
		AddRow(int64(202), 3.25)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT account_id, COALESCE(SUM(total_cost * COALESCE(account_rate_multiplier, 1)), 0) AS usage_total
		FROM usage_logs
		WHERE account_id = ANY($1) AND created_at >= $2
		GROUP BY account_id
	`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	totals, err := repo.listTodayUsageTotalsByAccountIDs(context.Background(), []int64{101, 202})
	require.NoError(t, err)
	require.Equal(t, map[int64]float64{
		101: 12.5,
		202: 3.25,
	}, totals)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSortAccountsByUsageTotals(t *testing.T) {
	t.Run("desc", func(t *testing.T) {
		accounts := []*dbent.Account{
			{ID: 1},
			{ID: 2},
			{ID: 3},
		}

		sortAccountsByUsageTotals(accounts, map[int64]float64{
			1: 10,
			2: 10,
			3: 5,
		}, service.AccountListSortOrderDesc)

		require.Equal(t, []int64{2, 1, 3}, accountEntityIDs(accounts))
	})

	t.Run("asc", func(t *testing.T) {
		accounts := []*dbent.Account{
			{ID: 1},
			{ID: 2},
			{ID: 3},
		}

		sortAccountsByUsageTotals(accounts, map[int64]float64{
			1: 10,
			2: 10,
			3: 5,
		}, service.AccountListSortOrderAsc)

		require.Equal(t, []int64{3, 1, 2}, accountEntityIDs(accounts))
	})
}

func accountEntityIDs(accounts []*dbent.Account) []int64 {
	ids := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
	}
	return ids
}
