package text

// This file defines interfaces commonly used throughout the project for
// handling text-related operations, such as rendering AST, processing file
// content and generating reports.

import "io"

// Writer is an interface that combines the functionality of standard
// I\O interfaces ([io.Writer], [io.ByteWriter], [io.StringWriter], providing
// a unified way to handle various types of data streams.
type Writer interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
}

// Renderer is an interface for rendering text.
//
// There is also a method to check if the content is renderable and can
// perform the rendering without any errors.
type Renderer interface {
	// Render writes the rendered content to the provided [Writer].
	// If the content cannot be properly rendered, it must return false and
	// not write any data in the buffer, otherwise.
	Render(buf Writer) (rendered bool)
}
