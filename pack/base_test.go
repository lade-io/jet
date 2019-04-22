package pack_test

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/lade-io/jet/pack"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errServer = errors.New("Server failed to start")
	testCases = map[string][]string{}
	testDir   = "../testdata"
	testPacks = []string{}
	testPool  = &dockertest.Pool{MaxWait: 30 * time.Second}
)

func init() {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	packs, err := ioutil.ReadDir(testDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, pack := range packs {
		name := pack.Name()
		packDir := filepath.Join(testDir, name)
		tests, err := ioutil.ReadDir(packDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, test := range tests {
			testCases[name] = append(testCases[name], test.Name())
		}
		testPacks = append(testPacks, name)
	}
	testPool.Client = client
}

func forEachCase(t *testing.T, fn func(t *testing.T, tp, tc, wd string)) {
	for _, tp := range testPacks {
		t.Run(tp, func(t *testing.T) {
			for _, tc := range testCases[tp] {
				tc := tc
				t.Run(tc, func(t *testing.T) {
					wd := filepath.Join(testDir, tp, tc)
					fn(t, tp, tc, wd)
				})
			}
		})
	}
}

func TestDetect(t *testing.T) {
	forEachCase(t, func(t *testing.T, testPack, testCase, workDir string) {
		t.Parallel()
		bp, err := pack.Detect(workDir)
		require.NoError(t, err)
		assert.Contains(t, bp.Metadata.Name, testPack)
	})
}

func TestBuild(t *testing.T) {
	forEachCase(t, func(t *testing.T, testPack, testCase, workDir string) {
		bp, err := pack.Detect(workDir)
		require.NoError(t, err)
		imageID, err := bp.BuildImage(testCase)
		defer assertRemoveImage(t, imageID)
		require.NoError(t, err)

		resource, err := testPool.RunWithOptions(&dockertest.RunOptions{
			Repository:   testCase,
			ExposedPorts: []string{"3000"},
			Env:          []string{"PORT=3000"},
		})
		require.NoError(t, err)

		err = testPool.Retry(func() error {
			container, err := testPool.Client.InspectContainer(resource.Container.ID)
			if err != nil {
				return err
			}
			if !container.State.Running {
				return errServer
			}
			resp, err := http.Get("http://localhost:" + resource.GetPort("3000/tcp"))
			if err != nil {
				return err
			}
			if resp.StatusCode == 500 {
				return errors.New(resp.Status)
			}
			return nil
		})
		assert.NoError(t, err)
		assert.NoError(t, testPool.Purge(resource))
	})
}

func assertRemoveImage(t *testing.T, imageID string) {
	history, err := testPool.Client.ImageHistory(imageID)
	require.NoError(t, err)

	remove := []string{}
	for _, item := range history {
		if item.ID == imageID {
			remove = append(remove, item.ID)
		} else {
			remove = append(remove, item.Tags...)
		}
	}

	for _, image := range remove {
		err = testPool.Client.RemoveImageExtended(image, docker.RemoveImageOptions{
			Force: true,
		})
		assert.NoError(t, err)
	}
}
