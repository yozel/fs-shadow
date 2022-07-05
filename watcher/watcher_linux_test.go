package watcher

import (
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_UseCase(t *testing.T) {
	testRoot := "/tmp/fs-shadow"
	_ = os.Mkdir(testRoot, os.ModePerm)

	tw, err := newLinuxPathWatcher(testRoot)
	assert.Equal(t, nil, err, "linux patch watcher creation error")

	// create folder
	folderName := "test1"
	folder := filepath.Join(testRoot, folderName)
	_ = os.Mkdir(folder, os.ModePerm)
	time.Sleep(3 * time.Second)
	assert.Equal(t, folderName, tw.FileTree.Subs[0].Name, "create:invalid folder name")

	// rename folder
	newFolderName := "test1-rename"
	renameFolder := filepath.Join(testRoot, newFolderName)
	_ = os.Rename(folder, renameFolder)
	time.Sleep(2 * time.Second)
	assert.Equal(t, newFolderName, tw.FileTree.Subs[0].Name, "rename:invalid folder name")

	// move to other directory
	moveDirectory := "/tmp/test1-rename"
	err = os.Rename(renameFolder, moveDirectory)
	time.Sleep(2 * time.Second)
	assert.Equal(t, 0, len(tw.FileTree.Subs), "remove:invalid subs length")

	tw.Close()
	_ = os.RemoveAll(testRoot)
	_ = os.Remove(moveDirectory)

}

func Test_Functionality(t *testing.T) {
	var err error
	var watcher *fsnotify.Watcher
	parentPath := "/tmp"
	testRoot := filepath.Join(parentPath, "fs-shadow")
	_ = os.Mkdir(testRoot, os.ModePerm)

	path := connector.NewFSPath(testRoot)

	watcher, err = fsnotify.NewWatcher()
	assert.Equal(t, nil, err, "watcher creation error")

	root := filenode.FileNode{
		Name: path.Name(),
		Meta: filenode.MetaData{
			IsDir: true,
		},
	}

	tw := TreeWatcher{
		FileTree:     &root,
		ParentPath:   path.ParentPath(),
		Path:         path,
		Watcher:      watcher,
		EventManager: event.NewEventHandler(),
	}
	err = tw.Create(path)
	assert.Equal(t, nil, err, "root node creation error")

	// Create folder
	newFolder := connector.NewFSPath(filepath.Join(testRoot, "folder"))
	_ = os.Mkdir(newFolder.String(), os.ModePerm)
	err = tw.Create(newFolder)
	assert.Equal(t, nil, err, "folder node creation error")
	assert.Equal(t, newFolder.Name(), tw.FileTree.Subs[0].Name, "create:invalid folder name")

	// Create file
	newFile := connector.NewFSPath(filepath.Join(testRoot, "file.txt"))
	_, _ = os.Create(newFile.String())
	err = tw.Create(newFile)
	assert.Equal(t, nil, err, "file node creation error")
	assert.Equal(t, newFile.Name(), tw.FileTree.Subs[1].Name, "create:invalid file name")

	// Rename
	renameFilePath := connector.NewFSPath(filepath.Join(testRoot, "file-rename.txt"))
	_ = os.Rename(newFile.String(), renameFilePath.String())
	err = tw.Rename(newFile, renameFilePath)
	assert.Equal(t, nil, err, "file node rename error")
	assert.Equal(t, renameFilePath.Name(), tw.FileTree.Subs[1].Name, "rename:filename is not changed")

	// Write
	renameEventFilePath := renameFilePath.ExcludePath(connector.NewFSPath(parentPath))
	node := tw.FileTree.Search(renameEventFilePath.String())
	assert.NotEqual(t, nil, node, "renamed file not found")
	oldSum := node.Meta.Sum

	f, _ := os.OpenFile(renameFilePath.String(), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	_, err = f.WriteString("test")
	_ = f.Close()

	err = tw.Write(renameFilePath)
	assert.Equal(t, nil, err, "file node write error")
	node = tw.FileTree.Search(renameEventFilePath.String())
	assert.NotEqual(t, oldSum, node.Meta.Sum, "updated file sums not equal")

	// Remove
	err = tw.Remove(renameFilePath)
	assert.Equal(t, nil, err, "file node remove error")
	assert.Equal(t, 1, len(root.Subs), "file node not removed")

	// Handler Create
	newFileName := connector.NewFSPath(filepath.Join(testRoot, "new-file.txt"))
	_, _ = os.Create(newFileName.String())
	e := event.Event{FromPath: newFileName.String(), Type: event.Create}
	err = tw.Handler(e)
	assert.Equal(t, nil, err, "handler error")
	assert.Equal(t, newFileName.Name(), tw.FileTree.Subs[1].Name, "handler: filename mismatch error")

	_ = os.RemoveAll(testRoot)

}
