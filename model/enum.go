package model

type ThreadType uint8

const (
	ThreadTypeUser ThreadType = iota
	ThreadTypeGroup
)
