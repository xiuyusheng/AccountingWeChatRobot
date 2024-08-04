package gpt

type Identity string

const (
	system         Identity = "system"
	user           Identity = "user"
	assistant      Identity = "assistant"
	openaiInternal Identity = "openai-internal"
)
