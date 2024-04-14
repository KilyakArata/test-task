package handler

import (
	"avito-testovoe/internal/cache"
	"avito-testovoe/internal/logger"
	sqlite "avito-testovoe/internal/storage"
	"context"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetBannerFromStorage(query sqlite.Query, ctx context.Context) (content string, active bool, err error) {
	args := m.Called(query, ctx)
	return args.Get(0).(string), args.Get(1).(bool), args.Error(2)
}
func (m *MockDB) GetAllBannersFromStorage(query sqlite.Query, ctx context.Context) (banners []sqlite.Banner, err error) {
	args := m.Called(query, ctx)
	return args.Get(0).([]sqlite.Banner), args.Error(1)
}
func (m *MockDB) PostBannerToStorage(banner sqlite.Banner, ctx context.Context) (id int, err error) {
	args := m.Called(banner, ctx)
	return args.Get(0).(int), args.Error(1)
}
func (m *MockDB) UpdateBannerInStorage(banner sqlite.BannerUpdate, ctx context.Context) (err error) {
	args := m.Called(banner, ctx)
	return args.Error(1)
}
func (m *MockDB) DeleteBannerFromStorage(id int, ctx context.Context) (err error) {
	args := m.Called(id, ctx)
	return args.Error(1)
}
func (m *MockDB) DeleteBannerFromStorageByFeature(featureId int, ctx context.Context) (ids []int, err error) {
	args := m.Called(featureId, ctx)
	return args.Get(0).([]int), args.Error(1)
}
func (m *MockDB) DeleteBannerFromStorageByTag(tag int, ctx context.Context) (ids []int, err error) {
	args := m.Called(tag, ctx)
	return args.Get(0).([]int), args.Error(1)
}
func (m *MockDB) CheckToken(token string, ctx context.Context) (role string, err error) {
	args := m.Called(token, ctx)
	return args.Get(0).(string), args.Error(1)
}
func (m *MockDB) GetBannerVersionsFromStorage(id int, ctx context.Context) (banners []sqlite.Banner, err error) {
	args := m.Called(id, ctx)
	return args.Get(0).([]sqlite.Banner), args.Error(1)
}

//TODO: чтобы возвращать ошибку errors.NEW("пример")

// Табличный тест для функции GetBanner
func TestGetBanner(t *testing.T) {
	tests := []struct {
		name            string
		token           string
		tagID           string
		featureID       string
		useLastRevision string
		expectedStatus  int
		args            []any
	}{
		{
			name:           "Нет токена",
			token:          "",
			tagID:          "1",
			featureID:      "1",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Неправильный токен",
			token:          "userWithoutAccess",
			tagID:          "1",
			featureID:      "1",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Нет тега",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			tagID:          "",
			featureID:      "1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Нет фичи",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			tagID:          "1",
			featureID:      "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Некорректный тег",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			tagID:          "abc",
			featureID:      "1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Некорректная фича",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			tagID:          "1",
			featureID:      "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:            "Некорректный запрос версии",
			token:           "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			tagID:           "1",
			featureID:       "1",
			useLastRevision: "некорректный",
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "Запрос последней версии",
			token:           "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			tagID:           "1",
			featureID:       "1",
			useLastRevision: "true",
			expectedStatus:  http.StatusOK,
			args:            []any{sqlite.Query{}, context.Background(), "{\"text\":\"Текст статьи\",\"title\":\"Заголовок 1\",\"url\":\"https://example.com/article\"}", true, nil},
		},
		{
			name:            "Запрос любой версии",
			token:           "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			tagID:           "1",
			featureID:       "1",
			useLastRevision: "false",
			expectedStatus:  http.StatusOK,
			args:            []any{sqlite.Query{}, context.Background(), "{\"text\":\"Текст статьи\",\"title\":\"Заголовок 1\",\"url\":\"https://example.com/article\"}", true, nil},
		},
		{
			name:            "Тег меньше нуля",
			token:           "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			tagID:           "-1",
			featureID:       "1",
			useLastRevision: "false",
			expectedStatus:  http.StatusBadRequest,
		},
		//TODO:
		//{
		//	name:            "Такого банера нет",
		//	token:           "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
		//	tagID:           "-1",
		//	featureID:       "1",
		//	useLastRevision: "false",
		//	expectedStatus:  http.StatusBadRequest,
		//},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			log := logger.SettUpLogger()
			mockS := new(MockDB)
			c := cache.New(60, 60)

			if len(tt.args) > 0 {
				mockS.On("GetBannerFromStorage", tt.args[0].(sqlite.Query), tt.args[1].(context.Context)).Return(tt.args[2].(string), tt.args[3].(bool), tt.args[4].(error))
			}

			mockS.On("CheckToken", tt.args[0].(sqlite.Query), tt.args[1].(context.Context)).Return(tt.args[2].(string), tt.args[3].(bool), tt.args[4].(error))

			// Создание запроса
			req, err := http.NewRequest("GET", "/user_banner?tag_id="+tt.tagID+"&feature_id="+tt.featureID+"&use_last_revision="+tt.useLastRevision, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("token", tt.token)

			h := Handler{
				S:   mockS,
				Log: log,
				C:   c,
				Ctx: ctx,
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.GetBanner)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

// TODO: Табличный тест для функции GetAllBanners
// TODO: Табличный тест для функции PostBanner
// TODO: Табличный тест для функции PatchBanner
// TODO: Табличный тест для функции DeleteBanner
// TODO: Табличный тест для функции DeleteBannerByTagOrFeature
// TODO: Табличный тест для функции GetBannerVersions
// TODO: Табличный тест для функции Verify
