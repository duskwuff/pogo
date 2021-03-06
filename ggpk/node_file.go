package ggpk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type FileNode struct {
	nodeCommon
	name       string
	signature  [32]byte
	headerSize int64
}

func (n *FileNode) Name() string      { return n.name }
func (n *FileNode) Type() string      { return "FILE" }
func (n *FileNode) Offset() int64     { return n.offset }
func (n *FileNode) Length() int64     { return n.length }
func (n *FileNode) Signature() []byte { return n.signature[:] }

func (g *File) initNodeFILE(offset int64, data []byte) (*FileNode, error) {
	var node physFILE
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &node); err != nil {
		return nil, fmt.Errorf("unable to read FILE header at %08x: %w", offset, err)
	}

	if len(data) < int(44+g.sizeofName(node.NameLen)) {
		data = make([]byte, 44+g.sizeofName(node.NameLen))
		_, err := g.file.ReadAt(data, offset)
		if err != nil {
			return nil, fmt.Errorf("unable to read FILE data at %08x: %w", offset, err)
		}
	}

	br := bytes.NewReader(data[44:])

	name, err := g.readStringFrom(br)
	if err != nil {
		return nil, fmt.Errorf("unable to read FILE name at %08x: %w", offset, err)
	}

	return &FileNode{
		nodeCommon: nodeCommon{
			src:    g,
			offset: offset,
			length: int64(node.Length),
		},
		name:       name,
		signature:  node.Signature,
		headerSize: int64(44 + g.sizeofName(node.NameLen)),
	}, nil
}

func (n *FileNode) Size() int64 {
	return n.nodeCommon.length - n.headerSize
}

func (n *FileNode) Reader() ReadSeekerAt {
	return io.NewSectionReader(
		n.src.file,
		n.offset+n.headerSize,
		n.nodeCommon.length-n.headerSize,
	)
}
