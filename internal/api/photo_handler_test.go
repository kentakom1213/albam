package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/indexer"
	"github.com/kentakom1213/albam/internal/model"
	"github.com/kentakom1213/albam/internal/storage"
)

func TestHandleGetPhoto(t *testing.T) {
	store := openTestStorage(t)
	saveTestLibrary(t, store)

	albums, _, err := store.ListAlbums(10, 0, storage.AlbumSortDateDesc)
	if err != nil {
		t.Fatal(err)
	}
	assets, _, err := store.ListAssetsByAlbumSlug(albums[0].Slug, 10, 0, storage.AssetSortTakenAtAsc)
	if err != nil {
		t.Fatal(err)
	}

	server := NewServer(store, config.Config{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/photos/"+assets[0].Slug, nil)

	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body PhotoResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Photo.ID != assets[0].Slug {
		t.Fatalf("photo id = %q, want %q", body.Photo.ID, assets[0].Slug)
	}
	if body.Photo.AlbumID != albums[0].Slug {
		t.Fatalf("album id = %q, want %q", body.Photo.AlbumID, albums[0].Slug)
	}
	if body.Photo.Links.Original != "/media/photos/"+assets[0].Slug+"/original" {
		t.Fatalf("original link = %q", body.Photo.Links.Original)
	}
	if body.Photo.Width == nil || *body.Photo.Width != 1600 {
		t.Fatalf("photo width = %v, want 1600", body.Photo.Width)
	}
	if body.Photo.Height == nil || *body.Photo.Height != 900 {
		t.Fatalf("photo height = %v, want 900", body.Photo.Height)
	}
	if body.Photo.AspectRatio == nil || *body.Photo.AspectRatio != float64(1600)/float64(900) {
		t.Fatalf("photo aspect ratio = %v, want %v", body.Photo.AspectRatio, float64(1600)/float64(900))
	}
	if body.Photo.TakenAt == nil || *body.Photo.TakenAt != "2026-04-02T12:34:56Z" {
		t.Fatalf("photo taken_at = %v, want 2026-04-02T12:34:56Z", body.Photo.TakenAt)
	}
	if body.Photo.GPSLatitude != nil || body.Photo.GPSLongitude != nil {
		t.Fatalf("gps = %v,%v, want hidden", body.Photo.GPSLatitude, body.Photo.GPSLongitude)
	}
	if body.Photo.CameraMake == nil || *body.Photo.CameraMake != "Canon" {
		t.Fatalf("camera make = %v, want Canon", body.Photo.CameraMake)
	}
	if body.Photo.CameraModel == nil || *body.Photo.CameraModel != "EOS R6" {
		t.Fatalf("camera model = %v, want EOS R6", body.Photo.CameraModel)
	}
	if body.Photo.LensModel == nil || *body.Photo.LensModel != "RF24-105mm F4 L IS USM" {
		t.Fatalf("lens model = %v, want RF24-105mm F4 L IS USM", body.Photo.LensModel)
	}
	if body.Photo.FocalLengthMM == nil || *body.Photo.FocalLengthMM != 50 {
		t.Fatalf("focal length = %v, want 50", body.Photo.FocalLengthMM)
	}
	if body.Photo.FocalLength35mm == nil || *body.Photo.FocalLength35mm != 50 {
		t.Fatalf("focal length 35mm = %v, want 50", body.Photo.FocalLength35mm)
	}
	if body.Photo.ApertureFNumber == nil || *body.Photo.ApertureFNumber != 4 {
		t.Fatalf("aperture = %v, want 4", body.Photo.ApertureFNumber)
	}
	if body.Photo.ExposureTimeSeconds == nil || *body.Photo.ExposureTimeSeconds != 0.01 {
		t.Fatalf("exposure time = %v, want 0.01", body.Photo.ExposureTimeSeconds)
	}
	if body.Photo.ISO == nil || *body.Photo.ISO != 400 {
		t.Fatalf("iso = %v, want 400", body.Photo.ISO)
	}
	if body.Photo.Orientation == nil || *body.Photo.Orientation != 1 {
		t.Fatalf("orientation = %v, want 1", body.Photo.Orientation)
	}
}

func TestHandleGetPhotoExposesGPSWhenConfigured(t *testing.T) {
	store := openTestStorage(t)
	saveTestLibrary(t, store)

	albums, _, err := store.ListAlbums(10, 0, storage.AlbumSortDateDesc)
	if err != nil {
		t.Fatal(err)
	}
	assets, _, err := store.ListAssetsByAlbumSlug(albums[0].Slug, 10, 0, storage.AssetSortTakenAtAsc)
	if err != nil {
		t.Fatal(err)
	}

	server := NewServer(store, config.Config{
		PrivacyConfig: config.PrivacyConfig{
			ExposeGPS: true,
		},
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/photos/"+assets[0].Slug, nil)

	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body PhotoResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Photo.GPSLatitude == nil || *body.Photo.GPSLatitude != 35.681236 {
		t.Fatalf("gps latitude = %v, want 35.681236", body.Photo.GPSLatitude)
	}
	if body.Photo.GPSLongitude == nil || *body.Photo.GPSLongitude != 139.767125 {
		t.Fatalf("gps longitude = %v, want 139.767125", body.Photo.GPSLongitude)
	}
}

func TestHandleListAlbumPhotosSortsByTakenAt(t *testing.T) {
	store := openTestStorage(t)
	saveTestLibrary(t, store)

	albums, _, err := store.ListAlbums(10, 0, storage.AlbumSortDateDesc)
	if err != nil {
		t.Fatal(err)
	}

	server := NewServer(store, config.Config{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/albums/"+albums[0].Slug+"/photos?sort=taken_at_desc", nil)
	server.Routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var descBody PhotosResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &descBody); err != nil {
		t.Fatal(err)
	}
	if got := descBody.Photos[0].Filename; got != "PXL_20260502_030405000.jpg" {
		t.Fatalf("newest filename = %q, want PXL_20260502_030405000.jpg", got)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/albums/"+albums[0].Slug+"/photos?sort=taken_at_asc", nil)
	server.Routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var ascBody PhotosResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &ascBody); err != nil {
		t.Fatal(err)
	}
	if got := ascBody.Photos[0].Filename; got != "PXL_20260402_030405000.jpg" {
		t.Fatalf("oldest filename = %q, want PXL_20260402_030405000.jpg", got)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/albums/"+albums[0].Slug+"/photos", nil)
	server.Routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var defaultBody PhotosResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &defaultBody); err != nil {
		t.Fatal(err)
	}
	if got := defaultBody.Photos[0].Filename; got != "PXL_20260402_030405000.jpg" {
		t.Fatalf("default filename = %q, want PXL_20260402_030405000.jpg", got)
	}
}

func TestHandleGetPhotoNotFound(t *testing.T) {
	store := openTestStorage(t)

	server := NewServer(store, config.Config{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/photos/missing-photo", nil)

	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func openTestStorage(t *testing.T) *storage.Storage {
	t.Helper()

	store, err := storage.Open(filepath.Join(t.TempDir(), "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}
	})

	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}

	return store
}

func saveTestLibrary(t *testing.T, store *storage.Storage) {
	t.Helper()

	takenAt := time.Date(2026, 4, 2, 12, 34, 56, 0, time.UTC)
	newerTakenAt := time.Date(2026, 5, 2, 12, 34, 56, 0, time.UTC)
	gpsLatitude := 35.681236
	gpsLongitude := 139.767125
	cameraMake := "Canon"
	cameraModel := "EOS R6"
	lensMake := "Canon"
	lensModel := "RF24-105mm F4 L IS USM"
	focalLengthMM := 50.0
	focalLength35mm := 50
	apertureFNumber := 4.0
	exposureTimeSeconds := 0.01
	iso := 400
	orientation := 1

	if err := store.SaveLibrary(&indexer.Library{
		Albums: []model.Album{
			{Path: "weekend-trip", Title: "Weekend Trip"},
		},
		Assets: []model.Asset{
			{
				Path:                "weekend-trip/PXL_20260402_030405000.jpg",
				Filename:            "PXL_20260402_030405000.jpg",
				Ext:                 ".jpg",
				Size:                123,
				ModTime:             time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
				Width:               1600,
				Height:              900,
				TakenAt:             &takenAt,
				GPSLatitude:         &gpsLatitude,
				GPSLongitude:        &gpsLongitude,
				CameraMake:          &cameraMake,
				CameraModel:         &cameraModel,
				LensMake:            &lensMake,
				LensModel:           &lensModel,
				FocalLengthMM:       &focalLengthMM,
				FocalLength35mm:     &focalLength35mm,
				ApertureFNumber:     &apertureFNumber,
				ExposureTimeSeconds: &exposureTimeSeconds,
				ISO:                 &iso,
				Orientation:         &orientation,
			},
			{
				Path:     "weekend-trip/PXL_20260502_030405000.jpg",
				Filename: "PXL_20260502_030405000.jpg",
				Ext:      ".jpg",
				Size:     456,
				ModTime:  time.Date(2026, 1, 3, 3, 4, 5, 0, time.UTC),
				Width:    900,
				Height:   1600,
				TakenAt:  &newerTakenAt,
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
}
