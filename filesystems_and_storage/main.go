// package main implements a simple in-memory filesystem
// Pre-requisites: libfuse3-dev, gcc, add build tags to settings.json
package main

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/winfsp/cgofuse/fuse"
)

func split(path string) []string {
	return strings.Split(path, "/")
}

type node struct {
	stat      fuse.Stat_t
	child     map[string]*node
	attr      map[string][]byte
	data      []byte
	openCount int
}

func newNode(deviceID uint64, inode uint64, mode uint32, userID uint32, groupID uint32) *node {
	tmstmp := fuse.Now()

	n := node{
		fuse.Stat_t{
			Dev:      deviceID,
			Ino:      inode,
			Mode:     mode,
			Uid:      userID,
			Gid:      groupID,
			Atim:     tmstmp,
			Mtim:     tmstmp,
			Ctim:     tmstmp,
			Birthtim: tmstmp,
		},
		nil, nil, nil, 0,
	}

	// extract the file type bits from the rest of the inode mode bits
	// and check if the inode is directory
	if n.stat.Mode&fuse.S_IFMT == fuse.S_IFDIR {
		n.child = map[string]*node{}
	}

	return &n
}

type FileSystem struct {
	fuse.FileSystemBase
	mtx          sync.Mutex
	inode        uint64
	root         *node
	openInodeMap map[uint64]*node
}

func NewFileSystem() *FileSystem {
	fs := FileSystem{}

	defer fs.sync()()

	fs.inode++
	fs.root = newNode(0, fs.inode, fuse.S_IFDIR|00774, 0, 0)
	fs.openInodeMap = map[uint64]*node{}

	return &fs
}

func (fs *FileSystem) sync() func() {
	fs.mtx.Lock()

	return func() {
		fs.mtx.Unlock()
	}
}

func (fs *FileSystem) lookupNode(path string, ancestor *node) (parent *node, name string, node *node) {
	parent = fs.root
	name = ""
	node = fs.root

	for _, p := range split(path) {
		if p != "" {
			if 255 < utf8.RuneCountInString(p) {
				panic(fuse.Error(-fuse.ENAMETOOLONG))
			}

			parent, name = node, p

			if node == nil {
				return
			}

			node = node.child[p]

			if ancestor != nil && node == ancestor {
				name = ""
				return
			}
		}
	}

	return
}

func (fs *FileSystem) makeNode(path string, mode uint32, deviceID uint64, data []byte) int {
	parent, name, node := fs.lookupNode(path, nil)

	if parent == nil {
		return -fuse.ENOENT
	}

	if node != nil {
		return -fuse.EEXIST
	}

	fs.inode++
	uid, gid, _ := fuse.Getcontext()
	node = newNode(deviceID, fs.inode, mode, uid, gid)
	if data != nil {
		node.data = make([]byte, len(data))
		node.stat.Size = int64(len(data))
		copy(node.data, data)
	}

	parent.child[name] = node
	parent.stat.Ctim = node.stat.Ctim
	parent.stat.Mtim = node.stat.Mtim

	return 0
}

func (fs *FileSystem) deleteNode(path string, dir bool) int {
	parent, name, node := fs.lookupNode(path, nil)

	if node == nil {
		return -fuse.ENOENT
	}

	if !dir && fuse.S_IFDIR == fuse.S_IFMT&node.stat.Mode {
		return -fuse.EISDIR
	}

	if dir && fuse.S_IFDIR != fuse.S_IFMT&node.stat.Mode {
		return -fuse.ENOTDIR
	}

	if len(node.child) > 0 {
		return -fuse.ENOTEMPTY
	}

	node.stat.Nlink--
	delete(parent.child, name)
	tmstmp := fuse.Now()
	node.stat.Atim = tmstmp
	parent.stat.Ctim = tmstmp
	parent.stat.Mtim = tmstmp

	return 0
}

func (fs *FileSystem) openNode(path string, dir bool) (int, uint64) {
	_, _, node := fs.lookupNode(path, nil)

	if node == nil {
		return -fuse.ENOENT, ^uint64(0)
	}

	if !dir && fuse.S_IFDIR == fuse.S_IFMT&node.stat.Mode {
		return -fuse.EISDIR, ^uint64(0)
	}

	if dir && fuse.S_IFDIR != fuse.S_IFMT&node.stat.Mode {
		return -fuse.ENOTDIR, ^uint64(0)
	}

	node.openCount++

	if node.openCount == 1 {
		fs.openInodeMap[node.stat.Ino] = node
	}

	return 0, node.stat.Ino
}

func (fs *FileSystem) closeNode(fileHandler uint64) int {
	node := fs.openInodeMap[fileHandler]
	node.openCount--

	if node.openCount == 0 {
		delete(fs.openInodeMap, node.stat.Ino)
	}

	return 0
}

func (fs *FileSystem) getNode(path string, fileHandler uint64) *node {
	if ^uint64(0) == fileHandler {
		_, _, node := fs.lookupNode(path, nil)

		return node
	} else {
		return fs.openInodeMap[fileHandler]
	}
}
