package index_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HT4w5/autoindex/pkg/index"
)

const (
	maxFileSize     = 1024
	maxEntryCount   = 1024
	randomTestSteps = 128
)

func writeContentMap(t *testing.T, dir string, contentMap map[string]index.Entry) {
	for _, v := range contentMap {
		if v.Type == index.TypeDir {
			d := filepath.Join(dir, v.Name)
			err := os.Mkdir(d, 0600)
			if err != nil {
				t.Fatalf("mkdir error: %v", err)
			}
			info, err := os.Stat(d)
			if err != nil {
				t.Fatalf("stat error: %v", err)
			}
			v.MTime = info.ModTime().Unix()
			contentMap[v.Name] = v
		} else {
			f, err := os.Create(filepath.Join(dir, v.Name))
			if err != nil {
				t.Fatalf("create error: %v", err)
			}
			err = f.Truncate(v.Size)
			if err != nil {
				t.Fatalf("truncate error: %v", err)
			}
			info, err := f.Stat()
			if err != nil {
				t.Fatalf("stat error: %v", err)
			}
			v.MTime = info.ModTime().Unix()
			contentMap[v.Name] = v
		}
	}
}

// Create temporary directory with random files and index.Response representing it
func makeRandomDir(r *rand.Rand, t *testing.T) (string, index.Response) {
	// Generate response first
	resp := index.Response{
		Type: index.TypeDir,
	}

	contentMap := make(map[string]index.Entry)

	n := r.IntN(maxEntryCount)

	for range n {
		e := makeEntry(r)
		contentMap[e.Name] = e
	}

	// Write to filesystem
	dir := t.TempDir()
	writeContentMap(t, dir, contentMap)

	resp.Contents = make([]index.Entry, 0, len(contentMap))
	for _, v := range contentMap {
		resp.Contents = append(resp.Contents, v)
	}

	return dir, resp
}

func makeEntry(r *rand.Rand) index.Entry {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, r.Int64())
	en := index.Entry{
		Name: hex.EncodeToString(buf.Bytes()),
	}
	if r.IntN(2) == 0 {
		// Create directory
		en.Type = index.TypeDir
	} else {
		// Create file
		en.Type = index.TypeFile
		en.Size = r.Int64N(maxFileSize)
	}
	return en
}

func entryEq(t *testing.T, exp index.Entry, got index.Entry) {
	if exp.Name != got.Name {
		t.Error(errMsg("name", exp.Name, got.Name))
	}
	if exp.Type != got.Type {
		t.Error(errMsg("type", exp.Type, got.Type))
	} else {
		if exp.Type == index.TypeFile {
			if exp.Size != got.Size {
				t.Error(errMsg("size", exp.Size, got.Size))
			}
		}
	}
	if exp.MTime != got.MTime {
		t.Error(errMsg("mtime", exp.MTime, got.MTime))
	}
}

func errMsg(name string, exp any, got any) string {
	return fmt.Sprintf(
		"%s mismatch: expected %v, got %v",
		name,
		exp,
		got,
	)
}

func TestRandomDirQuery(t *testing.T) {
	for i := range randomTestSteps {
		var seed [32]byte
		binary.BigEndian.PutUint64(seed[:], uint64(i))
		r := rand.New(rand.NewChaCha8(seed))
		test := func(t *testing.T) {
			dir, respExp := makeRandomDir(r, t)
			idx, err := index.New(
				index.WithRoot(dir),
			)
			if err != nil {
				t.Fatalf("error creating index: %v", err)
			}
			respGot, ok := idx.Query("")
			if !ok {
				t.Error("index query failed")
				return
			}

			if respGot.Type != index.TypeDir {
				t.Error(errMsg("resp.Type", index.TypeDir, respGot.Type))
			}

			// Create got map
			gotMap := make(map[string]index.Entry, 0)
			for _, v := range respGot.Contents {
				gotMap[v.Name] = v
			}

			for _, exp := range respExp.Contents {
				got, ok := gotMap[exp.Name]
				if !ok {
					t.Errorf("missing entry %s", exp.Name)
					continue
				}

				entryEq(t, exp, got)
			}

			if len(respExp.Contents) != len(respGot.Contents) {
				t.Error(errMsg("content length", len(respExp.Contents), len(respGot.Contents)))
			}

			// Query again to test cache
			respGot, ok = idx.Query("")
			if !ok {
				t.Error("index query failed")
				return
			}

			if respGot.Type != index.TypeDir {
				t.Error(errMsg("resp.Type", index.TypeDir, respGot.Type))
			}

			// Create got map
			gotMap = make(map[string]index.Entry, 0)
			for _, v := range respGot.Contents {
				gotMap[v.Name] = v
			}

			for _, exp := range respExp.Contents {
				got, ok := gotMap[exp.Name]
				if !ok {
					t.Errorf("missing entry %s", exp.Name)
					continue
				}

				entryEq(t, exp, got)
			}

			if len(respExp.Contents) != len(respGot.Contents) {
				t.Error(errMsg("content length", len(respExp.Contents), len(respGot.Contents)))
			}
		}
		test(t)
	}
}

func TestQueryFile(t *testing.T) {
	content := map[string]index.Entry{
		"file.dat": {
			Name: "file.dat",
			Size: 1024,
			Type: index.TypeFile,
		},
	}

	dir := t.TempDir()
	writeContentMap(t, dir, content)

	idx, err := index.New(
		index.WithRoot(dir),
	)
	if err != nil {
		t.Fatalf("error creating index: %v", err)
	}

	respGot, ok := idx.Query("/file.dat")
	if !ok {
		t.Error("index query failed")
		return
	}

	if respGot.Type != "file" {
		t.Error(errMsg("type", "file", respGot.Type))
	}
	if respGot.Size != 1024 {
		t.Error(errMsg("type", "file", respGot.Type))
	}

	respGot, ok = idx.Query("/file.dat")
	if !ok {
		t.Error("index query failed")
		return
	}

	if respGot.Type != "file" {
		t.Error(errMsg("type", "file", respGot.Type))
	}
	if respGot.Size != 1024 {
		t.Error(errMsg("type", "file", respGot.Type))
	}
}

const (
	minTestTTL = 0 * time.Second
	maxTestTTL = 5 * time.Second
)

func TestCacheExpiry(t *testing.T) {
	for ttl := minTestTTL; ttl <= maxTestTTL; ttl += 500 * time.Millisecond {
		t.Run(
			fmt.Sprintf("TestCacheExpiry_%s_TTL", ttl.String()),
			func(t *testing.T) {
				t.Parallel()
				content := map[string]index.Entry{
					"file.dat": {
						Name: "file.dat",
						Size: 1024,
						Type: index.TypeFile,
					},
				}

				dir := t.TempDir()
				writeContentMap(t, dir, content)

				idx, err := index.New(
					index.WithRoot(dir),
					index.WithTTL(0),
				)
				if err != nil {
					t.Fatalf("error creating index: %v", err)
				}

				_, ok := idx.Query("/file.dat")
				if !ok {
					t.Error("index query failed")
					return
				}

				err = os.Remove(filepath.Join(dir, "file.dat"))
				if err != nil {
					t.Fatalf("remove error: %v", err)
				}

				<-time.After(ttl)

				_, ok = idx.Query("/file.dat")
				if ok {
					t.Error("cache not expired")
					return
				}
			},
		)
	}
}
