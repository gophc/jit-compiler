package sys

import (
	"github.com/bspaans/jit-compiler/platform/sys/require"
	"io/fs"
	"os"
	"path"
	"runtime"
	"testing"
	"testing/fstest"
)

func Test_NewStat_t(t *testing.T) {
	tmpDir := t.TempDir()
	fileData := []byte{1, 2, 3, 4}

	dir := path.Join(tmpDir, "dir")
	require.NoError(t, os.Mkdir(dir, 0o700))
	osDirInfo, err := os.Stat(dir)
	require.NoError(t, err)

	file := path.Join(dir, "file")
	require.NoError(t, os.WriteFile(file, []byte{1, 2, 3, 4}, 0o400))
	osFileInfo, err := os.Stat(file)
	require.NoError(t, err)

	// A required privilege is not held by the client.
	/* link := path.Join(dir, "file-link")
	require.NoError(t, os.Symlink(file, link))
	osSymlinkInfo, err := os.Lstat(link)
	require.NoError(t, err) */

	osFileSt := NewStat_t(osFileInfo)
	testFS := fstest.MapFS{
		"dir": {
			Mode:    osDirInfo.Mode(),
			ModTime: osDirInfo.ModTime(),
		},
		"dir/file": {
			Data:    fileData,
			Mode:    osFileInfo.Mode(),
			ModTime: osFileInfo.ModTime(),
		},
		"dir/file-sys": {
			// intentionally skip other fields to prove sys is read.
			Sys: &osFileSt,
		},
	}

	fsDirInfo, err := testFS.Stat("dir")
	require.NoError(t, err)
	fsFileInfo, err := testFS.Stat("dir/file")
	require.NoError(t, err)
	fsFileInfoWithSys, err := testFS.Stat("dir/file-sys")
	require.NoError(t, err)

	tests := []struct {
		name            string
		info            fs.FileInfo
		expectDevIno    bool
		expectedMode    fs.FileMode
		expectedSize    int64
		expectAtimCtime bool
	}{
		{
			name:            "os dir",
			info:            osDirInfo,
			expectDevIno:    true,
			expectedMode:    fs.ModeDir | 0o0700,
			expectedSize:    osDirInfo.Size(), // OS dependent
			expectAtimCtime: true,
		},
		{
			name:            "fs dir",
			info:            fsDirInfo,
			expectDevIno:    false,
			expectedMode:    fs.ModeDir | 0o0700,
			expectedSize:    0,
			expectAtimCtime: false,
		},
		{
			name:            "os file",
			info:            osFileInfo,
			expectDevIno:    true,
			expectedMode:    0o0400,
			expectedSize:    int64(len(fileData)),
			expectAtimCtime: true,
		},
		{
			name:            "fs file",
			info:            fsFileInfo,
			expectDevIno:    false,
			expectedMode:    0o0400,
			expectedSize:    int64(len(fileData)),
			expectAtimCtime: false,
		},
		{
			name:            "fs file with Stat_t in Sys",
			info:            fsFileInfoWithSys,
			expectDevIno:    true,
			expectedMode:    0o0400,
			expectedSize:    int64(len(fileData)),
			expectAtimCtime: true,
		},

		// A required privilege is not held by the client.
		/* {
			name:            "os symlink",
			info:            osSymlinkInfo,
			expectDevIno:    true,
			expectedMode:    fs.ModeSymlink,
			expectedSize:    osSymlinkInfo.Size(), // OS dependent
			expectAtimCtime: true,
		}, */
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			st := NewStat_t(tc.info)

			//goland:noinspection GoBoolExpressions
			if tc.expectDevIno && runtime.GOOS != "windows" {
				require.NotEqual(t, uint64(0), st.Dev)
				require.NotEqual(t, uint64(0), st.Ino)
			} else {
				require.Zero(t, st.Dev)
				require.Zero(t, st.Ino)
			}

			// link mode may differ on windows, so mask
			require.Equal(t, tc.expectedMode, st.Mode&tc.expectedMode)

			//goland:noinspection GoBoolExpressions
			if SysParseable && runtime.GOOS != "windows" {
				switch st.Nlink {
				case 2, 4: // dirents may include dot entries.
					require.Equal(t, fs.ModeDir, st.Mode.Type())
				default:
					require.Equal(t, uint64(1), st.Nlink)
				}
			} else { // Nlink is possibly wrong, but not zero.
				require.Equal(t, uint64(1), st.Nlink)
			}

			require.Equal(t, tc.expectedSize, st.Size)

			if tc.expectAtimCtime && SysParseable {
				// We don't validate times strictly because it is os-dependent
				// what updates times. There are edge cases for symlinks, too.
				require.NotEqual(t, EpochNanos(0), st.Ctim)
				require.NotEqual(t, EpochNanos(0), st.Mtim)
				require.NotEqual(t, EpochNanos(0), st.Mtim)
			} else { // mtim is used for atim and ctime
				require.Equal(t, st.Mtim, st.Ctim)
				require.NotEqual(t, EpochNanos(0), st.Mtim)
				require.Equal(t, st.Mtim, st.Atim)
			}
		})
	}
}
