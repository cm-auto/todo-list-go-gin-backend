package models

type ParentAndChildren[T any, U any] struct {
	Parent   T   `json:"parent"`
	Children []U `json:"children"`
}

func NewParentAndChildren[T any, U any](parent T, children []U) ParentAndChildren[T, U] {
	return ParentAndChildren[T, U]{
		Parent:   parent,
		Children: children,
	}
}
