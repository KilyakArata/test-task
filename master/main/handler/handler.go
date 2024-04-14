package handler

import (
	sqlite "avito-testovoe/internal/storage"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

// GetBanner Получение баннера для пользователя
func (h *Handler) GetBanner(w http.ResponseWriter, r *http.Request) {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	token := r.Header.Get("token")

	ok := h.Verify(token, ReadPermission, w)

	if !ok {
		return
	}

	var query sqlite.Query

	tag, err := strconv.Atoi(r.URL.Query().Get("tag_id"))
	if err != nil {
		h.Log.Error("Некорректный тег")
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	query.TagId = tag

	query.FeatureId, err = strconv.Atoi(r.URL.Query().Get("feature_id"))
	if err != nil {
		h.Log.Error("Некорректная фича")
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	revision := r.URL.Query().Get("use_last_revision")
	if revision != "" {
		if query.Revision, err = strconv.ParseBool(revision); err != nil {
			h.Log.Error("Некорректная версия запрашивается")
			http.Error(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
	}

	if query.FeatureId <= 0 || query.TagId <= 0 {
		h.Log.Error("Некорректные данные")
		http.Error(w, "Фича или тег некорректны", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("%d %d", query.FeatureId, query.TagId)

	bannerCache, activeCache, ok := h.C.Get(key)
	if ok && !query.Revision {
		if !activeCache {
			ok = h.Verify(token, WritePermission, w)
			if !ok {
				return
			}
		}

		resp, err := json.Marshal(bannerCache)
		if err != nil {
			h.Log.Error("Внутренняя ошибка сервера:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resp)
		if err != nil {
			h.Log.Error("Внутренняя ошибка сервера:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.Log.Info("Получен баннер пользователя из кэша")
		return
	}

	banner, active, err := h.S.GetBannerFromStorage(query, h.Ctx)
	if err != nil {
		h.Log.Error("Баннер не найден:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if !active {
		ok = h.Verify(token, WritePermission, w)
		if !ok {
			return
		}
	}

	var content map[string]string
	err = json.Unmarshal([]byte(banner), &content)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(content)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.C.Set(key, active, content)
	h.Log.Info("Получен баннер пользователя из базы данных")
}

// GetAllBanners Получение всех баннеров c фильтрацией по фиче и/или тегу
func (h *Handler) GetAllBanners(w http.ResponseWriter, r *http.Request) {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	token := r.Header.Get("token")

	ok := h.Verify(token, WritePermission, w)

	if !ok {
		return
	}

	var query sqlite.Query

	tag := r.URL.Query().Get("tag_id")
	if tag != "" {
		tagId, err := strconv.Atoi(tag)
		if err != nil {
			h.Log.Error("Некорректные данные")
			http.Error(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		query.TagId = tagId
	}

	feature := r.URL.Query().Get("feature_id")
	if feature != "" {
		featureId, err := strconv.Atoi(feature)
		if err != nil {
			h.Log.Error("Некорректные данные")
			http.Error(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		query.FeatureId = featureId
	}
	limit := r.URL.Query().Get("limit")
	if limit != "" {
		limitquery, err := strconv.Atoi(limit)
		if err != nil {
			h.Log.Error("Некорректные данные")
			http.Error(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		query.Limit = limitquery
	}

	offset := r.URL.Query().Get("offset")
	if offset != "" {
		offsetquery, err := strconv.Atoi(offset)
		if err != nil {
			h.Log.Error("Некорректные данные")
			http.Error(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		query.Offset = offsetquery
	}

	if query.FeatureId == 0 && query.TagId == 0 {
		h.Log.Error("Некорректные данные: нет указателя на фичу и тег")
		http.Error(w, "Некорректные данные: нет указателя на фичу и тег", http.StatusBadRequest)
		return
	}

	banners, err := h.S.GetAllBannersFromStorage(query, h.Ctx)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(banners)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Log.Info("Получены баннеры по запросу пользователя")
}

// PostBanner Создание нового баннера
func (h *Handler) PostBanner(w http.ResponseWriter, r *http.Request) {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	token := r.Header.Get("token")

	ok := h.Verify(token, WritePermission, w)

	if !ok {
		return
	}

	var banner sqlite.Banner
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		h.Log.Error("Некорректные данные:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &banner); err != nil {
		h.Log.Error("Некорректные данные:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if banner.FeatureId < 1 {
		w.WriteHeader(http.StatusBadRequest)
		h.Log.Error("Некорректные данные")
		return
	}

	for _, tag := range banner.TagIds {
		if tag < 1 {
			w.WriteHeader(http.StatusBadRequest)
			h.Log.Error("Некорректные данные")
			return
		}
	}

	idLastBanner, err := h.S.PostBannerToStorage(banner, h.Ctx)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stringId := fmt.Sprint(idLastBanner)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(stringId))
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Log.Info("Добавлен новый баннер по запросу пользователя под номером:" + stringId)
}

// PatchBanner Обновление содержимого баннера
func (h *Handler) PatchBanner(w http.ResponseWriter, r *http.Request) {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	id := chi.URLParam(r, "id")

	token := r.Header.Get("token")

	ok := h.Verify(token, WritePermission, w)

	if !ok {
		return
	}

	var banner sqlite.BannerUpdate
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)

	if err != nil {
		h.Log.Error("Некорректные данные:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &banner); err != nil {
		h.Log.Error("Некорректные данные:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	banner.BannerId, err = strconv.Atoi(id)
	if err != nil {
		h.Log.Error("Некорректные данные:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if banner.FeatureId < 1 {
		w.WriteHeader(http.StatusBadRequest)
		h.Log.Error("Некорректные данные")
		return
	}

	for _, tag := range banner.TagIds {
		if tag < 1 {
			w.WriteHeader(http.StatusBadRequest)
			h.Log.Error("Некорректные данные")
			return
		}
	}

	err = h.S.UpdateBannerInStorage(banner, h.Ctx)
	if err != nil {
		h.Log.Error("Баннер не найден:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	for _, tag := range banner.TagIds {
		key := fmt.Sprintf("%d %d", banner.FeatureId, tag)
		h.C.Set(key, banner.IsActive, banner.Content)
	}

	w.WriteHeader(http.StatusOK)
	h.Log.Info("Обновлен баннер по запросу пользователя под номером:" + id)
}

// DeleteBanner Удаление баннера по идентификатору
func (h *Handler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	id := chi.URLParam(r, "id")

	token := r.Header.Get("token")

	ok := h.Verify(token, WritePermission, w)

	if !ok {
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.Log.Error("Некорректные данные:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	keys, err := h.S.DeleteBannerFromStorage(idInt, h.Ctx)
	if err != nil {
		h.Log.Error("Баннер не найден:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	h.C.Delete(keys)

	w.WriteHeader(http.StatusNoContent)
	h.Log.Info("Удален баннер по запросу пользователя под номером:" + id)
}

// DeleteBannerByTagOrFeature Удаление баннера по тэгу или фиче
func (h *Handler) DeleteBannerByTagOrFeature(w http.ResponseWriter, r *http.Request) {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	token := r.Header.Get("token")

	ok := h.Verify(token, WritePermission, w)

	if !ok {
		return
	}

	var query sqlite.Query

	tag := r.URL.Query().Get("tag_id")
	if tag != "" {
		tagId, err := strconv.Atoi(tag)
		if err != nil {
			h.Log.Error("Некорректные данные")
			http.Error(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		query.TagId = tagId
	}

	feature := r.URL.Query().Get("feature_id")
	if feature != "" {
		featureId, err := strconv.Atoi(feature)
		if err != nil {
			h.Log.Error("Некорректные данные")
			http.Error(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		query.FeatureId = featureId
	}

	if (query.FeatureId > 0 && query.TagId > 0) || (query.FeatureId < 0 || query.TagId < 0) {
		h.Log.Error("Некорректные данные")
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.Log.Info("Удалены баннеры по запросу пользователя")

	go func() {
		keys := []string{}
		if query.FeatureId > 0 {
			keys1, err := h.S.DeleteBannerFromStorageByFeature(query.FeatureId, h.Ctx)
			if err != nil {
				h.Log.Error("Баннер не найден по фиче:", err)
			}
			keys = keys1
		} else {
			keys2, err := h.S.DeleteBannerFromStorageByTag(query.TagId, h.Ctx)
			if err != nil {
				h.Log.Error("Баннер не найден по тегу:", err)
			}
			keys = keys2
		}
		h.C.Delete(keys)
	}()

}

// GetBannerVersions Получение старыйх версий баннера
func (h *Handler) GetBannerVersions(w http.ResponseWriter, r *http.Request) {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	id := chi.URLParam(r, "id")

	token := r.Header.Get("token")

	ok := h.Verify(token, WritePermission, w)

	if !ok {
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.Log.Error("Некорректные данные:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	banners, err := h.S.GetBannerVersionsFromStorage(idInt, h.Ctx)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(banners)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		h.Log.Error("Внутренняя ошибка сервера:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Log.Info("Получены старые версии баннера по запросу пользователя")

}
