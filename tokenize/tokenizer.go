package tokenize

type Tokenizer interface {
	SplitTokens(input string, size int) ([][]int, error)
	Decode(chunk []int) (string, error)
}
