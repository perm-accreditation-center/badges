package ooxml

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	maxFileSize = 50 << 20
	maxXMLSize  = 25 << 20
)

// Node is a minimal OOXML tree node containing visible text.
type Node struct {
	Name     xml.Name
	Children []*Node
	Text     string
}

func (n *Node) ChildrenNamed(name string) []*Node {
	result := make([]*Node, 0)
	for _, child := range n.Children { if child.Name.Local == name { result = append(result, child) } }
	return result
}

func (n *Node) DescendantsNamed(name string) []*Node {
	result := make([]*Node, 0)
	for _, child := range n.Children {
		if child.Name.Local == name { result = append(result, child) }
		result = append(result, child.DescendantsNamed(name)...)
	}
	return result
}

func (n *Node) VisibleText() string { return strings.Join(visibleText(n), " ") }

// Document is the visible content of word/document.xml.
type Document struct { Root *Node }

// Text returns all visible document text.
func (d *Document) Text() string {
	if d == nil || d.Root == nil { return "" }
	return strings.Join(visibleText(d.Root), " ")
}

func visibleText(node *Node) []string {
	parts := make([]string, 0)
	if strings.TrimSpace(node.Text) != "" { parts = append(parts, node.Text) }
	for _, child := range node.Children { parts = append(parts, visibleText(child)... ) }
	return parts
}

// ReadDocument opens the document XML without extracting the ZIP archive.
func ReadDocument(path string) (*Document, error) {
	info, err := os.Stat(path)
	if err != nil { return nil, err }
	if info.Size() > maxFileSize { return nil, fmt.Errorf("DOCX exceeds 50 MiB limit") }
	archive, err := zip.OpenReader(path)
	if err != nil { return nil, fmt.Errorf("open DOCX: %w", err) }
	defer archive.Close()

	var part *zip.File
	for _, file := range archive.File {
		if file.Name == "word/document.xml" { part = file; break }
	}
	if part == nil { return nil, fmt.Errorf("DOCX has no word/document.xml") }
	if part.UncompressedSize64 > maxXMLSize { return nil, fmt.Errorf("document XML exceeds 25 MiB limit") }
	reader, err := part.Open()
	if err != nil { return nil, fmt.Errorf("open document XML: %w", err) }
	defer reader.Close()
	return decode(reader)
}

func decode(reader io.Reader) (*Document, error) {
	decoder := xml.NewDecoder(io.LimitReader(reader, maxXMLSize+1))
	var root *Node
	stack := make([]*Node, 0)
	deletedDepth := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF { break }
		if err != nil { return nil, fmt.Errorf("parse document XML: %w", err) }
		switch value := token.(type) {
		case xml.Directive:
			return nil, fmt.Errorf("XML directives are not allowed")
		case xml.StartElement:
			node := &Node{Name: value.Name}
			if len(stack) == 0 { root = node } else { parent := stack[len(stack)-1]; parent.Children = append(parent.Children, node) }
			stack = append(stack, node)
			if value.Name.Local == "del" { deletedDepth++ }
		case xml.EndElement:
			if value.Name.Local == "del" && deletedDepth > 0 { deletedDepth-- }
			if len(stack) > 0 { stack = stack[:len(stack)-1] }
		case xml.CharData:
			if deletedDepth == 0 && len(stack) > 0 { stack[len(stack)-1].Text += string(value) }
		}
	}
	if root == nil { return nil, fmt.Errorf("document XML is empty") }
	return &Document{Root: root}, nil
}
