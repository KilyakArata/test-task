package sqlite

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuthorization(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		role    string
		ctx     context.Context
		errNeed bool
	}{
		{
			name:    "Нет токена",
			token:   "",
			ctx:     context.Background(),
			errNeed: true,
		},
		{
			name:    "Нет такого токена",
			token:   "1766c485cd55dc48caf9accbf97c2733d9e4a72b9372f9ac873335a86d77f56c",
			ctx:     context.Background(),
			errNeed: true,
		},
		{
			name:    "Проверка роли админа",
			token:   "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			role:    "Admin",
			ctx:     context.Background(),
			errNeed: false,
		},
		{
			name:    "Проверка роли пользователя",
			token:   "b512d97e7cbf97c273e4db073bbb547aa65a84589227f8f3d9e4a72b9372a24d",
			role:    "User",
			ctx:     context.Background(),
			errNeed: false,
		},
	}

	db, err := sql.Open("sqlite", "storage.db")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Storage{Db: db}

			roleTake, err := s.CheckToken(tt.token, tt.ctx)
			if err != nil {
				if tt.errNeed {
					require.Error(t, err)
				} else {
					t.Fatal(err)
				}
			}
			assert.Equal(t, tt.role, roleTake)
		})
	}
}
