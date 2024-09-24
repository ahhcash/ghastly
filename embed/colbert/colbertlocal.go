package colbert

import (
	"github.com/sugarme/transformer/bert"
)

const (
	hfBaseUrl     = "https://huggingface.co/api"
	modelCacheDir = "./static/models/"
)

type ColBERT struct {
	model     *bert.BertModel
	tokenizer *bert.Tokenizer
}

// figure out how the fuck I can run colbert inference locally
