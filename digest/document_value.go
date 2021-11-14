package digest

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/hint"

	"github.com/ProtoconNet/mitum-did/did"
)

var (
	DocumentValueType = hint.Type("mitum-did-document-value")
	DocumentValueHint = hint.NewHint(DocumentValueType, "v0.0.1")
)

type DocumentValue struct {
	doc    did.DocumentData
	height base.Height
}

func NewDocumentValue(
	doc did.DocumentData,
	height base.Height,
) DocumentValue {

	return DocumentValue{
		doc:    doc,
		height: height,
	}
}

func (dv DocumentValue) Hint() hint.Hint {
	return DocumentValueHint
}

func (dv DocumentValue) Document() did.DocumentData {
	return dv.doc
}

func (dv DocumentValue) Height() base.Height {
	return dv.height
}
