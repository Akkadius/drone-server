// Copyright 2024 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

package stages

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/errors"
	"github.com/drone/drone/mock"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

// For invalid stage number test and stage not found test, it is not applicable for DeclineV2 as we are not dealing with individual stage here.
// We are dealing with list of stages here so no UT is possible for individual stage number verification or individual stage not found.

// this test verifies that a 400 bad request status is returned
// from the http.Handler with a human-readable error message if
// the build number url parameter fails to parse.
func TestDeclineBuild_InvalidBuildNumber(t *testing.T) {
	c := new(chi.Context)
	c.URLParams.Add("owner", "octocat")
	c.URLParams.Add("name", "hello-world")
	c.URLParams.Add("number", "I")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleDeclineBuild(nil, nil, nil)(w, r)
	if got, want := w.Code, 400; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.New("Invalid build number")
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies that a 404 not found status is returned
// from the http.Handler with a human-readable error message if
// the repository is not found in the database.
func TestDeclineBuild_RepoNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockRepo := &core.Repository{
		Namespace: "octocat",
		Name:      "hello-world",
	}

	repos := mock.NewMockRepositoryStore(controller)
	repos.EXPECT().FindName(gomock.Any(), mockRepo.Namespace, mockRepo.Name).Return(nil, sql.ErrNoRows)

	c := new(chi.Context)
	c.URLParams.Add("owner", "octocat")
	c.URLParams.Add("name", "hello-world")
	c.URLParams.Add("number", "1")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleDeclineBuild(repos, nil, nil)(w, r)
	if got, want := w.Code, 404; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.New("Repository not found")
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies that a 404 not found status is returned
// from the http.Handler with a human-readable error message if
// the build is not found in the database.
func TestDeclineBuild_BuildNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockRepo := &core.Repository{
		Namespace: "octocat",
		Name:      "hello-world",
	}

	repos := mock.NewMockRepositoryStore(controller)
	repos.EXPECT().FindName(gomock.Any(), mockRepo.Namespace, mockRepo.Name).Return(mockRepo, nil)

	builds := mock.NewMockBuildStore(controller)
	builds.EXPECT().FindNumber(gomock.Any(), mockRepo.ID, int64(1)).Return(nil, sql.ErrNoRows)

	c := new(chi.Context)
	c.URLParams.Add("owner", "octocat")
	c.URLParams.Add("name", "hello-world")
	c.URLParams.Add("number", "1")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleDeclineBuild(repos, builds, nil)(w, r)
	if got, want := w.Code, 404; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.New("Build not found")
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies that a 400 bad request status is returned
// from the http.Handler with a human-readable error message if
// the build status is not Blocked.
func TestDeclineBuild_InvalidStatus(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockRepo := &core.Repository{
		Namespace: "octocat",
		Name:      "hello-world",
	}
	mockBuild := &core.Build{
		ID:     111,
		Number: 1,
		Status: core.StatusPending,
	}
	mockStage1 := &core.Stage{
		ID:     222,
		Number: 2,
		Status: core.StatusPending,
	}

	mockStage2 := &core.Stage{
		ID:     333,
		Number: 3,
		Status: core.StatusPending,
	}

	mockStageList := []*core.Stage{mockStage1, mockStage2}

	repos := mock.NewMockRepositoryStore(controller)
	repos.EXPECT().FindName(gomock.Any(), mockRepo.Namespace, mockRepo.Name).Return(mockRepo, nil)

	builds := mock.NewMockBuildStore(controller)
	builds.EXPECT().FindNumber(gomock.Any(), mockRepo.ID, int64(1)).Return(mockBuild, nil)

	stages := mock.NewMockStageStore(controller)
	stages.EXPECT().List(gomock.Any(), mockBuild.ID).Return(mockStageList, nil)

	c := new(chi.Context)
	c.URLParams.Add("owner", "octocat")
	c.URLParams.Add("name", "hello-world")
	c.URLParams.Add("number", "1")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleDeclineBuild(repos, builds, stages)(w, r)
	if got, want := w.Code, 400; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.New(`Cannot decline build with status "pending"`)
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}
