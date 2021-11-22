package did

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/util/hint"
)

var (
	CreateDocumentsItemSingleFileType   = hint.Type("mitum-did-create-documents-single-file")
	CreateDocumentsItemSingleFileHint   = hint.NewHint(CreateDocumentsItemSingleFileType, "v0.0.1")
	CreateDocumentsItemSingleFileHinter = BaseCreateDocumentsItem{hint: CreateDocumentsItemSingleFileHint}
)

type CreateDocumentsItemSingleFile struct {
	BaseCreateDocumentsItem
}

func NewCreateDocumentsItemSingleFile(
	sm Summary,
	documentid currency.Big,
	signcode, title string,
	size currency.Big,
	cid currency.CurrencyID,
) CreateDocumentsItemSingleFile {
	return CreateDocumentsItemSingleFile{
		BaseCreateDocumentsItem: NewBaseCreateDocumentsItem(
			CreateDocumentsItemSingleFileHint,
			sm,
			documentid,
			signcode,
			title,
			size,
			cid,
		),
	}
}

func (it CreateDocumentsItemSingleFile) IsValid([]byte) error {
	if err := it.BaseCreateDocumentsItem.IsValid(nil); err != nil {
		return err
	}
	return nil
}
