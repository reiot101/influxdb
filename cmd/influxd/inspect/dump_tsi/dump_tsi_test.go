package dump_tsi

import (
	"bytes"
	"fmt"
	"github.com/influxdata/influxdb/v2/models"
	"github.com/influxdata/influxdb/v2/tsdb"
	"github.com/influxdata/influxdb/v2/tsdb/index/tsi1"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

func Test_DumpTSI_NoFile(t *testing.T) {
	//dir, file := makeTSIFile(t, tsiParams{

	//})
}

func Test_DumpTSI_NotAFile(t *testing.T) {
	dir, _ := makeTSIFile(t, tsiParams{})
	defer os.RemoveAll(dir)
	runCommand(t, cmdParams{path: dir, expectErr: true})
}

func Test_DumpTSI_EmptyFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "dumptsi")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	file, err := os.CreateTemp(dir, "*."+tsi1.IndexFileExt)
	require.NoError(t, err)
	runCommand(t, cmdParams{path: file.Name(), expectErr: true})
}

func Test_DumpTSI_IndexFile(t *testing.T) {
	dir, file := makeTSIFile(t, tsiParams{properExt: true})
	fmt.Println(file)
	defer os.RemoveAll(dir)
	fmt.Println(dir)
	runCommand(t, cmdParams{path: dir, args: []string{file}})
}

type cmdParams struct {
	path string
	args []string
	expectErr bool
	expectOut string
}

func runCommand(t *testing.T, params cmdParams) {
	cmd := NewDumpTSICommand()

	b := bytes.NewBufferString("")
	//cmd.SetOut(b)
	//cmd.SetErr(b)

	cmd.SetArgs(append(params.args, "--series-file", params.path, "--series", "--measurements", "--tag-keys", "--tag-value-series"))
	fmt.Println(params.path)
	if params.expectErr {
		require.Error(t, cmd.Execute())
	} else {
		require.NoError(t, cmd.Execute())
	}

	out, err := io.ReadAll(b)
	require.NoError(t, err)
	require.Contains(t, string(out), params.expectOut)
}

type tsiParams struct {
	properExt bool
	badEntry bool
}

func makeTSIFile(t *testing.T, params tsiParams) (string, string) {
	t.Helper()

	dir, err := os.MkdirTemp("", "dumptsi")
	fmt.Println(dir)
	require.NoError(t, err)

	if params.properExt {
		sfile := tsdb.NewSeriesFile(dir)
		require.NoError(t, sfile.Open())
		defer sfile.Close()

		f, err := createIndexFile(t, dir, sfile, []series{
			{Name: []byte("mem"), Tags: models.NewTags(map[string]string{"region": "east"})},
			{Name: []byte("cpu"), Tags: models.NewTags(map[string]string{"region": "east"})},
			{Name: []byte("cpu"), Tags: models.NewTags(map[string]string{"region": "west"})},
		})
		require.NoError(t, err)

		if params.badEntry {
			require.NoError(t, os.WriteFile(f.Path(), []byte("foobar"), 0666))
		}
		return dir, f.Path() // here
	} else {
		file, err := os.CreateTemp(dir, "dumptsi*.txt")
		require.NoError(t, err)
		return dir, file.Name()
	}
}

// series represents name/tagset pairs that are used in testing.
type series struct {
	Name    []byte
	Tags    models.Tags
	Deleted bool
}

// createIndexFile creates an index file with a given set of series.
func createIndexFile(t *testing.T, dir string, sfile *tsdb.SeriesFile, series []series) (*tsi1.IndexFile, error) {
	t.Helper()

	lf, err := createLogFile(t, dir, sfile, series)
	require.NoError(t, err)

	// Write index file to buffer.
	var buf bytes.Buffer
	_, err = lf.CompactTo(&buf, 4096, 6, nil)
	require.NoError(t, err)

	// Load index file from buffer.
	f := tsi1.NewIndexFile(sfile)
	//require.NoError(t, f.Open())
	defer f.Close()
	err = f.UnmarshalBinary(buf.Bytes())
	require.NoError(t, err)
	return f, nil
}

// createLogFile creates a new temporary log file and adds a list of series.
func createLogFile(t *testing.T, dir string, sfile *tsdb.SeriesFile, series []series) (*tsi1.LogFile, error) {
	t.Helper()

	f := newLogFile(t, dir, sfile)
	require.NoError(t, f.Open())
	seriesSet := tsdb.NewSeriesIDSet()
	for _, serie := range series {
		_, err := f.AddSeriesList(seriesSet, [][]byte{serie.Name}, []models.Tags{serie.Tags})
		require.NoError(t, err)
	}
	return f, nil
}

func newLogFile(t *testing.T, dir string, sfile *tsdb.SeriesFile) *tsi1.LogFile {
	t.Helper()

	file, err := os.CreateTemp(dir, "tsi1-log-file-")
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return tsi1.NewLogFile(sfile, file.Name())
}